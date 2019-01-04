package http

import (
	"encoding/json"
	"net/http"

	"github.com/filebrowser/filebrowser/rules"
	"github.com/filebrowser/filebrowser/settings"
)

type settingsData struct {
	Signup   bool                  `json:"signup"`
	Defaults settings.UserDefaults `json:"defaults"`
	Rules    []rules.Rule          `json:"rules"`
	Branding settings.Branding     `json:"branding"`
	Shell    []string              `json:"shell"`
	Commands map[string][]string   `json:"commands"`
}

func (e *env) settingsGetHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := e.getAdminUser(w, r)
	if !ok {
		return
	}

	settings, err := e.Settings.Get()
	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
		return
	}

	data := &settingsData{
		Signup:   settings.Signup,
		Defaults: settings.Defaults,
		Rules:    settings.Rules,
		Branding: settings.Branding,
		Shell:    settings.Shell,
		Commands: settings.Commands,
	}

	renderJSON(w, r, data)
}

func (e *env) settingsPutHandler(w http.ResponseWriter, r *http.Request) {
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

	settings, err := e.Settings.Get()
	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
		return
	}

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

	err = e.Settings.Save(settings)
	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
	}
}
