package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/filebrowser/filebrowser/types"
)

func (e *Env) loginHandler(w http.ResponseWriter, r *http.Request) {
	user, err := e.Auther.Auth(r)
	if err == types.ErrNoPermission {
		httpErr(w, r, http.StatusForbidden, nil)
	} else if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
	} else {
		e.printToken(w, r, user)
	}
}

type signupBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (e *Env) signupHandler(w http.ResponseWriter, r *http.Request) {
	if !e.Settings.Signup {
		httpErr(w, r, http.StatusForbidden, nil)
		return
	}

	if r.Body == nil {
		httpErr(w, r, http.StatusBadRequest, nil)
		return
	}

	info := &signupBody{}
	err := json.NewDecoder(r.Body).Decode(info)
	if err != nil {
		httpErr(w, r, http.StatusBadRequest, nil)
		return
	}

	if info.Password == "" || info.Username == "" {
		httpErr(w, r, http.StatusBadRequest, nil)
		return
	}

	user := &types.User{
		Username: info.Username,
		Locale:   e.Settings.Defaults.Locale,
		Perm:     e.Settings.Defaults.Perm,
		ViewMode: e.Settings.Defaults.ViewMode,
		Commands: e.Settings.Defaults.Commands,
		Scope:    e.Settings.Defaults.Scope,
	}

	pwd, err := types.HashPwd(info.Password)
	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
		return
	}

	user.Password = pwd
	err = e.Store.Users.Save(user)
	if err == types.ErrExist {
		httpErr(w, r, http.StatusConflict, nil)
		return
	} else if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
		return
	}

	httpErr(w, r, http.StatusOK, nil)
}

type userInfo struct {
	ID           uint              `json:"id"`
	Locale       string            `json:"locale"`
	ViewMode     types.ViewMode    `json:"viewMode"`
	Perm         types.Permissions `json:"perm"`
	LockPassword bool              `json:"lockPassword"`
}

type authToken struct {
	User userInfo `json:"user"`
	jwt.StandardClaims
}

type extractor []string

func (e extractor) ExtractToken(r *http.Request) (string, error) {
	token, _ := request.AuthorizationHeaderExtractor.ExtractToken(r)

	// Checks if the token isn't empty and if it contains two dots.
	// The former prevents incompatibility with URLs that previously
	// used basic auth.
	if token != "" && strings.Count(token, ".") == 2 {
		return token, nil
	}

	auth := r.URL.Query().Get("auth")
	if auth == "" {
		return "", request.ErrNoTokenInRequest
	}

	return auth, nil
}

func (e *Env) auth(next http.HandlerFunc) http.HandlerFunc {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return e.Settings.Key, nil
	}

	nextWithUser := func(w http.ResponseWriter, r *http.Request, id uint) {
		ctx := context.WithValue(r.Context(), keyUserID, id)
		next(w, r.WithContext(ctx))
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var tk authToken
		token, err := request.ParseFromRequestWithClaims(r, &extractor{}, &tk, keyFunc)

		if err != nil || !token.Valid {
			httpErr(w, r, http.StatusForbidden, nil)
			return
		}

		nextWithUser(w, r, tk.User.ID)
	}
}

func (e *Env) printToken(w http.ResponseWriter, r *http.Request, user *types.User) {
	claims := &authToken{
		User: userInfo{
			ID:           user.ID,
			Locale:       user.Locale,
			ViewMode:     user.ViewMode,
			Perm:         user.Perm,
			LockPassword: user.LockPassword,
		},
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
			Issuer:    "File Browser",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(e.Settings.Key)

	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
	} else {
		w.Header().Set("Content-Type", "cty")
		w.Write([]byte(signed))
	}
}
