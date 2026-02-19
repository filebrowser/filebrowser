package fbhttp

import (
	"log"
	"net/http"
	"sync"

	"github.com/filebrowser/filebrowser/v2/users"
	"golang.org/x/net/webdav"
)

var (
	webDavLocks   = make(map[uint]webdav.LockSystem)
	webDavLocksMu sync.Mutex
)

func getLockSystem(uid uint) webdav.LockSystem {
	webDavLocksMu.Lock()
	defer webDavLocksMu.Unlock()
	if _, ok := webDavLocks[uid]; !ok {
		webDavLocks[uid] = webdav.NewMemLS()
	}
	return webDavLocks[uid]
}

func webDavHandler(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.settings.EnableWebDAV {
		return http.StatusForbidden, nil
	}

	// WebDAV uses Basic Auth
	username, password, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="File Browser"`)
		return http.StatusUnauthorized, nil
	}

	user, err := d.store.Users.Get(d.server.Root, username)
	if err != nil {
		w.Header().Set("WWW-Authenticate", `Basic realm="File Browser"`)
		return http.StatusUnauthorized, nil
	}

	if !users.CheckPwd(password, user.Password) {
		w.Header().Set("WWW-Authenticate", `Basic realm="File Browser"`)
		return http.StatusUnauthorized, nil
	}

	// TODO: Verify permissions?
	// For now, we assume if you can login, you can WebDAV.
	// But we might want to respect ReadOnly, etc.
	// Check methods?

	switch r.Method {
	case "PUT", "DELETE", "MKCOL", "COPY", "MOVE":
		if !user.Perm.Modify { // modify includes delete? renaming?
			return http.StatusForbidden, nil
		}
	}

	handler := &webdav.Handler{
		Prefix:     "/webdav",
		FileSystem: &webDavFS{fs: user.Fs},
		LockSystem: getLockSystem(user.ID),
		Logger: func(r *http.Request, err error) {
			if err != nil {
				log.Printf("WebDAV error: %s [%s]: %v", r.RemoteAddr, r.Method, err)
			}
		},
	}

	handler.ServeHTTP(w, r)
	return 0, nil
}
