package fbhttp

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	gopath "path"
	"path/filepath"
	"strings"

	libErrors "github.com/filebrowser/filebrowser/v2/errors"
	imgErrors "github.com/filebrowser/filebrowser/v2/img"
)

// slashClean canonicalizes a virtual path to the absolute, "/"-separated,
// lexically cleaned form that both the rule checker and the scoped filesystem
// expect.
//
// Paths reach us either verbatim from a request or built by the OS — afero.Walk
// and filepath.Join use "\" on Windows. There the filesystem resolves "\" as a
// separator while the rule checker treated it as an ordinary character, so
// "/allow\..\Secret.txt" evaded a rule for "/Secret.txt" and still opened it.
// Normalizing with the host's own separator makes both layers agree on which
// file a path names, and is a no-op on POSIX, where "\" is a legal character in
// a filename and must stay one.
func slashClean(name string) string {
	return cleanSeparators(name, string(filepath.Separator))
}

// canonicalizeRequestPath rewrites the request path to its canonical virtual
// form, so that the filesystem operation, the hook, the stored share record and
// the thumbnail cache key all use the same string the rule check approved.
//
// A trailing separator is preserved: handlers distinguish "/dir/" from "/dir"
// to decide whether to create a directory, and gopath.Clean would drop it.
func canonicalizeRequestPath(r *http.Request) {
	p := slashClean(r.URL.Path)
	trailing := strings.HasSuffix(r.URL.Path, "/") || strings.HasSuffix(r.URL.Path, string(filepath.Separator))
	if p != "/" && trailing {
		p += "/"
	}
	r.URL.Path = p
}

// cleanSeparators is slashClean with the host separator passed in explicitly,
// so the Windows behaviour can be exercised on any platform.
func cleanSeparators(name, sep string) string {
	if sep != "/" {
		name = strings.ReplaceAll(name, sep, "/")
	}
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return gopath.Clean(name)
}

func renderJSON(w http.ResponseWriter, _ *http.Request, data interface{}) (int, error) {
	marsh, err := json.Marshal(data)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(marsh); err != nil {
		return http.StatusInternalServerError, err
	}

	return 0, nil
}

func errToStatus(err error) int {
	switch {
	case err == nil:
		return http.StatusOK
	case os.IsPermission(err):
		return http.StatusForbidden
	case os.IsNotExist(err), errors.Is(err, libErrors.ErrNotExist):
		return http.StatusNotFound
	case os.IsExist(err), errors.Is(err, libErrors.ErrExist):
		return http.StatusConflict
	case errors.Is(err, libErrors.ErrPermissionDenied):
		return http.StatusForbidden
	case errors.Is(err, libErrors.ErrInvalidRequestParams):
		return http.StatusBadRequest
	case errors.Is(err, libErrors.ErrRootUserDeletion):
		return http.StatusForbidden
	case errors.Is(err, imgErrors.ErrImageTooLarge):
		return http.StatusRequestEntityTooLarge
	default:
		return http.StatusInternalServerError
	}
}

// This is an adaptation if http.StripPrefix in which we don't
// return 404 if the page doesn't have the needed prefix.
func stripPrefix(prefix string, h http.Handler) http.Handler {
	if prefix == "" || prefix == "/" {
		return h
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, prefix)
		rp := strings.TrimPrefix(r.URL.RawPath, prefix)

		// If the path is exactly the prefix (no trailing slash), redirect to
		// the prefix with a trailing slash so the router receives "/" instead
		// of "", which would otherwise cause a redirect to the site root.
		if p == "" {
			http.Redirect(w, r, prefix+"/", http.StatusMovedPermanently)
			return
		}

		r2 := new(http.Request)
		*r2 = *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.Path = p
		r2.URL.RawPath = rp
		h.ServeHTTP(w, r2)
	})
}
