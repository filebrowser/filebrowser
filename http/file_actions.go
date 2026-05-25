package fbhttp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	gopath "path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mholt/archives"
	"github.com/spf13/afero"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/fileutils"
)

type clamAVPathScanResponse struct {
	JobID    string `json:"jobId,omitempty"`
	Status   string `json:"status"`
	Message  string `json:"message"`
	Path     string `json:"path"`
	Scanned  int    `json:"scanned"`
	Infected bool   `json:"infected"`
	Threat   string `json:"threat,omitempty"`
	Done     bool   `json:"done"`
}

type archiveCreateRequest struct {
	Items       []string `json:"items"`
	Destination string   `json:"destination,omitempty"`
	Format      string   `json:"format,omitempty"`
	Rename      *bool    `json:"rename,omitempty"`
	Overwrite   bool     `json:"overwrite,omitempty"`
}

type archiveCreateResponse struct {
	JobID       string `json:"jobId,omitempty"`
	Status      string `json:"status,omitempty"`
	Message     string `json:"message,omitempty"`
	Destination string `json:"destination"`
	Archived    int    `json:"archived"`
	Format      string `json:"format"`
	Done        bool   `json:"done,omitempty"`
}

type archiveCreatePlan struct {
	Paths       []string
	Destination string
	Extension   string
	Archiver    archives.Archival
	Overwrite   bool
}

const fileActionJobTTL = 2 * time.Hour

type clamAVScanJobStore struct {
	mu   sync.Mutex
	jobs map[string]clamAVScanJobEntry
}

type clamAVScanJobEntry struct {
	Response  clamAVPathScanResponse
	UpdatedAt time.Time
}

type archiveCreateJobStore struct {
	mu   sync.Mutex
	jobs map[string]archiveCreateJobEntry
}

type archiveCreateJobEntry struct {
	Response  archiveCreateResponse
	UpdatedAt time.Time
}

var (
	clamAVScanJobs    = &clamAVScanJobStore{jobs: map[string]clamAVScanJobEntry{}}
	archiveCreateJobs = &archiveCreateJobStore{jobs: map[string]archiveCreateJobEntry{}}
)

var clamAVPathScanHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Download {
		return http.StatusForbidden, nil
	}

	requestedPath := gopath.Clean("/" + r.URL.Path)
	if requestedPath == "/" || !d.Check(requestedPath) {
		return http.StatusForbidden, nil
	}

	// Manual folder scans can take longer than a reverse proxy timeout. The
	// default path is asynchronous: return a job immediately and let the
	// frontend poll for completion. A synchronous mode is retained for direct
	// API callers that explicitly request it.
	if strings.EqualFold(r.URL.Query().Get("sync"), "true") {
		return runClamAVPathScanSync(w, r, d, requestedPath)
	}

	jobID := newFileActionJobID()
	response := clamAVPathScanResponse{
		JobID:   jobID,
		Status:  "running",
		Message: fmt.Sprintf("Security scan started for %q.", requestedPath),
		Path:    requestedPath,
	}
	clamAVScanJobs.set(jobID, response)

	go runClamAVPathScanJob(jobID, d, requestedPath)

	w.WriteHeader(http.StatusAccepted)
	return renderJSON(w, r, response)
})

var clamAVScanJobStatusHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Download {
		return http.StatusForbidden, nil
	}

	jobID := strings.Trim(gopath.Clean("/"+r.URL.Path), "/")
	if jobID == "" {
		return http.StatusBadRequest, fberrors.ErrInvalidRequestParams
	}

	response, ok := clamAVScanJobs.get(jobID)
	if !ok {
		return http.StatusNotFound, os.ErrNotExist
	}

	return renderJSON(w, r, response)
})

var archiveCreateHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Download || !d.user.Perm.Create {
		return http.StatusForbidden, nil
	}

	request := archiveCreateRequest{Format: "zip"}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return http.StatusBadRequest, err
	}

	plan, status, err := prepareArchiveCreatePlan(request, d)
	if err != nil {
		return status, err
	}

	if strings.EqualFold(r.URL.Query().Get("sync"), "true") {
		response, err := runArchiveCreateSync(r.Context(), d, plan)
		if err != nil {
			return errToStatus(err), err
		}
		return renderJSON(w, r, response)
	}

	jobID := newFileActionJobID()
	response := archiveCreateResponse{
		JobID:       jobID,
		Status:      "running",
		Message:     fmt.Sprintf("Archive creation started: %s", plan.Destination),
		Destination: plan.Destination,
		Format:      strings.TrimPrefix(plan.Extension, "."),
	}
	archiveCreateJobs.set(jobID, response)

	go runArchiveCreateJob(jobID, d, plan)

	w.WriteHeader(http.StatusAccepted)
	return renderJSON(w, r, response)
})

var archiveCreateJobStatusHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Download || !d.user.Perm.Create {
		return http.StatusForbidden, nil
	}

	jobID := strings.Trim(gopath.Clean("/"+r.URL.Path), "/")
	if jobID == "" {
		return http.StatusBadRequest, fberrors.ErrInvalidRequestParams
	}

	response, ok := archiveCreateJobs.get(jobID)
	if !ok {
		return http.StatusNotFound, os.ErrNotExist
	}

	return renderJSON(w, r, response)
})

func runClamAVPathScanSync(w http.ResponseWriter, r *http.Request, d *data, requestedPath string) (int, error) {
	scanned, threat, err := scanExistingPathWithClamAV(r.Context(), d, requestedPath)
	if threat != nil {
		response := clamAVPathScanResponse{
			Status:   "infected",
			Message:  manualClamAVThreatMessage(threat),
			Path:     requestedPath,
			Scanned:  scanned,
			Infected: true,
			Threat:   cleanClamAVMessage(threat.Threat, threat.FileName),
			Done:     true,
		}
		return renderJSON(w, r, response)
	}
	if err != nil {
		return clamAVHTTPStatus(err), err
	}

	return renderJSON(w, r, cleanClamAVPathScanResponse(requestedPath, scanned))
}

func runClamAVPathScanJob(jobID string, d *data, requestedPath string) {
	response := clamAVPathScanResponse{JobID: jobID, Status: "running", Path: requestedPath}
	defer func() {
		if recovered := recover(); recovered != nil {
			response.Status = "error"
			response.Done = true
			response.Message = fmt.Sprintf("Security scan failed for %q: %v", requestedPath, recovered)
			clamAVScanJobs.set(jobID, response)
		}
	}()

	scanned, threat, err := scanExistingPathWithClamAV(context.Background(), d, requestedPath)
	response.Scanned = scanned
	response.Done = true

	if threat != nil {
		response.Status = "infected"
		response.Message = manualClamAVThreatMessage(threat)
		response.Infected = true
		response.Threat = cleanClamAVMessage(threat.Threat, threat.FileName)
		clamAVScanJobs.set(jobID, response)
		return
	}

	if err != nil {
		response.Status = "error"
		response.Message = err.Error()
		clamAVScanJobs.set(jobID, response)
		return
	}

	response = cleanClamAVPathScanResponse(requestedPath, scanned)
	response.JobID = jobID
	clamAVScanJobs.set(jobID, response)
}

func cleanClamAVPathScanResponse(requestedPath string, scanned int) clamAVPathScanResponse {
	message := fmt.Sprintf("Security scan completed. No malware was found in %q.", requestedPath)
	if scanned != 1 {
		message = fmt.Sprintf("Security scan completed. No malware was found in %d files under %q.", scanned, requestedPath)
	}

	return clamAVPathScanResponse{
		Status:  "clean",
		Message: message,
		Path:    requestedPath,
		Scanned: scanned,
		Done:    true,
	}
}

func runArchiveCreateSync(ctx context.Context, d *data, plan *archiveCreatePlan) (archiveCreateResponse, error) {
	response := archiveCreateResponse{Destination: plan.Destination, Format: strings.TrimPrefix(plan.Extension, "."), Done: true}
	err := d.RunHook(func() error {
		archived, archiveErr := createArchiveFromPaths(ctx, d, plan.Paths, plan.Destination, plan.Archiver, plan.Overwrite)
		response.Archived = archived
		return archiveErr
	}, "upload", plan.Destination, "", d.user)
	if err != nil {
		_ = d.user.Fs.Remove(plan.Destination)
		return response, err
	}

	response.Status = "done"
	response.Message = fmt.Sprintf("Archive created: %s", plan.Destination)
	return response, nil
}

func runArchiveCreateJob(jobID string, d *data, plan *archiveCreatePlan) {
	response := archiveCreateResponse{
		JobID:       jobID,
		Status:      "running",
		Destination: plan.Destination,
		Format:      strings.TrimPrefix(plan.Extension, "."),
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = d.user.Fs.Remove(plan.Destination)
			response.Status = "error"
			response.Done = true
			response.Message = fmt.Sprintf("Archive creation failed for %s: %v", plan.Destination, recovered)
			archiveCreateJobs.set(jobID, response)
		}
	}()

	completed, err := runArchiveCreateSync(context.Background(), d, plan)
	completed.JobID = jobID
	if err != nil {
		completed.Status = "error"
		completed.Done = true
		completed.Message = err.Error()
		archiveCreateJobs.set(jobID, completed)
		return
	}

	archiveCreateJobs.set(jobID, completed)
}

