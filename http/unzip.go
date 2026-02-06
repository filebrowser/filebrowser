package fbhttp

import (
	"archive/zip"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/spf13/afero"
)

func unzipHandler() handleFunc {
	return withUser(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
		src := r.URL.Path
		dst := r.URL.Query().Get("destination")
		dst = filepath.Clean(dst)
		override := r.URL.Query().Get("override") == "true"

		// Check permissions and source/destination access permissions
		if !d.server.UnzipEnabled || !d.user.Perm.Create || !d.Check(src) || !d.Check(dst) {
			return http.StatusForbidden, nil
		}

		// Get zip file data
		file, err := files.NewFileInfo(&files.FileOptions{
			Fs:         d.user.Fs,
			Path:       src,
			Modify:     d.user.Perm.Modify,
			Expand:     false,
			ReadHeader: d.server.TypeDetectionByHeader,
			Checker:    d,
		})

		// Check if the zip file exist
		if err != nil {
			if errors.Is(err, afero.ErrFileNotFound) {
				return http.StatusNotFound, err
			}
			return errToStatus(err), err
		}

		// Check the zip file size
		if file.Size > d.server.MaxZipFileSize {
			return http.StatusBadRequest, fberrors.ErrZipFileIsTooLarge
		}

		// Open zip file
		reader, err := zip.OpenReader(file.RealPath())
		if err != nil {
			return http.StatusInternalServerError, err
		}
		defer reader.Close()

		// Check total file entries
		if len(reader.File) > d.server.MaxZipFileEntries {
			return http.StatusBadRequest, fberrors.ErrZipFileIsTooLarge
		}

		// Check total uncompressed size
		var totalUncompressedSize uint64
		for _, f := range reader.File {
			totalUncompressedSize += f.UncompressedSize64
			if totalUncompressedSize > d.server.MaxTotalUncompressedSize {
				return http.StatusBadRequest, fberrors.ErrUncompressSizeIsTooLarge
			}
		}

		for _, f := range reader.File {
			// Check uncompressed file rate
			if f.UncompressedSize64 == 0 {
				if f.CompressedSize64 > 0 {
					return http.StatusBadRequest, fberrors.ErrInvalidZipEntry
				}
			} else {
				ratio := float64(f.CompressedSize64) / float64(f.UncompressedSize64)
				if ratio < d.server.MaxUncompressedSizeRate {
					return http.StatusBadRequest, fberrors.ErrCompressionRateIsTooLarge
				}
			}

			// Prevent "Zip Slip"
			cleanName := filepath.Clean(f.Name)
			if strings.HasPrefix(cleanName, "/") || strings.HasPrefix(cleanName, "../") || strings.Contains(cleanName, "/../") {
				return http.StatusBadRequest, fberrors.ErrInvalidZipFilePath
			}
			outPath := filepath.Join(dst, cleanName)

			// Check user's permissions to create the file
			if !d.Check(outPath) {
				return http.StatusForbidden, fberrors.ErrInvalidZipFilePath
			}

			// Create directories
			if f.FileInfo().IsDir() {
				if err := d.user.Fs.MkdirAll(outPath, 0755); err != nil {
					return http.StatusInternalServerError, err
				}
				continue
			}

			// Create parents directories
			if err := d.user.Fs.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
				return http.StatusInternalServerError, err
			}

			// Skip if override is not allowed and the file exists
			if !override {
				if exists, _ := afero.Exists(d.user.Fs, outPath); exists {
					continue
				}
			}

			// Check single file uncompress size
			if f.UncompressedSize64 > d.server.MaxUncompressedFileSize {
				return http.StatusInternalServerError, fberrors.ErrUncompressSizeIsTooLarge
			}

			// Create the archive
			rc, err := f.Open()
			if err != nil {
				return http.StatusInternalServerError, err
			}

			outFile, err := d.user.Fs.Create(outPath)
			if err != nil {
				rc.Close()
				return http.StatusInternalServerError, err
			}

			limited := io.LimitReader(rc, int64(d.server.MaxUncompressedFileSize))
			n, err := io.Copy(outFile, limited)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			// Verify that the file did not "lie" about its size.
			if n > int64(f.UncompressedSize64) {
				outFile.Close()
				rc.Close()
				return http.StatusBadRequest, fberrors.ErrInvalidZipEntry
			}

			outFile.Close()
			rc.Close()
		}

		return http.StatusOK, nil
	})
}
