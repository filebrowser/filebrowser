package fbhttp

import (
	"net/http"
	"strconv"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/users"
)

// withQuotaCheck wraps a handler function with quota validation middleware.
// It checks if the user's quota would be exceeded by the operation before allowing it.
func withQuotaCheck(fn handleFunc) handleFunc {
	return func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		// Skip quota check if user has no quota limit or quota is not enforced
		if d.user.QuotaLimit == 0 || !d.user.EnforceQuota {
			return fn(w, r, d)
		}

		// Get the file size from the request
		fileSize, err := getFileSize(r)
		if err != nil {
			// If we can't determine file size, allow the operation
			// (it will be checked during actual write)
			return fn(w, r, d)
		}

		// Check if the operation would exceed quota
		exceeded, currentUsage, err := users.CheckQuotaExceeded(d.user, fileSize)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		if exceeded {
			// Return 413 Payload Too Large with quota information
			w.Header().Set("X-Quota-Limit", users.FormatQuotaDisplay(d.user.QuotaLimit))
			w.Header().Set("X-Quota-Used", users.FormatQuotaDisplay(currentUsage))
			return http.StatusRequestEntityTooLarge, fberrors.ErrQuotaExceeded
		}

		// Quota check passed, proceed with the operation
		return fn(w, r, d)
	}
}

// getFileSize extracts the file size from the HTTP request.
// It checks the Upload-Length header (for TUS uploads) first, then Content-Length.
func getFileSize(r *http.Request) (int64, error) {
	// Try to get size from Upload-Length header (TUS protocol)
	if uploadLength := r.Header.Get("Upload-Length"); uploadLength != "" {
		size, err := strconv.ParseInt(uploadLength, 10, 64)
		if err == nil && size > 0 {
			return size, nil
		}
	}

	// Try to get size from Content-Length header
	if r.ContentLength > 0 {
		return r.ContentLength, nil
	}

	// If neither header is available, return 0
	// This might happen with chunked transfer encoding
	return 0, nil
}