func prepareArchiveCreatePlan(request archiveCreateRequest, d *data) (*archiveCreatePlan, int, error) {
	paths, err := cleanArchiveCreateItems(request.Items, d)
	if err != nil {
		return nil, errToStatus(err), err
	}

	extension, archiver, err := archiveCreateFormat(request.Format)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	destination := strings.TrimSpace(request.Destination)
	if destination == "" {
		destination = defaultArchiveCreateDestination(paths, extension)
	} else {
		destination = gopath.Clean("/" + destination)
	}

	if destination == "/" || !d.Check(destination) {
		return nil, http.StatusForbidden, fberrors.ErrPermissionDenied
	}
	if request.Overwrite && !d.user.Perm.Modify {
		return nil, http.StatusForbidden, fberrors.ErrPermissionDenied
	}
	if !strings.HasSuffix(strings.ToLower(destination), strings.ToLower(extension)) {
		destination += extension
	}

	rename := true
	if request.Rename != nil {
		rename = *request.Rename
	}
	if rename {
		destination = addVersionSuffix(destination, d.user.Fs)
	} else if !request.Overwrite {
		if _, statErr := d.user.Fs.Stat(destination); statErr == nil {
			return nil, http.StatusConflict, os.ErrExist
		}
	}

	return &archiveCreatePlan{
		Paths:       paths,
		Destination: destination,
		Extension:   extension,
		Archiver:    archiver,
		Overwrite:   request.Overwrite,
	}, http.StatusOK, nil
}

func newFileActionJobID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

func (s *clamAVScanJobStore) set(jobID string, response clamAVPathScanResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked()
	s.jobs[jobID] = clamAVScanJobEntry{Response: response, UpdatedAt: time.Now()}
}

func (s *clamAVScanJobStore) get(jobID string) (clamAVPathScanResponse, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked()
	entry, ok := s.jobs[jobID]
	return entry.Response, ok
}

func (s *clamAVScanJobStore) cleanupLocked() {
	cutoff := time.Now().Add(-fileActionJobTTL)
	for id, entry := range s.jobs {
		if entry.UpdatedAt.Before(cutoff) {
			delete(s.jobs, id)
		}
	}
}

func (s *archiveCreateJobStore) set(jobID string, response archiveCreateResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked()
	s.jobs[jobID] = archiveCreateJobEntry{Response: response, UpdatedAt: time.Now()}
}

func (s *archiveCreateJobStore) get(jobID string) (archiveCreateResponse, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked()
	entry, ok := s.jobs[jobID]
	return entry.Response, ok
}

func (s *archiveCreateJobStore) cleanupLocked() {
	cutoff := time.Now().Add(-fileActionJobTTL)
	for id, entry := range s.jobs {
		if entry.UpdatedAt.Before(cutoff) {
			delete(s.jobs, id)
		}
	}
}

