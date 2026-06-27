package fbhttp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	"os"
	gopath "path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/mholt/archives"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/files"
)

const archiveTextPreviewLimit int64 = 10 * 1024 * 1024

var errArchiveEntryFound = errors.New("archive entry found")

type archiveExtractResponse struct {
	Destination string `json:"destination"`
	Extracted   int    `json:"extracted"`
}

type archiveOpenResult struct {
	extractor archives.Extractor
	stream    io.Reader
	close     func() error
}

func cleanArchiveInnerPath(inner string) (string, error) {
	inner = strings.TrimSpace(strings.ReplaceAll(inner, "\\", "/"))
	inner = strings.TrimPrefix(inner, "/")

	if inner == "" || inner == "." {
		return ".", nil
	}

	cleaned := gopath.Clean(inner)
	if cleaned == "." {
		return ".", nil
	}

	if gopath.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") {
		return "", fmt.Errorf("unsafe archive path: %w", fberrors.ErrInvalidRequestParams)
	}

	return cleaned, nil
}

func archiveAPIDisplayPath(inner string) string {
	if inner == "." || inner == "" {
		return "/"
	}
	return "/" + inner
}

func archiveEntryPath(entry string) (string, error) {
	entry = strings.ReplaceAll(entry, "\\", "/")
	entry = strings.TrimPrefix(entry, "/")
	entry = strings.TrimSuffix(entry, "/")
	return cleanArchiveInnerPath(entry)
}

func stripKnownArchiveExtension(name string) string {
	lower := strings.ToLower(name)
	multi := []string{
		".tar.gz", ".tgz", ".tar.bz2", ".tbz2", ".tar.xz", ".txz",
		".tar.zst", ".tzst", ".tar.lz4", ".tlz4", ".tar.br", ".tar.sz",
	}
	for _, ext := range multi {
		if strings.HasSuffix(lower, ext) {
			return name[:len(name)-len(ext)]
		}
	}
	if ext := filepath.Ext(name); ext != "" {
		return strings.TrimSuffix(name, ext)
	}
	return name + "-extracted"
}

func openArchive(ctx context.Context, archivePath string, d *data) (*archiveOpenResult, error) {
	archivePath = gopath.Clean("/" + archivePath)
	if !d.Check(archivePath) {
		return nil, fberrors.ErrPermissionDenied
	}

	file, err := files.NewFileInfo(&files.FileOptions{
		Fs:         d.user.Fs,
		Path:       archivePath,
		Modify:     false,
		Expand:     true,
		ReadHeader: d.server.TypeDetectionByHeader,
		Checker:    d,
		Content:    false,
	})
	if err != nil {
		return nil, err
	}
	if file.IsDir || file.Type != "archive" {
		return nil, fmt.Errorf("resource is not a supported archive: %w", fberrors.ErrInvalidRequestParams)
	}

	f, err := d.user.Fs.Open(archivePath)
	if err != nil {
		return nil, err
	}

	format, stream, err := archives.Identify(ctx, file.Name, f)
	if err != nil {
		_ = f.Close()
		return nil, err
	}

	extractor, ok := format.(archives.Extractor)
	if !ok {
		_ = f.Close()
		return nil, fmt.Errorf("archive format is not extractable: %w", fberrors.ErrInvalidRequestParams)
	}

	return &archiveOpenResult{
		extractor: extractor,
		stream:    stream,
		close:     f.Close,
	}, nil
}

func detectArchiveResourceType(name string, size int64, opener func() (fs.File, error), readHeader bool) string {
	extension := strings.ToLower(filepath.Ext(name))
	if files.IsArchiveName(name) {
		return "archive"
	}

	mimetype := mime.TypeByExtension(extension)
	var buffer []byte
	if readHeader && mimetype == "" && opener != nil {
		f, err := opener()
		if err == nil {
			defer f.Close()
			buffer = make([]byte, 512)
			n, _ := f.Read(buffer)
			buffer = buffer[:n]
			mimetype = http.DetectContentType(buffer)
		}
	}

	switch {
	case strings.HasPrefix(mimetype, "video"):
		return "video"
	case strings.HasPrefix(mimetype, "audio"):
		return "audio"
	case strings.HasPrefix(mimetype, "image"):
		return "image"
	case strings.HasSuffix(mimetype, "pdf"):
		return "pdf"
	case size <= archiveTextPreviewLimit && isArchiveTextMIME(mimetype, extension):
		return "textImmutable"
	default:
		return "blob"
	}
}

