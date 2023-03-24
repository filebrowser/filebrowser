package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mholt/archiver"
	"github.com/shirou/gopsutil/disk"
	"github.com/spf13/afero"

	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/fileutils"
)

var resourceGetHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	file, err := files.NewFileInfo(files.FileOptions{
		Fs:         d.user.Fs,
		Path:       r.URL.Path,
		Modify:     d.user.Perm.Modify,
		Expand:     true,
		ReadHeader: d.server.TypeDetectionByHeader,
		Checker:    d,
		Content:    true,
	})

	// if the path does not exist and its the trash dir - create it
	if os.IsNotExist(err) && d.user.TrashDir != "" {
		if d.user.FullPath(r.URL.Path) == d.user.FullPath(d.user.TrashDir) {
			err = d.user.Fs.MkdirAll(d.user.TrashDir, 0775) //nolint:gomnd
			if err != nil {
				return errToStatus(err), err
			}

			file, err = files.NewFileInfo(files.FileOptions{
				Fs:         d.user.Fs,
				Path:       r.URL.Path,
				Modify:     d.user.Perm.Modify,
				Expand:     true,
				ReadHeader: d.server.TypeDetectionByHeader,
				Checker:    d,
				Content:    true,
			})
		}
	}

	if err != nil {
		return errToStatus(err), err
	}

	if file.IsSymlink && symlinkOutOfScope(d, file) {
		return errToStatus(errors.ErrNotExist), errors.ErrNotExist
	}

	if r.URL.Query().Get("disk_usage") == "true" {
		du, inodes, err := fileutils.DiskUsage(file.Fs, file.Path)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		file.DiskUsage = du
		file.Inodes = inodes
		file.Content = ""
		return renderJSON(w, r, file)
	}

	if file.IsDir {
		file.Listing.Sorting = d.user.Sorting
		file.Listing.ApplySort()
		file.Listing.FilterItems(func(fi *files.FileInfo) bool {
			// remove files that should be hidden
			_, exists := d.server.HiddenFiles[fi.Name]
			if exists {
				return false
			}

			// remove symlinks that link outside base path
			if fi.IsSymlink {
				link := fi.Link
				isAbs := filepath.IsAbs(link)

				if !isAbs {
					link = filepath.Join(d.user.FullPath(file.Path), link)
				}
				link = filepath.Clean(link)

				if !strings.HasPrefix(link, d.server.Root) {
					return false
				}

				if isAbs {
					fi.Link = strings.TrimPrefix(link, d.server.Root)
				}
			}

			return true
		})
		return renderJSON(w, r, file)
	}

	if checksum := r.URL.Query().Get("checksum"); checksum != "" {
		err := file.Checksum(checksum)
		if err == errors.ErrInvalidOption {
			return http.StatusBadRequest, nil
		} else if err != nil {
			return http.StatusInternalServerError, err
		}

		// do not waste bandwidth if we just want the checksum
		file.Content = ""
	}

	return renderJSON(w, r, file)
})

func resourceDeleteHandler(fileCache FileCache) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if r.URL.Path == "/" || !d.user.Perm.Delete {
			return http.StatusForbidden, nil
		}

		file, err := files.NewFileInfo(files.FileOptions{
			Fs:         d.user.Fs,
			Path:       r.URL.Path,
			Modify:     d.user.Perm.Modify,
			Expand:     false,
			ReadHeader: d.server.TypeDetectionByHeader,
			Checker:    d,
		})
		if err != nil {
			return errToStatus(err), err
		}

		// delete thumbnails
		err = delThumbs(r.Context(), fileCache, file)
		if err != nil {
			return errToStatus(err), err
		}

		skipTrash := r.URL.Query().Get("skip_trash") == "true"

		if d.user.TrashDir == "" || skipTrash {
			err = d.RunHook(func() error {
				return d.user.Fs.RemoveAll(r.URL.Path)
			}, "delete", r.URL.Path, "", d.user)
		} else {
			src := r.URL.Path
			dst := d.user.TrashDir

			if !d.Check(src) || !d.Check(dst) {
				return http.StatusForbidden, nil
			}

			src = path.Clean("/" + src)
			dst = path.Clean("/" + dst)

			err = d.user.Fs.MkdirAll(dst, 0775) //nolint:gomnd
			if err != nil {
				return errToStatus(err), err
			}

			dst = path.Join(dst, file.Name)
			err = fileutils.MoveFile(d.user.Fs, src, dst)
		}

		if err != nil {
			return errToStatus(err), err
		}

		return http.StatusOK, nil
	})
}

