package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/asdine/storm/v3"
	"github.com/spf13/afero"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/share"
	"github.com/filebrowser/filebrowser/v2/storage/bolt"
	"github.com/filebrowser/filebrowser/v2/users"
)

func TestPublicShareHandlerAuthentication(t *testing.T) {
	t.Parallel()

	const passwordBcrypt = "$2y$10$TFAmdCbyd/mEZDe5fUeZJu.MaJQXRTwdqb/IQV.eTn6dWrF58gCSe" //nolint:gosec
	testCases := map[string]struct {
		share              *share.Link
		req                *http.Request
		expectedStatusCode int
	}{
		"Public share, no auth required": {
			share:              &share.Link{Hash: "h", UserID: 1},
			req:                newHTTPRequest(t),
			expectedStatusCode: 200,
		},
		"Private share, no auth provided, 401": {
			share:              &share.Link{Hash: "h", UserID: 1, PasswordHash: passwordBcrypt, Token: "123"},
			req:                newHTTPRequest(t),
			expectedStatusCode: 401,
		},
		"Private share, authentication via token": {
			share:              &share.Link{Hash: "h", UserID: 1, PasswordHash: passwordBcrypt, Token: "123"},
			req:                newHTTPRequest(t, func(r *http.Request) { r.URL.RawQuery = "token=123" }),
			expectedStatusCode: 200,
		},
		"Private share, authentication via invalid token, 401": {
			share:              &share.Link{Hash: "h", UserID: 1, PasswordHash: passwordBcrypt, Token: "123"},
			req:                newHTTPRequest(t, func(r *http.Request) { r.URL.RawQuery = "token=1234" }),
			expectedStatusCode: 401,
		},
		"Private share, authentication via password": {
			share:              &share.Link{Hash: "h", UserID: 1, PasswordHash: passwordBcrypt, Token: "123"},
			req:                newHTTPRequest(t, func(r *http.Request) { r.Header.Set("X-SHARE-PASSWORD", "password") }),
			expectedStatusCode: 200,
		},
		"Private share, authentication via invalid password, 401": {
			share:              &share.Link{Hash: "h", UserID: 1, PasswordHash: passwordBcrypt, Token: "123"},
			req:                newHTTPRequest(t, func(r *http.Request) { r.Header.Set("X-SHARE-PASSWORD", "wrong-password") }),
			expectedStatusCode: 401,
		},
	}

	for name, tc := range testCases {
		for handlerName, handler := range map[string]handleFunc{"public share handler": publicShareHandler, "public dl handler": publicDlHandler} {
			name, tc, handlerName, handler := name, tc, handlerName, handler
			t.Run(fmt.Sprintf("%s: %s", handlerName, name), func(t *testing.T) {
				t.Parallel()

				dbPath := filepath.Join(t.TempDir(), "db")
				db, err := storm.Open(dbPath)
				if err != nil {
					t.Fatalf("failed to open db: %v", err)
				}

				t.Cleanup(func() {
					if err := db.Close(); err != nil { //nolint:govet
						t.Errorf("failed to close db: %v", err)
					}
				})

				storage, err := bolt.NewStorage(db)
				if err != nil {
					t.Fatalf("failed to get storage: %v", err)
				}
				if err := storage.Share.Save(tc.share); err != nil {
					t.Fatalf("failed to save share: %v", err)
				}
				if err := storage.Users.Save(&users.User{Username: "username", Password: "pw"}); err != nil {
					t.Fatalf("failed to save user: %v", err)
				}
				if err := storage.Settings.Save(&settings.Settings{Key: []byte("key")}); err != nil {
					t.Fatalf("failed to save settings: %v", err)
				}

				storage.Users = &customFSUser{
					Store: storage.Users,
					fs:    &afero.MemMapFs{},
				}

				recorder := httptest.NewRecorder()
				handler := handle(handler, "", storage, &settings.Server{})

				handler.ServeHTTP(recorder, tc.req)
				result := recorder.Result()
				defer result.Body.Close()
				if result.StatusCode != tc.expectedStatusCode {
					t.Errorf("expected status code %d, got status code %d", tc.expectedStatusCode, result.StatusCode)
				}
			})
		}
	}
}

func newHTTPRequest(t *testing.T, requestModifiers ...func(*http.Request)) *http.Request {
	t.Helper()
	r, err := http.NewRequest(http.MethodGet, "h", http.NoBody)
	if err != nil {
		t.Fatalf("failed to construct request: %v", err)
	}
	for _, modify := range requestModifiers {
		modify(r)
	}
	return r
}

type customFSUser struct {
	users.Store
	fs afero.Fs
}

func (cu *customFSUser) Get(baseScope string, id interface{}) (*users.User, error) {
	user, err := cu.Store.Get(baseScope, id)
	if err != nil {
		return nil, err
	}
	user.Fs = cu.fs

	return user, nil
}