func isArchiveTextMIME(mimetype, extension string) bool {
	if strings.HasPrefix(mimetype, "text") {
		return true
	}
	switch mimetype {
	case "application/json", "application/xml", "application/javascript", "application/x-javascript", "application/x-yaml", "text/yaml":
		return true
	}
	switch extension {
	case ".json", ".yaml", ".yml", ".toml", ".xml", ".js", ".ts", ".tsx", ".jsx", ".vue", ".css", ".scss", ".md", ".markdown", ".go", ".py", ".sh", ".bash", ".zsh", ".txt", ".log", ".conf", ".ini", ".env", ".sql", ".csv":
		return true
	}
	return false
}

func newArchiveFileInfo(archivePath, inner string, info fs.FileInfo, isDir bool, typ string) *files.FileInfo {
	displayPath := archiveAPIDisplayPath(inner)
	name := ""
	if inner != "." {
		name = gopath.Base(inner)
	} else {
		name = gopath.Base(archivePath)
	}
	if info != nil && name == "" {
		name = info.Name()
	}

	size := int64(0)
	modTime := time.Time{}
	mode := os.FileMode(0)
	if info != nil {
		size = info.Size()
		modTime = info.ModTime()
		mode = info.Mode()
	}
	if isDir {
		mode |= os.ModeDir
		typ = "dir"
	}

	return &files.FileInfo{
		Path:             displayPath,
		Name:             name,
		Size:             size,
		Extension:        filepath.Ext(name),
		ModTime:          modTime,
		Mode:             mode,
		IsDir:            isDir,
		Type:             typ,
		Archive:          true,
		ArchivePath:      archivePath,
		ArchiveInnerPath: displayPath,
	}
}

func buildArchiveResource(ctx context.Context, archivePath, inner string, d *data, includeContent bool) (*files.FileInfo, error) {
	opened, err := openArchive(ctx, archivePath, d)
	if err != nil {
		return nil, err
	}
	defer opened.close()

	target, err := cleanArchiveInnerPath(inner)
	if err != nil {
		return nil, err
	}

	children := map[string]*files.FileInfo{}
	var exact *files.FileInfo
	hasChildren := target == "."

	err = opened.extractor.Extract(ctx, opened.stream, func(ctx context.Context, entry archives.FileInfo) error {
		entryPath, cleanErr := archiveEntryPath(entry.NameInArchive)
		if cleanErr != nil || entryPath == "." {
			return nil
		}

		info := entry.FileInfo
		entryIsDir := false
		if info != nil {
			entryIsDir = info.IsDir()
		}
		if strings.HasSuffix(strings.ReplaceAll(entry.NameInArchive, "\\", "/"), "/") {
			entryIsDir = true
		}

		if entryPath == target {
			typ := detectArchiveResourceType(entryPath, 0, entry.Open, false)
			if info != nil {
				typ = detectArchiveResourceType(entryPath, info.Size(), entry.Open, false)
			}
			exact = newArchiveFileInfo(archivePath, entryPath, info, entryIsDir, typ)

			if !entryIsDir && includeContent && exact.Type == "textImmutable" && exact.Size <= archiveTextPreviewLimit && entry.Open != nil {
				f, openErr := entry.Open()
				if openErr != nil {
					return openErr
				}
				defer f.Close()
				data, readErr := io.ReadAll(io.LimitReader(f, archiveTextPreviewLimit+1))
				if readErr != nil {
					return readErr
				}
				if int64(len(data)) <= archiveTextPreviewLimit {
					exact.Content = string(data)
				}
			}
			return nil
		}

		var rest string
		if target == "." {
			rest = entryPath
		} else {
			prefix := target + "/"
			if !strings.HasPrefix(entryPath, prefix) {
				return nil
			}
			hasChildren = true
			rest = strings.TrimPrefix(entryPath, prefix)
		}

		if rest == "" {
			return nil
		}

		parts := strings.Split(rest, "/")
		childInner := parts[0]
		if target != "." {
			childInner = target + "/" + parts[0]
		}

		if len(parts) > 1 {
			if _, ok := children[childInner]; !ok {
				children[childInner] = newArchiveFileInfo(archivePath, childInner, nil, true, "dir")
			}
			return nil
		}

		typ := detectArchiveResourceType(entryPath, 0, entry.Open, false)
		if info != nil {
			typ = detectArchiveResourceType(entryPath, info.Size(), entry.Open, false)
		}
		children[childInner] = newArchiveFileInfo(archivePath, childInner, info, entryIsDir, typ)
		return nil
	})
	if err != nil {
		return nil, err
	}

	if exact != nil && !exact.IsDir {
		return exact, nil
	}

	if exact == nil && !hasChildren {
		return nil, os.ErrNotExist
	}

	root := exact
	if root == nil {
		root = newArchiveFileInfo(archivePath, target, nil, true, "dir")
	}
	root.IsDir = true
	root.Type = "dir"
	root.Listing = &files.Listing{Items: []*files.FileInfo{}}

	keys := make([]string, 0, len(children))
	for key := range children {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		child := children[key]
		child.Archive = true
		child.ArchivePath = archivePath
		child.ArchiveInnerPath = child.Path
		if child.IsDir {
			root.Listing.NumDirs++
		} else {
			root.Listing.NumFiles++
		}
		root.Listing.Items = append(root.Listing.Items, child)
	}

	root.Sorting = d.user.Sorting
	root.ApplySort()
	return root, nil
}

var archiveResourceHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Download {
		return http.StatusAccepted, nil
	}

	archivePath := gopath.Clean("/" + r.URL.Path)
	inner := r.URL.Query().Get("inner")

	resource, err := buildArchiveResource(r.Context(), archivePath, inner, d, true)
	if err != nil {
		return errToStatus(err), err
	}

	return renderJSON(w, r, resource)
})

var archiveRawHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Download {
		return http.StatusAccepted, nil
	}

	archivePath := gopath.Clean("/" + r.URL.Path)
	inner, err := cleanArchiveInnerPath(r.URL.Query().Get("inner"))
	if err != nil {
		return http.StatusBadRequest, err
	}
	if inner == "." {
		return http.StatusBadRequest, fmt.Errorf("archive raw path must be a file: %w", fberrors.ErrInvalidRequestParams)
	}

	opened, err := openArchive(r.Context(), archivePath, d)
	if err != nil {
		return errToStatus(err), err
	}
	defer opened.close()

	err = opened.extractor.Extract(r.Context(), opened.stream, func(ctx context.Context, entry archives.FileInfo) error {
		entryPath, cleanErr := archiveEntryPath(entry.NameInArchive)
		if cleanErr != nil || entryPath != inner {
			return nil
		}

		if entry.FileInfo != nil && entry.FileInfo.IsDir() {
			return fberrors.ErrIsDirectory
		}
		if entry.Open == nil {
			return os.ErrNotExist
		}

		f, openErr := entry.Open()
		if openErr != nil {
			return openErr
		}
		defer f.Close()

		name := gopath.Base(entryPath)
		if r.URL.Query().Get("inline") == "true" {
			w.Header().Set("Content-Disposition", "inline; filename*=utf-8''"+url.PathEscape(name))
		} else {
			w.Header().Set("Content-Disposition", "attachment; filename*=utf-8''"+url.PathEscape(name))
		}
		if contentType := mime.TypeByExtension(filepath.Ext(name)); contentType != "" {
			w.Header().Set("Content-Type", contentType)
		} else {
			w.Header().Set("Content-Type", "application/octet-stream")
		}
		w.Header().Add("Content-Security-Policy", `script-src 'none';`)
		w.Header().Set("Cache-Control", "private")

		if entry.FileInfo != nil && entry.FileInfo.Size() >= 0 {
			w.Header().Set("Content-Length", fmt.Sprintf("%d", entry.FileInfo.Size()))
		}

		if _, copyErr := io.Copy(w, f); copyErr != nil {
			return copyErr
		}
		return errArchiveEntryFound
	})
	if errors.Is(err, errArchiveEntryFound) {
		return 0, nil
	}
	if err != nil {
		return errToStatus(err), err
	}

	return http.StatusNotFound, os.ErrNotExist
})

var archiveExtractHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Create {
		return http.StatusForbidden, nil
	}
	if r.URL.Query().Get("override") == "true" && !d.user.Perm.Modify {
		return http.StatusForbidden, nil
	}

	archivePath := gopath.Clean("/" + r.URL.Path)
	inner, err := cleanArchiveInnerPath(r.URL.Query().Get("inner"))
	if err != nil {
		return http.StatusBadRequest, err
	}

	destination := r.URL.Query().Get("destination")
	if destination == "" {
		baseName := stripKnownArchiveExtension(gopath.Base(archivePath))
		destination = gopath.Join(gopath.Dir(archivePath), baseName)
	}
	destination, err = url.QueryUnescape(strings.ReplaceAll(destination, "+", "%2B"))
	if err != nil {
		return http.StatusBadRequest, err
	}
	destination = gopath.Clean("/" + destination)
	if destination == "/" || !d.Check(destination) {
		return http.StatusForbidden, nil
	}

	if r.URL.Query().Get("rename") == "true" {
		destination = addVersionSuffix(destination, d.user.Fs)
	} else if r.URL.Query().Get("override") != "true" {
		if _, statErr := d.user.Fs.Stat(destination); statErr == nil {
			return http.StatusConflict, os.ErrExist
		}
	}

	response := archiveExtractResponse{Destination: destination}
	err = d.RunHook(func() error {
		extracted, extractErr := extractArchiveTo(r.Context(), d, archivePath, inner, destination, r.URL.Query().Get("override") == "true")
		response.Extracted = extracted
		return extractErr
	}, "upload", archivePath, destination, d.user)
	if err != nil {
		return errToStatus(err), err
	}

	return renderJSON(w, r, response)
})

func extractArchiveTo(ctx context.Context, d *data, archivePath, inner, destination string, override bool) (int, error) {
	opened, err := openArchive(ctx, archivePath, d)
	if err != nil {
		return 0, err
	}
	defer opened.close()

	extracted := 0
	found := inner == "."

	err = opened.extractor.Extract(ctx, opened.stream, func(ctx context.Context, entry archives.FileInfo) error {
		entryPath, cleanErr := archiveEntryPath(entry.NameInArchive)
		if cleanErr != nil || entryPath == "." {
			return nil
		}

		rel := ""
		if inner == "." {
			rel = entryPath
		} else if entryPath == inner {
			found = true
			rel = gopath.Base(entryPath)
		} else {
			prefix := inner + "/"
			if !strings.HasPrefix(entryPath, prefix) {
				return nil
			}
			found = true
			rel = strings.TrimPrefix(entryPath, prefix)
		}
		if rel == "" || rel == "." {
			return nil
		}

		rel, cleanErr = cleanArchiveInnerPath(rel)
		if cleanErr != nil || rel == "." {
			return nil
		}

		target := gopath.Clean(gopath.Join(destination, rel))
		if target == destination || !strings.HasPrefix(target, destination+"/") {
			return fmt.Errorf("unsafe extraction target: %w", fberrors.ErrInvalidRequestParams)
		}
		if !d.Check(target) {
			return fberrors.ErrPermissionDenied
		}

		mode := os.FileMode(0)
		entryIsDir := strings.HasSuffix(strings.ReplaceAll(entry.NameInArchive, "\\", "/"), "/")
		if entry.FileInfo != nil {
			mode = entry.FileInfo.Mode()
			entryIsDir = entryIsDir || entry.FileInfo.IsDir()
		}
		if mode&os.ModeSymlink != 0 || mode&os.ModeIrregular != 0 || mode&os.ModeNamedPipe != 0 || mode&os.ModeDevice != 0 {
			return nil
		}

		if entryIsDir {
			if err := d.user.Fs.MkdirAll(target, d.settings.DirMode); err != nil {
				return err
			}
			return nil
		}

		if !override {
			if _, err := d.user.Fs.Stat(target); err == nil {
				return os.ErrExist
			}
		}
		if err := d.user.Fs.MkdirAll(gopath.Dir(target), d.settings.DirMode); err != nil {
			return err
		}
		if entry.Open == nil {
			return nil
		}
		f, err := entry.Open()
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := writeFile(d.user.Fs, target, f, d.settings.FileMode, d.settings.DirMode); err != nil {
			return err
		}
		extracted++
		return nil
	})
	if err != nil {
		return extracted, err
	}
	if !found {
		return extracted, os.ErrNotExist
	}
	return extracted, nil
}