func resourcePostHandler(fileCache FileCache) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Create || !d.Check(r.URL.Path) {
			return http.StatusForbidden, nil
		}

		// Directories creation on POST.
		if strings.HasSuffix(r.URL.Path, "/") {
			err := d.user.Fs.MkdirAll(r.URL.Path, 0775) //nolint:gomnd
			return errToStatus(err), err
		}

		// Archive creation on POST.
		if strings.HasSuffix(r.URL.Path, "/archive") {
			if !d.user.Perm.Create {
				return http.StatusForbidden, nil
			}

			return archiveHandler(r, d)
		}

		file, err := files.NewFileInfo(files.FileOptions{
			Fs:         d.user.Fs,
			Path:       r.URL.Path,
			Modify:     d.user.Perm.Modify,
			Expand:     false,
			ReadHeader: d.server.TypeDetectionByHeader,
			Checker:    d,
		})
		if err == nil {
			if r.URL.Query().Get("override") != "true" {
				return http.StatusConflict, nil
			}

			// Permission for overwriting the file
			if !d.user.Perm.Modify {
				return http.StatusForbidden, nil
			}

			err = delThumbs(r.Context(), fileCache, file)
			if err != nil {
				return errToStatus(err), err
			}
		}

		err = d.RunHook(func() error {
			info, writeErr := writeFile(d.user.Fs, r.URL.Path, r.Body)
			if writeErr != nil {
				return writeErr
			}

			etag := fmt.Sprintf(`"%x%x"`, info.ModTime().UnixNano(), info.Size())
			w.Header().Set("ETag", etag)
			return nil
		}, "upload", r.URL.Path, "", d.user)

		if err != nil {
			_ = d.user.Fs.RemoveAll(r.URL.Path)
		}

		return errToStatus(err), err
	})
}

var resourcePutHandler = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.user.Perm.Modify || !d.Check(r.URL.Path) {
		return http.StatusForbidden, nil
	}

	// Only allow PUT for files.
	if strings.HasSuffix(r.URL.Path, "/") {
		return http.StatusMethodNotAllowed, nil
	}

	exists, err := afero.Exists(d.user.Fs, r.URL.Path)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if !exists {
		return http.StatusNotFound, nil
	}

	err = d.RunHook(func() error {
		info, writeErr := writeFile(d.user.Fs, r.URL.Path, r.Body)
		if writeErr != nil {
			return writeErr
		}

		etag := fmt.Sprintf(`"%x%x"`, info.ModTime().UnixNano(), info.Size())
		w.Header().Set("ETag", etag)
		return nil
	}, "save", r.URL.Path, "", d.user)

	return errToStatus(err), err
})

func checkSrcDstAccess(src, dst string, d *data) error {
	if !d.Check(src) || !d.Check(dst) {
		return errors.ErrPermissionDenied
	}

	if dst == "/" || src == "/" {
		return errors.ErrPermissionDenied
	}

	if err := checkParent(src, dst); err != nil {
		return errors.ErrInvalidRequestParams
	}

	return nil
}

func resourcePatchHandler(fileCache FileCache) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		src := r.URL.Path
		dst := r.URL.Query().Get("destination")
		action := r.URL.Query().Get("action")

		if action == "chmod" {
			err := chmodActionHandler(r, d)
			return errToStatus(err), err
		}

		dst, err := url.QueryUnescape(dst)
		if err != nil {
			return errToStatus(err), err
		}

		err = checkSrcDstAccess(src, dst, d)
		if err != nil {
			return errToStatus(err), err
		}

		override := r.URL.Query().Get("override") == "true"
		rename := r.URL.Query().Get("rename") == "true"
		unarchive := action == "unarchive"
		if !override && !rename && !unarchive {
			if _, err = d.user.Fs.Stat(dst); err == nil {
				return http.StatusConflict, nil
			}
		}

		if rename {
			dst = addVersionSuffix(dst, d.user.Fs)
		}

		// Permission for overwriting the file
		if override && !d.user.Perm.Modify {
			return http.StatusForbidden, nil
		}

		err = d.RunHook(func() error {
			if unarchive {
				return unarchiveAction(src, dst, d, override)
			}
			return patchAction(r.Context(), action, src, dst, d, fileCache)
		}, action, src, dst, d.user)

		return errToStatus(err), err
	})
}

func checkParent(src, dst string) error {
	rel, err := filepath.Rel(src, dst)
	if err != nil {
		return err
	}

	rel = filepath.ToSlash(rel)
	if !strings.HasPrefix(rel, "../") && rel != ".." && rel != "." {
		return errors.ErrSourceIsParent
	}

	return nil
}