func scanExistingPathWithClamAV(ctx context.Context, d *data, scanPath string) (int, *clamAVThreatError, error) {
	cfg := d.settings.ClamAV
	if strings.TrimSpace(cfg.URL) == "" {
		return 0, nil, &clamAVServiceError{Message: "no ClamAV URL is configured"}
	}

	info, err := d.user.Fs.Stat(scanPath)
	if err != nil {
		return 0, nil, err
	}

	scanned := 0
	scanFile := func(filePath string) error {
		if !d.Check(filePath) {
			return nil
		}

		file, openErr := d.user.Fs.Open(filePath)
		if openErr != nil {
			return openErr
		}
		defer file.Close()

		scanned++
		return scanWithClamAV(ctx, cfg, strings.TrimPrefix(filePath, "/"), file)
	}

	if !info.IsDir() {
		err = scanFile(scanPath)
		if err != nil {
			var threat *clamAVThreatError
			if errors.As(err, &threat) {
				return scanned, threat, nil
			}
			return scanned, nil, err
		}
		return scanned, nil, nil
	}

	err = afero.Walk(d.user.Fs, scanPath, func(filePath string, info fs.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if filePath == scanPath {
			return nil
		}
		if !d.Check(filePath) {
			if info != nil && info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info == nil || info.IsDir() {
			return nil
		}
		return scanFile(filePath)
	})
	if err != nil {
		var threat *clamAVThreatError
		if errors.As(err, &threat) {
			return scanned, threat, nil
		}
		return scanned, nil, err
	}

	return scanned, nil, nil
}

func manualClamAVThreatMessage(threat *clamAVThreatError) string {
	fileName := strings.TrimSpace(threat.FileName)
	if fileName == "" {
		fileName = "selected file"
	}

	threatName := cleanClamAVMessage(threat.Threat, fileName)
	details := cleanClamAVMessage(threat.Details, fileName)

	switch {
	case threatName != "":
		return fmt.Sprintf("Security scan found malware in %q. Threat: %s. No files were modified or removed.", fileName, threatName)
	case details != "":
		return fmt.Sprintf("Security scan found malware in %q. Scanner details: %s. No files were modified or removed.", fileName, details)
	default:
		return fmt.Sprintf("Security scan found malware in %q. No files were modified or removed.", fileName)
	}
}

func cleanArchiveCreateItems(items []string, d *data) ([]string, error) {
	if len(items) == 0 {
		return nil, fberrors.ErrInvalidRequestParams
	}

	paths := make([]string, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		item = gopath.Clean("/" + strings.TrimSpace(item))
		if item == "/" || !d.Check(item) {
			return nil, fberrors.ErrPermissionDenied
		}
		if _, err := d.user.Fs.Stat(item); err != nil {
			return nil, err
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		paths = append(paths, item)
	}

	if len(paths) == 0 {
		return nil, fberrors.ErrInvalidRequestParams
	}
	return paths, nil
}

func archiveCreateFormat(format string) (string, archives.Archival, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", "zip":
		return ".zip", archives.Zip{}, nil
	case "tar":
		return ".tar", archives.Tar{}, nil
	case "targz", "tar.gz", "tgz":
		return ".tar.gz", archives.CompressedArchive{Compression: archives.Gz{}, Archival: archives.Tar{}}, nil
	case "tarbz2", "tar.bz2", "tbz2":
		return ".tar.bz2", archives.CompressedArchive{Compression: archives.Bz2{}, Archival: archives.Tar{}}, nil
	case "tarxz", "tar.xz", "txz":
		return ".tar.xz", archives.CompressedArchive{Compression: archives.Xz{}, Archival: archives.Tar{}}, nil
	case "tarlz4", "tar.lz4":
		return ".tar.lz4", archives.CompressedArchive{Compression: archives.Lz4{}, Archival: archives.Tar{}}, nil
	case "tarsz", "tar.sz":
		return ".tar.sz", archives.CompressedArchive{Compression: archives.Sz{}, Archival: archives.Tar{}}, nil
	case "tarbr", "tar.br":
		return ".tar.br", archives.CompressedArchive{Compression: archives.Brotli{}, Archival: archives.Tar{}}, nil
	case "tarzst", "tar.zst":
		return ".tar.zst", archives.CompressedArchive{Compression: archives.Zstd{}, Archival: archives.Tar{}}, nil
	default:
		return "", nil, fmt.Errorf("archive format is not supported: %w", fberrors.ErrInvalidRequestParams)
	}
}

func defaultArchiveCreateDestination(paths []string, extension string) string {
	parent := gopath.Dir(paths[0])
	name := "archive"
	if len(paths) == 1 {
		name = strings.TrimSpace(gopath.Base(paths[0]))
		if name == "" || name == "." || name == "/" {
			name = "archive"
		}
	}

	return gopath.Join(parent, name+extension)
}

func createArchiveFromPaths(ctx context.Context, d *data, paths []string, destination string, archiver archives.Archival, overwrite bool) (int, error) {
	commonDir := archiveCreateCommonDir(paths)
	allFiles := make([]archives.FileInfo, 0)

	for _, path := range paths {
		archiveFiles, err := getFiles(d, path, commonDir)
		if err != nil {
			return 0, err
		}
		allFiles = append(allFiles, archiveFiles...)
	}

	if len(allFiles) == 0 {
		return 0, os.ErrNotExist
	}

	if err := d.user.Fs.MkdirAll(gopath.Dir(destination), d.settings.DirMode); err != nil {
		return 0, err
	}
	if !overwrite {
		if _, err := d.user.Fs.Stat(destination); err == nil {
			return 0, os.ErrExist
		}
	}

	out, err := d.user.Fs.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, d.settings.FileMode)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	if err := archiver.Archive(ctx, out, allFiles); err != nil {
		return 0, err
	}

	return len(allFiles), nil
}

func archiveCreateCommonDir(paths []string) string {
	if len(paths) == 1 {
		return gopath.Dir(paths[0])
	}
	commonDir := fileutils.CommonPrefix('/', paths...)
	if commonDir == "" {
		return "/"
	}
	return commonDir
}
