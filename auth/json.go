package auth

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/filebrowser/filebrowser/lib"
)

// MethodJSONAuth is used to identify json auth.
const MethodJSONAuth lib.AuthMethod = "json"

type jsonCred struct {
	Password  string `json:"password"`
	Username  string `json:"username"`
	ReCaptcha string `json:"recaptcha"`
}

// JSONAuth is a json implementaion of an Auther.
type JSONAuth struct {
	ReCaptcha *ReCaptcha
	instance  *lib.FileBrowser
}

// Auth authenticates the user via a json in content body.
func (a *JSONAuth) Auth(r *http.Request) (*lib.User, error) {
	var cred jsonCred

	if r.Body == nil {
		return nil, lib.ErrNoPermission
	}

	err := json.NewDecoder(r.Body).Decode(&cred)
	if err != nil {
		return nil, lib.ErrNoPermission
	}

	// If ReCaptcha is enabled, check the code.
	if a.ReCaptcha != nil && len(a.ReCaptcha.Secret) > 0 {
		ok, err := a.ReCaptcha.Ok(cred.ReCaptcha)

		if err != nil {
			return nil, err
		}

		if !ok {
			return nil, lib.ErrNoPermission
		}
	}

	u, err := a.instance.GetUser(cred.Username)
	if err != nil || !lib.CheckPwd(cred.Password, u.Password) {
		return nil, lib.ErrNoPermission
	}

	return u, nil
}

// SetInstance attaches the instance to the auther.
func (a *JSONAuth) SetInstance(i *lib.FileBrowser) {
	a.instance = i
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