// Checks if path contains symlink to out-of-scope targets.
// Returns error ErrNotExist if it does.
func symlinkOutOfScope(d *data, file *files.FileInfo) bool {
	var err error

	link := ""
	if lsf, ok := d.user.Fs.(afero.LinkReader); ok {
		if link, err = lsf.ReadlinkIfPossible(file.Path); err != nil {
			return false
		}
	}

	if !filepath.IsAbs(link) {
		link = filepath.Join(d.user.FullPath(file.Path), link)
	}
	link = filepath.Clean(link)

	return !strings.HasPrefix(link, d.server.Root)
}

func addVersionSuffix(source string, fs afero.Fs) string {
	counter := 1
	dir, name := path.Split(source)
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)

	for {
		if _, err := fs.Stat(source); err != nil {
			break
		}
		renamed := fmt.Sprintf("%s(%d)%s", base, counter, ext)
		source = path.Join(dir, renamed)
		counter++
	}

	return source
}

func writeFile(fs afero.Fs, dst string, in io.Reader) (os.FileInfo, error) {
	dir, _ := path.Split(dst)
	err := fs.MkdirAll(dir, 0775) //nolint:gomnd
	if err != nil {
		return nil, err
	}

	file, err := fs.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0775) //nolint:gomnd
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, err = io.Copy(file, in)
	if err != nil {
		return nil, err
	}

	// Gets the info about the file.
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	return info, nil
}

func delThumbs(ctx context.Context, fileCache FileCache, file *files.FileInfo) error {
	for _, previewSizeName := range PreviewSizeNames() {
		size, _ := ParsePreviewSize(previewSizeName)
		if err := fileCache.Delete(ctx, previewCacheKey(file, size)); err != nil {
			return err
		}
	}

	return nil
}

func patchAction(ctx context.Context, action, src, dst string, d *data, fileCache FileCache) error {
	switch action {
	// TODO: use enum
	case "copy":
		if !d.user.Perm.Create {
			return errors.ErrPermissionDenied
		}

		return fileutils.Copy(d.user.Fs, src, dst, d.server.Root)
	case "rename":
		if !d.user.Perm.Rename {
			return errors.ErrPermissionDenied
		}
		src = path.Clean("/" + src)
		dst = path.Clean("/" + dst)

		file, err := files.NewFileInfo(files.FileOptions{
			Fs:         d.user.Fs,
			Path:       src,
			Modify:     d.user.Perm.Modify,
			Expand:     false,
			ReadHeader: false,
			Checker:    d,
		})
		if err != nil {
			return err
		}

		// delete thumbnails
		err = delThumbs(ctx, fileCache, file)
		if err != nil {
			return err
		}

		return fileutils.MoveFile(d.user.Fs, src, dst)
	default:
		return fmt.Errorf("unsupported action %s: %w", action, errors.ErrInvalidRequestParams)
	}
}

type DiskUsageResponse struct {
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
}

//lint:ignore U1000 unused in this fork
var diskUsage = withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	file, err := files.NewFileInfo(files.FileOptions{
		Fs:         d.user.Fs,
		Path:       r.URL.Path,
		Modify:     d.user.Perm.Modify,
		Expand:     false,
		ReadHeader: false,
		Checker:    d,
		Content:    false,
	})
	if err != nil {
		return errToStatus(err), err
	}
	fPath := file.RealPath()
	if !file.IsDir {
		return renderJSON(w, r, &DiskUsageResponse{
			Total: 0,
			Used:  0,
		})
	}

	usage, err := disk.UsageWithContext(r.Context(), fPath)
	if err != nil {
		return errToStatus(err), err
	}
	return renderJSON(w, r, &DiskUsageResponse{
		Total: usage.Total,
		Used:  usage.Used,
	})
})

func unarchiveAction(src, dst string, d *data, overwrite bool) error {
	if !d.user.Perm.Create {
		return errors.ErrPermissionDenied
	}

	src = d.user.FullPath(path.Clean("/" + src))
	dst = d.user.FullPath(path.Clean("/" + dst))

	arch, err := archiver.ByExtension(src)
	if err != nil {
		return err
	}

	switch a := arch.(type) {
	case *archiver.Rar:
		a.OverwriteExisting = overwrite
	case *archiver.Tar:
		a.OverwriteExisting = overwrite
	case *archiver.TarBz2:
		a.OverwriteExisting = overwrite
	case *archiver.TarGz:
		a.OverwriteExisting = overwrite
	case *archiver.TarLz4:
		a.OverwriteExisting = overwrite
	case *archiver.TarSz:
		a.OverwriteExisting = overwrite
	case *archiver.TarXz:
		a.OverwriteExisting = overwrite
	case *archiver.Zip:
		a.OverwriteExisting = overwrite
	}

	unarchiver, ok := arch.(archiver.Unarchiver)
	if ok {
		if err := unarchiver.Unarchive(src, dst); err != nil {
			return errors.ErrInvalidRequestParams
		}

		return nil
	}

	decompressor, ok := arch.(archiver.Decompressor)
	if ok {
		return fileutils.Decompress(src, dst, decompressor)
	}

	return errors.ErrInvalidRequestParams
}

