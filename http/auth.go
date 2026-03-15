package fbhttp

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	fbAuth "github.com/filebrowser/filebrowser/v2/auth"
	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/users"
)

const (
	DefaultTokenExpirationTime = time.Hour * 2
)

type userInfo struct {
	ID                    uint              `json:"id"`
	Locale                string            `json:"locale"`
	ViewMode              users.ViewMode    `json:"viewMode"`
	SingleClick           bool              `json:"singleClick"`
	RedirectAfterCopyMove bool              `json:"redirectAfterCopyMove"`
	Perm                  users.Permissions `json:"perm"`
	Commands              []string          `json:"commands"`
	LockPassword          bool              `json:"lockPassword"`
	HideDotfiles          bool              `json:"hideDotfiles"`
	DateFormat            bool              `json:"dateFormat"`
	Username              string            `json:"username"`
	AceEditorTheme        string            `json:"aceEditorTheme"`
}

func userInfoFrom(user *users.User) userInfo {
	return userInfo{
		ID:                    user.ID,
		Locale:                user.Locale,
		ViewMode:              user.ViewMode,
		SingleClick:           user.SingleClick,
		RedirectAfterCopyMove: user.RedirectAfterCopyMove,
		Perm:                  user.Perm,
		LockPassword:          user.LockPassword,
		Commands:              user.Commands,
		HideDotfiles:          user.HideDotfiles,
		DateFormat:            user.DateFormat,
		Username:              user.Username,
		AceEditorTheme:        user.AceEditorTheme,
	}
}

func extractToken(r *http.Request) string {
	token := r.Header.Get("X-Auth")
	if token == "" {
		token = r.URL.Query().Get("auth")
	}
	return token
}

func withUser(fn handleFunc) handleFunc {
	return func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		tokenStr := extractToken(r)
		if tokenStr == "" {
			return http.StatusUnauthorized, nil
		}

		token, err := d.store.Tokens.Get(tokenStr)
		if err != nil {
			if errors.Is(err, fberrors.ErrNotExist) {
				return http.StatusUnauthorized, nil
			}
			return http.StatusInternalServerError, err
		}

		if token.IsExpired() {
			_ = d.store.Tokens.Delete(tokenStr)
			return http.StatusUnauthorized, nil
		}

		d.user, err = d.store.Users.Get(d.server.Root, token.UserID)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		return fn(w, r, d)
	}
}

func withAdmin(fn handleFunc) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		if !d.user.Perm.Admin {
			return http.StatusForbidden, nil
		}

		return fn(w, r, d)
	})
}

func loginHandler(tokenExpireTime time.Duration) handleFunc {
	return func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		auther, err := d.store.Auth.Get(d.settings.AuthMethod)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		user, err := auther.Auth(r, d.store.Users, d.settings, d.server)
		switch {
		case errors.Is(err, os.ErrPermission):
			return http.StatusForbidden, nil
		case err != nil:
			return http.StatusInternalServerError, err
		}

		return createAndReturnToken(w, d, user, tokenExpireTime)
	}
}

type signupBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var signupHandler = func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
	if !d.settings.Signup {
		return http.StatusMethodNotAllowed, nil
	}

	if r.Body == nil {
		return http.StatusBadRequest, nil
	}

	info := &signupBody{}
	err := json.NewDecoder(r.Body).Decode(info)
	if err != nil {
		return http.StatusBadRequest, err
	}

	if info.Password == "" || info.Username == "" {
		return http.StatusBadRequest, nil
	}

	user := &users.User{
		Username: info.Username,
	}

	d.settings.Defaults.Apply(user)

	// Users signed up via the signup handler should never become admins, even
	// if that is the default permission.
	user.Perm.Admin = false

	pwd, err := users.ValidateAndHashPwd(info.Password, d.settings.MinimumPasswordLength)
	if err != nil {
		return http.StatusBadRequest, err
	}

	user.Password = pwd
	if d.settings.CreateUserDir {
		user.Scope = ""
	}

	userHome, err := d.settings.MakeUserDir(user.Username, user.Scope, d.server.Root)
	if err != nil {
		log.Printf("create user: failed to mkdir user home dir: [%s]", userHome)
		return http.StatusInternalServerError, err
	}
	user.Scope = userHome
	log.Printf("new user: %s, home dir: [%s].", user.Username, userHome)

	err = d.store.Users.Save(user)
	if errors.Is(err, fberrors.ErrExist) {
		return http.StatusConflict, err
	} else if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func renewHandler(tokenExpireTime time.Duration) handleFunc {
	return withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {

		oldToken := extractToken(r)
		if oldToken != "" {
			_ = d.store.Tokens.Delete(oldToken)
		}

		return createAndReturnToken(w, d, d.user, tokenExpireTime)
	})
}

var logoutHandler = withUser(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
	tokenStr := extractToken(r)
	if tokenStr != "" {
		_ = d.store.Tokens.Delete(tokenStr)
	}
	return http.StatusOK, nil
})

var meHandler = withUser(func(w http.ResponseWriter, _ *http.Request, d *data) (int, error) {
	info := userInfoFrom(d.user)
	return renderJSON(w, nil, info)
})

func createAndReturnToken(w http.ResponseWriter, d *data, user *users.User, tokenExpirationTime time.Duration) (int, error) {
	tokenStr, err := fbAuth.GenerateToken()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	token := &fbAuth.Token{
		Token:     tokenStr,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(tokenExpirationTime),
		CreatedAt: time.Now(),
	}

	if err := d.store.Tokens.Save(token); err != nil {
		return http.StatusInternalServerError, err
	}

	w.Header().Set("Content-Type", "text/plain")
	if _, err := w.Write([]byte(tokenStr)); err != nil {
		return http.StatusInternalServerError, err
	}
	return 0, nil
}
