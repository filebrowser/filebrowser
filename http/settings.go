package fbhttp

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/filebrowser/filebrowser/v2/rules"
	"github.com/filebrowser/filebrowser/v2/settings"
)

type settingsData struct {
	Signup                bool                  `json:"signup"`
	HideLoginButton       bool                  `json:"hideLoginButton"`
	CreateUserDir         bool                  `json:"createUserDir"`
	MinimumPasswordLength uint                  `json:"minimumPasswordLength"`
	UserHomeBasePath      string                `json:"userHomeBasePath"`
	Defaults              settings.UserDefaults `json:"defaults"`
	AuthMethod            settings.AuthMethod   `json:"authMethod"`
	Rules                 []rules.Rule          `json:"rules"`
	Branding              settings.Branding     `json:"branding"`
	Tus                   settings.Tus          `json:"tus"`
	Collabora             settings.Collabora    `json:"collabora"`
	ConvertX              settings.ConvertX     `json:"convertx"`
	ClamAV                settings.ClamAV       `json:"clamav"`
	Shell                 []string              `json:"shell"`
	Commands              map[string][]string   `json:"commands"`
}

var settingsGetHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	data := &settingsData{
		Signup:                d.settings.Signup,
		HideLoginButton:       d.settings.HideLoginButton,
		CreateUserDir:         d.settings.CreateUserDir,
		MinimumPasswordLength: d.settings.MinimumPasswordLength,
		UserHomeBasePath:      d.settings.UserHomeBasePath,
		Defaults:              d.settings.Defaults,
		AuthMethod:            d.settings.AuthMethod,
		Rules:                 d.settings.Rules,
		Branding:              d.settings.Branding,
		Tus:                   d.settings.Tus,
		Collabora:             d.collaboraConfig(),
		ConvertX:              d.settings.ConvertX,
		ClamAV:                d.settings.ClamAV,
		Shell:                 d.settings.Shell,
		Commands:              d.settings.Commands,
	}

	return renderJSON(w, r, data)
})

var settingsPutHandler = withAdmin(func(_ http.ResponseWriter, r *http.Request, d *data) (int, error) {
	req := &settingsData{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return http.StatusBadRequest, err
	}

	d.settings.Signup = req.Signup
	d.settings.CreateUserDir = req.CreateUserDir
	d.settings.MinimumPasswordLength = req.MinimumPasswordLength
	d.settings.UserHomeBasePath = req.UserHomeBasePath
	d.settings.Defaults = req.Defaults
	d.settings.Rules = req.Rules
	d.settings.Branding = req.Branding
	d.settings.Tus = req.Tus
	req.Collabora.Configured = true
	d.settings.Collabora = req.Collabora
	req.ConvertX.Configured = true
	d.settings.ConvertX = req.ConvertX
	d.settings.ClamAV = req.ClamAV
	d.settings.Shell = req.Shell
	d.settings.Commands = req.Commands
	d.settings.HideLoginButton = req.HideLoginButton

	err = d.store.Settings.Save(d.settings)
	return errToStatus(err), err
})

var settingsClamAVTestHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	cfg := d.settings.ClamAV

	if r.Body != nil {
		defer r.Body.Close()
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&cfg); err != nil && !errors.Is(err, io.EOF) {
			return http.StatusBadRequest, err
		}
	}

	if err := testClamAVConnection(r.Context(), cfg); err != nil {
		return clamAVHTTPStatus(err), err
	}

	return renderJSON(w, r, map[string]string{
		"status":  "OK",
		"message": "ClamAV connection successful",
	})
})

var settingsConvertXTestHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	cfg := d.settings.ConvertX

	if r.Body != nil {
		defer r.Body.Close()
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&cfg); err != nil && !errors.Is(err, io.EOF) {
			return http.StatusBadRequest, err
		}
	}

	if err := testConvertXConnection(r.Context(), cfg); err != nil {
		return http.StatusBadGateway, err
	}

	return renderJSON(w, r, map[string]string{
		"status":  "OK",
		"message": "ConvertX connection successful",
	})
})
