package fbhttp

import (
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/asdine/storm/v3"
	"github.com/spf13/afero"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage/bolt"
	"github.com/filebrowser/filebrowser/v2/users"
)

func TestWebDavHandler(t *testing.T) {
	const password = "password"
	hashedPassword, _ := users.HashPwd(password)

	testCases := map[string]struct {
		enableWebDAV       bool
		method             string
		username           string
		password           string
		expectedStatusCode int
	}{
		"WebDAV disabled": {
			enableWebDAV:       false,
			username:           "admin",
			password:           password,
			expectedStatusCode: 403,
		},
		"WebDAV enabled, no auth": {
			enableWebDAV:       true,
			username:           "",
			password:           "",
			expectedStatusCode: 401,
		},
		"WebDAV enabled, wrong password": {
			enableWebDAV:       true,
			username:           "admin",
			password:           "wrong",
			expectedStatusCode: 401,
		},
		"WebDAV enabled, correct auth": {
			enableWebDAV:       true,
			username:           "admin",
			password:           password,
			expectedStatusCode: 207, // PROPFIND on root returns Multi-Status
		},
		"WebDAV enabled, no auth, OPTIONS": {
			enableWebDAV:       true,
			method:             "OPTIONS",
			username:           "",
			password:           "",
			expectedStatusCode: 200,
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			dbPath := filepath.Join(t.TempDir(), "db")
			db, err := storm.Open(dbPath)
			if err != nil {
				t.Fatalf("failed to open db: %v", err)
			}
			t.Cleanup(func() { db.Close() })

			st, err := bolt.NewStorage(db)
			if err != nil {
				t.Fatalf("failed to get storage: %v", err)
			}

			// Save auth method
			if err := st.Auth.Save(&auth.JSONAuth{}); err != nil {
				t.Fatalf("failed to save auth: %v", err)
			}

			// Save user
			user := &users.User{
				Username: "admin",
				Password: hashedPassword,
				Perm:     users.Permissions{Admin: true},
			}
			if err := st.Users.Save(user); err != nil {
				t.Fatalf("failed to save user: %v", err)
			}

			// Mock filesystem that satisfies BasePathFs assertion
			memFs := afero.NewMemMapFs()
			baseFs := afero.NewBasePathFs(memFs, "/")

			// Custom user store to inject FS
			st.Users = &customFSUser{
				Store: st.Users,
				fs:    baseFs,
			}

			server := &settings.Server{
				Root: "/",
			}

			// Mock settings
			if err := st.Settings.Save(&settings.Settings{
				Key:        []byte("key"),
				AuthMethod: auth.MethodJSONAuth,
			}); err != nil {
				t.Fatalf("failed to save settings: %v", err)
			}

			method := "PROPFIND"
			if tc.method != "" {
				method = tc.method
			}
			req := httptest.NewRequest(method, "/webdav/", nil)
			if tc.username != "" {
				req.SetBasicAuth(tc.username, tc.password)
			}

			recorder := httptest.NewRecorder()

			// We handle directly with monkey wrapper logic simulated or just call webDavHandler with populated data
			// Since webDavHandler calls d.store etc, we can construct data manually.

			d := &data{
				server:   server,
				store:    st,
				settings: &settings.Settings{AuthMethod: auth.MethodJSONAuth, EnableWebDAV: tc.enableWebDAV},
			}

			// webDavHandler signature: func(w http.ResponseWriter, r *http.Request, d *data) (int, error)
			status, err := webDavHandler(recorder, req, d)

			// If status is 0, it means handler handled it (written to response)
			// If webDavHandler returns 0, nil, we check recorder code.

			if status != 0 {
				if status != tc.expectedStatusCode {
					t.Errorf("expected status %d, got %d", tc.expectedStatusCode, status)
				}
			} else {
				// Handler wrote response
				res := recorder.Result()
				if res.StatusCode != tc.expectedStatusCode {
					t.Errorf("expected status %d, got %d", tc.expectedStatusCode, res.StatusCode)
				}
			}
		})
	}
}
