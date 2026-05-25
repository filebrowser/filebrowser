package fbhttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	gopath "path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/settings"
)

const defaultConvertXTimeout = 2 * time.Minute
const convertXJobPollInterval = 1500 * time.Millisecond

var errConvertXIntegrationAPINotFound = errors.New("ConvertX integration API not found")

type convertXTargetsResponse struct {
	Success bool                `json:"success"`
	From    string              `json:"from"`
	Targets map[string][]string `json:"targets"`
	Message string              `json:"message,omitempty"`
}

type convertXConvertRequest struct {
	Path        string `json:"path"`
	ConvertTo   string `json:"convertTo"`
	Converter   string `json:"converter,omitempty"`
	Destination string `json:"destination,omitempty"`
	Rename      *bool  `json:"rename,omitempty"`
	Overwrite   bool   `json:"overwrite,omitempty"`
}

type convertXConvertResponse struct {
	JobID       string `json:"jobId,omitempty"`
	Status      string `json:"status"`
	Message     string `json:"message,omitempty"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	ConvertTo   string `json:"convertTo"`
	Converter   string `json:"converter,omitempty"`
	Done        bool   `json:"done"`
}

type convertXConvertPlan struct {
	Source      string
	Destination string
	From        string
	ConvertTo   string
	Converter   string
	Overwrite   bool
}

type convertXSubmitResponse struct {
	Success        bool   `json:"success"`
	JobID          string `json:"jobId"`
	Status         string `json:"status"`
	InputFile      string `json:"inputFile"`
	From           string `json:"from"`
	ConvertTo      string `json:"convertTo"`
	Converter      string `json:"converter"`
	ExpectedOutput string `json:"expectedOutput"`
	StatusURL      string `json:"statusUrl"`
	Message        string `json:"message,omitempty"`
}

type convertXJobFile struct {
	Input       string `json:"input"`
	Output      string `json:"output"`
	Status      string `json:"status"`
	DownloadURL string `json:"downloadUrl"`
}

type convertXJobResponse struct {
	Success bool              `json:"success"`
	JobID   string            `json:"jobId"`
	Status  string            `json:"status"`
	Total   int               `json:"total"`
	Done    int               `json:"done"`
	Files   []convertXJobFile `json:"files"`
	Message string            `json:"message,omitempty"`
}

type convertXConversionJobStore struct {
	mu   sync.Mutex
	jobs map[string]convertXConversionJobEntry
}

type convertXConversionJobEntry struct {
	Response  convertXConvertResponse
	UpdatedAt time.Time
}

var convertXConversionJobs = &convertXConversionJobStore{jobs: map[string]convertXConversionJobEntry{}}

func normalizeConvertXURL(raw string) (string, error) {
	cleaned := strings.TrimRight(strings.TrimSpace(raw), "/")
	if cleaned == "" {
		return "", errors.New("no ConvertX URL is configured")
	}

	parsed, err := url.Parse(cleaned)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("ConvertX URL must be a valid absolute URL, for example http://convertx.lan:3000")
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", errors.New("ConvertX URL must use http or https")
	}

	return cleaned, nil
}

func convertXTimeout(raw string) time.Duration {
	if strings.TrimSpace(raw) == "" {
		return defaultConvertXTimeout
	}

	duration, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil || duration <= 0 {
		return defaultConvertXTimeout
	}

	return duration
}

func testConvertXConnection(ctx context.Context, cfg settings.ConvertX) error {
	base, err := normalizeConvertXURL(cfg.URL)
	if err != nil {
		return err
	}

	timeout := convertXTimeout(cfg.Timeout)
	client := &http.Client{Timeout: timeout}

	// The Convert to... action requires the ConvertX File Browser integration API,
	// not only the browser UI. A reachable root page is not enough.
	return probeConvertXURL(ctx, client, base+"/api/health", cfg.APIKey, true)
}

func probeConvertXURL(ctx context.Context, client *http.Client, targetURL string, apiKey string, integrationAPI bool) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return err
	}

	addConvertXAuth(req, apiKey)

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to reach ConvertX: %w", err)
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 1024))

	if res.StatusCode >= 200 && res.StatusCode < 400 {
		return nil
	}
	if integrationAPI && res.StatusCode == http.StatusNotFound {
		return fmt.Errorf("%w at %s", errConvertXIntegrationAPINotFound, targetURL)
	}

	return fmt.Errorf("ConvertX returned HTTP %d for %s", res.StatusCode, targetURL)
}

var convertXTargetsHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Download {
		return http.StatusForbidden, nil
	}
	if !d.settings.ConvertX.Enabled {
		return http.StatusBadRequest, errors.New("ConvertX integration is disabled")
	}

	from := strings.TrimSpace(r.URL.Query().Get("from"))
	if from == "" {
		return http.StatusBadRequest, fberrors.ErrInvalidRequestParams
	}

	result, err := fetchConvertXTargets(r.Context(), d.settings.ConvertX, from)
	if err != nil {
		return http.StatusBadGateway, err
	}

	return renderJSON(w, r, result)
})

var convertXConvertHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Download || !d.user.Perm.Create {
		return http.StatusForbidden, nil
	}
	if !d.settings.ConvertX.Enabled {
		return http.StatusBadRequest, errors.New("ConvertX integration is disabled")
	}

	request := convertXConvertRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return http.StatusBadRequest, err
	}

	plan, status, err := prepareConvertXConversionPlan(request, d)
	if err != nil {
		return status, err
	}

	if strings.EqualFold(r.URL.Query().Get("sync"), "true") {
		response, err := runConvertXConversionSync(r.Context(), d, plan)
		if err != nil {
			return errToStatus(err), err
		}
		return renderJSON(w, r, response)
	}

	jobID := newFileActionJobID()
	response := convertXConvertResponse{
		JobID:       jobID,
		Status:      "running",
		Message:     fmt.Sprintf("Conversion started: %s → %s", plan.Source, plan.Destination),
		Source:      plan.Source,
		Destination: plan.Destination,
		ConvertTo:   plan.ConvertTo,
		Converter:   plan.Converter,
	}
	convertXConversionJobs.set(jobID, response)

	go runConvertXConversionJob(jobID, d, plan)

	w.WriteHeader(http.StatusAccepted)
	return renderJSON(w, r, response)
})

var convertXConversionJobStatusHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Download || !d.user.Perm.Create {
		return http.StatusForbidden, nil
	}

	jobID := strings.Trim(gopath.Clean("/"+r.URL.Path), "/")
	if jobID == "" {
		return http.StatusBadRequest, fberrors.ErrInvalidRequestParams
	}

	response, ok := convertXConversionJobs.get(jobID)
	if !ok {
		return http.StatusNotFound, os.ErrNotExist
	}

	return renderJSON(w, r, response)
})

func prepareConvertXConversionPlan(request convertXConvertRequest, d *data) (*convertXConvertPlan, int, error) {
	source := cleanFileBrowserPath(request.Path)
	if source == "/" || !d.Check(source) {
		return nil, http.StatusForbidden, fberrors.ErrPermissionDenied
	}

	info, err := d.user.Fs.Stat(source)
	if err != nil {
		return nil, errToStatus(err), err
	}
	if info.IsDir() {
		return nil, http.StatusBadRequest, errors.New("ConvertX can convert files, not folders")
	}

	convertTo := sanitizeConvertXFormat(request.ConvertTo)
	if convertTo == "" {
		return nil, http.StatusBadRequest, errors.New("missing target conversion format")
	}

	converter := strings.TrimSpace(request.Converter)
	if converter != "" && !safeConvertXToken(converter) {
		return nil, http.StatusBadRequest, errors.New("invalid ConvertX converter name")
	}

	from := strings.TrimPrefix(strings.ToLower(filepath.Ext(source)), ".")
	if from == "" {
		return nil, http.StatusBadRequest, errors.New("the selected file has no extension; ConvertX cannot detect the source format")
	}

	destination := strings.TrimSpace(request.Destination)
	if destination == "" {
		destination = defaultConvertXDestination(source, convertTo)
	} else {
		destination = cleanFileBrowserPath(destination)
	}

	if destination == "/" || !d.Check(destination) {
		return nil, http.StatusForbidden, fberrors.ErrPermissionDenied
	}
	if request.Overwrite && !d.user.Perm.Modify {
		return nil, http.StatusForbidden, fberrors.ErrPermissionDenied
	}
	if strings.HasSuffix(destination, "/") {
		return nil, http.StatusBadRequest, errors.New("conversion destination must be a file path")
	}
	if !strings.HasSuffix(strings.ToLower(destination), "."+convertTo) {
		destination += "." + convertTo
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

	return &convertXConvertPlan{
		Source:      source,
		Destination: destination,
		From:        from,
		ConvertTo:   convertTo,
		Converter:   converter,
		Overwrite:   request.Overwrite,
	}, http.StatusOK, nil
}

func runConvertXConversionSync(ctx context.Context, d *data, plan *convertXConvertPlan) (convertXConvertResponse, error) {
	response := convertXConvertResponse{
		Status:      "running",
		Source:      plan.Source,
		Destination: plan.Destination,
		ConvertTo:   plan.ConvertTo,
		Converter:   plan.Converter,
	}

	err := d.RunHook(func() error {
		return convertXConvertAndSave(ctx, d, plan)
	}, "upload", plan.Destination, "", d.user)
	if err != nil {
		_ = d.user.Fs.Remove(plan.Destination)
		return response, err
	}

	response.Status = "done"
	response.Done = true
	response.Message = fmt.Sprintf("Converted file saved to %s", plan.Destination)
	return response, nil
}

func runConvertXConversionJob(jobID string, d *data, plan *convertXConvertPlan) {
	response := convertXConvertResponse{
		JobID:       jobID,
		Status:      "running",
		Source:      plan.Source,
		Destination: plan.Destination,
		ConvertTo:   plan.ConvertTo,
		Converter:   plan.Converter,
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = d.user.Fs.Remove(plan.Destination)
			response.Status = "error"
			response.Done = true
			response.Message = fmt.Sprintf("Conversion failed for %s: %v", plan.Source, recovered)
			convertXConversionJobs.set(jobID, response)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), convertXTimeout(d.settings.ConvertX.Timeout))
	defer cancel()

	completed, err := runConvertXConversionSync(ctx, d, plan)
	completed.JobID = jobID
	if err != nil {
		completed.Status = "error"
		completed.Done = true
		completed.Message = err.Error()
		convertXConversionJobs.set(jobID, completed)
		return
	}

	convertXConversionJobs.set(jobID, completed)
}

func convertXConvertAndSave(ctx context.Context, d *data, plan *convertXConvertPlan) error {
	file, err := d.user.Fs.Open(plan.Source)
	if err != nil {
		return err
	}
	defer file.Close()

	submit, err := submitConvertXConversion(ctx, d.settings.ConvertX, filepath.Base(plan.Source), plan.From, plan.ConvertTo, plan.Converter, file)
	if err != nil {
		return err
	}

	download, err := waitForConvertXOutput(ctx, d.settings.ConvertX, submit)
	if err != nil {
		return err
	}
	defer download.Close()

	_, err = writeFile(d.user.Fs, plan.Destination, download, d.settings.FileMode, d.settings.DirMode)
	if err != nil {
		return err
	}

	return nil
}

func fetchConvertXTargets(ctx context.Context, cfg settings.ConvertX, from string) (*convertXTargetsResponse, error) {
	base, err := normalizeConvertXURL(cfg.URL)
	if err != nil {
		return nil, err
	}

	from = sanitizeConvertXFormat(from)
	if from == "" {
		return nil, errors.New("invalid source format")
	}

	targetURL := base + "/api/conversions?from=" + url.QueryEscape(from)
	client := &http.Client{Timeout: convertXTimeout(cfg.Timeout)}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	addConvertXAuth(req, cfg.APIKey)

	var result convertXTargetsResponse
	if err := doConvertXJSON(client, req, &result); err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, convertXMessageError(result.Message)
	}
	if result.Targets == nil {
		result.Targets = map[string][]string{}
	}

	return &result, nil
}

func submitConvertXConversion(ctx context.Context, cfg settings.ConvertX, fileName, from, convertTo, converter string, file io.Reader) (*convertXSubmitResponse, error) {
	base, err := normalizeConvertXURL(cfg.URL)
	if err != nil {
		return nil, err
	}

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		part, err := writer.CreateFormFile("file", fileName)
		if err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		if _, err := io.Copy(part, file); err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		if err := writer.WriteField("from", from); err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		if err := writer.WriteField("convert_to", convertTo); err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		if strings.TrimSpace(converter) != "" {
			if err := writer.WriteField("converter", strings.TrimSpace(converter)); err != nil {
				_ = pw.CloseWithError(err)
				return
			}
		}
		if err := writer.Close(); err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		_ = pw.Close()
	}()

	targetURL := base + "/api/convert"
	client := &http.Client{Timeout: convertXTimeout(cfg.Timeout)}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, pr)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	addConvertXAuth(req, cfg.APIKey)

	var result convertXSubmitResponse
	if err := doConvertXJSON(client, req, &result); err != nil {
		return nil, err
	}
	if !result.Success || result.JobID == "" {
		return nil, convertXMessageError(result.Message)
	}

	return &result, nil
}

func waitForConvertXOutput(ctx context.Context, cfg settings.ConvertX, submit *convertXSubmitResponse) (io.ReadCloser, error) {
	base, err := normalizeConvertXURL(cfg.URL)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: convertXTimeout(cfg.Timeout)}
	ticker := time.NewTicker(convertXJobPollInterval)
	defer ticker.Stop()

	for {
		job, err := fetchConvertXJob(ctx, client, cfg.APIKey, base, submit.JobID)
		if err != nil {
			return nil, err
		}

		if strings.EqualFold(job.Status, "failed") || strings.EqualFold(job.Status, "error") {
			return nil, convertXMessageError(job.Message)
		}

		if strings.EqualFold(job.Status, "completed") || strings.EqualFold(job.Status, "done") {
			downloadURL := firstConvertXDownloadURL(job, submit.ExpectedOutput)
			if downloadURL == "" {
				return nil, errors.New("ConvertX completed the job but did not expose a converted output file")
			}
			return downloadConvertXOutput(ctx, client, cfg.APIKey, base, downloadURL)
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("ConvertX conversion timed out: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

func fetchConvertXJob(ctx context.Context, client *http.Client, apiKey, base, jobID string) (*convertXJobResponse, error) {
	targetURL := base + "/api/jobs/" + url.PathEscape(jobID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, err
	}
	addConvertXAuth(req, apiKey)

	var result convertXJobResponse
	if err := doConvertXJSON(client, req, &result); err != nil {
		return nil, err
	}
	if !result.Success {
		return nil, convertXMessageError(result.Message)
	}
	return &result, nil
}

func downloadConvertXOutput(ctx context.Context, client *http.Client, apiKey, base, downloadURL string) (io.ReadCloser, error) {
	if strings.HasPrefix(downloadURL, "/") {
		downloadURL = base + downloadURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return nil, err
	}
	addConvertXAuth(req, apiKey)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		defer res.Body.Close()
		return nil, convertXHTTPError(res)
	}

	return res.Body, nil
}

func doConvertXJSON(client *http.Client, req *http.Request, target interface{}) error {
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to reach ConvertX: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return convertXHTTPError(res)
	}

	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("failed to parse ConvertX response: %w", err)
	}

	return nil
}

func convertXHTTPError(res *http.Response) error {
	limited, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
	message := strings.TrimSpace(string(limited))
	if message == "" {
		message = http.StatusText(res.StatusCode)
	}
	return fmt.Errorf("ConvertX returned HTTP %d: %s", res.StatusCode, message)
}

func addConvertXAuth(req *http.Request, apiKey string) {
	if strings.TrimSpace(apiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))
	}
}

func convertXMessageError(message string) error {
	message = strings.TrimSpace(message)
	if message == "" {
		message = "ConvertX conversion failed"
	}
	return errors.New(message)
}

func firstConvertXDownloadURL(job *convertXJobResponse, expectedOutput string) string {
	for _, file := range job.Files {
		if strings.EqualFold(file.Status, "done") && file.Output == expectedOutput && file.DownloadURL != "" {
			return file.DownloadURL
		}
	}
	for _, file := range job.Files {
		if strings.EqualFold(file.Status, "done") && file.DownloadURL != "" {
			return file.DownloadURL
		}
	}
	return ""
}

func cleanFileBrowserPath(raw string) string {
	cleaned := strings.TrimSpace(raw)
	cleaned = strings.TrimPrefix(cleaned, "/files")
	return gopath.Clean("/" + cleaned)
}

func sanitizeConvertXFormat(raw string) string {
	value := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(raw, ".")))
	if !safeConvertXToken(value) {
		return ""
	}
	return value
}

func safeConvertXToken(value string) bool {
	if value == "" || strings.Contains(value, "..") {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}
		return false
	}
	return true
}

func defaultConvertXDestination(source, convertTo string) string {
	dir, name := gopath.Split(source)
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	if base == "" {
		base = strings.TrimSuffix(name, ".")
	}
	return gopath.Join(dir, base+"."+convertTo)
}

func (s *convertXConversionJobStore) set(jobID string, response convertXConvertResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked()
	s.jobs[jobID] = convertXConversionJobEntry{Response: response, UpdatedAt: time.Now()}
}

func (s *convertXConversionJobStore) get(jobID string) (convertXConvertResponse, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked()
	entry, ok := s.jobs[jobID]
	return entry.Response, ok
}

func (s *convertXConversionJobStore) cleanupLocked() {
	cutoff := time.Now().Add(-fileActionJobTTL)
	for id, entry := range s.jobs {
		if entry.UpdatedAt.Before(cutoff) {
			delete(s.jobs, id)
		}
	}
}
