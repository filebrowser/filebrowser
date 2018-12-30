package auth

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/filebrowser/filebrowser/types"
)

// MethodJSONAuth is used to identify json auth.
const MethodJSONAuth types.AuthMethod = "json"

type jsonCred struct {
	Password  string `json:"password"`
	Username  string `json:"username"`
	ReCaptcha string `json:"recaptcha"`
}

// JSONAuth is a json implementaion of an auther.
type JSONAuth struct {
	ReCaptcha *ReCaptcha
	Store     *types.UsersVerify `json:"-"`
}

// Auth authenticates the user via a json in content body.
func (a JSONAuth) Auth(r *http.Request) (*types.User, error) {
	var cred jsonCred

	if r.Body == nil {
		return nil, types.ErrNoPermission
	}

	err := json.NewDecoder(r.Body).Decode(&cred)
	if err != nil {
		return nil, types.ErrNoPermission
	}

	// If ReCaptcha is enabled, check the code.
	if a.ReCaptcha != nil && len(a.ReCaptcha.Secret) > 0 {
		ok, err := a.ReCaptcha.Ok(cred.ReCaptcha)

		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, types.ErrNoPermission
		}
	}

	u, err := a.Store.GetByUsername(cred.Username)
	if err != nil || !types.CheckPwd(cred.Password, u.Password) {
		return nil, types.ErrNoPermission
	}

	return u, nil
}

const reCaptchaAPI = "/recaptcha/api/siteverify"

// ReCaptcha identifies a recaptcha conenction.
type ReCaptcha struct {
	Host   string `json:"host"`
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

// Ok checks if a reCaptcha responde is correct.
func (r *ReCaptcha) Ok(response string) (bool, error) {
	body := url.Values{}
	body.Set("secret", r.Key)
	body.Add("response", response)

	client := &http.Client{}

	resp, err := client.Post(
		r.Host+reCaptchaAPI,
		"application/x-www-form-urlencoded",
		strings.NewReader(body.Encode()),
	)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	var data struct {
		Success bool `json:"success"`
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return false, err
	}

	return data.Success, nil
}
