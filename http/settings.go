package http

import (
	"encoding/json"
	"net/http"

	"github.com/filebrowser/filebrowser/lib"
	"github.com/jinzhu/copier"
)

type settingsData struct {
	Signup   bool                `json:"signup"`
	Defaults lib.UserDefaults  `json:"defaults"`
	Rules    []lib.Rule        `json:"rules"`
	Branding lib.Branding      `json:"branding"`
	Shell    []string            `json:"shell"`
	Commands map[string][]string `json:"commands"`
}

func (e *Env) settingsGetHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := e.getAdminUser(w, r)
	if !ok {
		return
	}

	e.RLockSettings()
	defer e.RUnlockSettings()

	data := &settingsData{
		Signup:   e.GetSettings().Signup,
		Defaults: e.GetSettings().Defaults,
		Rules:    e.GetSettings().Rules,
		Branding: e.GetSettings().Branding,
		Shell:    e.GetSettings().Shell,
		Commands: e.GetSettings().Commands,
	}

	renderJSON(w, r, data)
}

func (e *Env) settingsPutHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := e.getAdminUser(w, r)
	if !ok {
		return
	}

	req := &settingsData{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		httpErr(w, r, http.StatusBadRequest, err)
		return
	}

	e.RLockSettings()
	settings := &lib.Settings{}
	err = copier.Copy(settings, e.GetSettings())
	e.RUnlockSettings()

	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
		return
	}

	settings.Signup = req.Signup
	settings.Defaults = req.Defaults
	settings.Rules = req.Rules
	settings.Branding = req.Branding
	settings.Shell = req.Shell
	settings.Commands = req.Commands

	err = e.SaveSettings(settings)
	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
	}
}