func archiveHandler(r *http.Request, d *data) (int, error) {
	dir := strings.TrimSuffix(r.URL.Path, "/archive")

	destDir, err := files.NewFileInfo(files.FileOptions{
		Fs:         d.user.Fs,
		Path:       dir,
		Modify:     d.user.Perm.Modify,
		Expand:     false,
		ReadHeader: false,
		Checker:    d,
	})
	if err != nil {
		return errToStatus(err), err
	}

	filenames, err := parseQueryFiles(r, destDir, d.user)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	archFile, err := parseQueryFilename(r, destDir)
	if err != nil {
		return http.StatusUnprocessableEntity, errors.NewHTTPError(err, "validation.emptyName")
	}

	extension, ar, err := parseArchiver(r.URL.Query().Get("algo"))
	if err != nil {
		return http.StatusInternalServerError, err
	}

	archFile += extension

	_, err = d.user.Fs.Stat(archFile)
	if err == nil {
		return http.StatusUnprocessableEntity, errors.NewHTTPError(err, "resource.alreadyExists")
	}

	dir, _ = path.Split(archFile)
	err = d.user.Fs.MkdirAll(dir, 0775) //nolint:gomnd
	if err != nil {
		return errToStatus(err), err
	}

	for i, path := range filenames {
		_, err = d.user.Fs.Stat(path)
		if err != nil {
			return errToStatus(err), err
		}
		filenames[i] = d.user.FullPath(path)
	}

	dst := d.user.FullPath(archFile)
	err = ar.Archive(filenames, dst)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return errToStatus(err), err
}

func parseQueryFilename(r *http.Request, f *files.FileInfo) (string, error) {
	name := r.URL.Query().Get("name")
	name, err := url.QueryUnescape(strings.Replace(name, "+", "%2B", -1))
	if err != nil {
		return "", err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("empty name provided")
	}
	return filepath.Join(f.Path, slashClean(name)), nil
}

func parseArchiver(algo string) (string, archiver.Archiver, error) {
	switch algo {
	case "zip", "true", "":
		return ".zip", archiver.NewZip(), nil
	case "tar":
		return ".tar", archiver.NewTar(), nil
	case "targz":
		return ".tar.gz", archiver.NewTarGz(), nil
	case "tarbz2":
		return ".tar.bz2", archiver.NewTarBz2(), nil
	case "tarxz":
		return ".tar.xz", archiver.NewTarXz(), nil
	case "tarlz4":
		return ".tar.lz4", archiver.NewTarLz4(), nil
	case "tarsz":
		return ".tar.sz", archiver.NewTarSz(), nil
	default:
		return "", nil, fmt.Errorf("format not implemented")
	}
}

func chmodActionHandler(r *http.Request, d *data) error {
	target := r.URL.Path
	perms := r.URL.Query().Get("permissions")
	recursive := r.URL.Query().Get("recursive") == "true"
	recursionType := r.URL.Query().Get("type")

	if !d.user.Perm.Modify {
		return errors.ErrPermissionDenied
	}

	if !d.Check(target) || target == "/" {
		return errors.ErrPermissionDenied
	}

	mode, err := strconv.ParseUint(perms, 10, 32) //nolint:gomnd
	if err != nil {
		return errors.ErrInvalidRequestParams
	}

	info, err := d.user.Fs.Stat(target)
	if err != nil {
		return err
	}

	if recursive && info.IsDir() {
		var recFilter func(i os.FileInfo) bool

		switch recursionType {
		case "directories":
			recFilter = func(i os.FileInfo) bool {
				return i.IsDir()
			}
		case "files":
			recFilter = func(i os.FileInfo) bool {
				return !i.IsDir()
			}
		default:
			recFilter = func(i os.FileInfo) bool {
				return true
			}
		}

		return afero.Walk(d.user.Fs, target, func(name string, info os.FileInfo, err error) error {
			if err == nil {
				if recFilter(info) {
					err = d.user.Fs.Chmod(name, os.FileMode(mode))
				}
			}
			return err
		})
	}

	return d.user.Fs.Chmod(target, os.FileMode(mode))
}
