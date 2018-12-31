package http

import (
	"encoding/json"
	"net/http"

	"github.com/filebrowser/filebrowser/types"
	"github.com/jinzhu/copier"
)

type settingsData struct {
	Signup   bool                `json:"signup"`
	Defaults types.UserDefaults  `json:"defaults"`
	Rules    []types.Rule        `json:"rules"`
	Branding types.Branding      `json:"branding"`
	Commands map[string][]string `json:"commands"`
}

func (e *Env) settingsGetHandler(w http.ResponseWriter, r *http.Request) {
	_, ok := e.getAdminUser(w, r)
	if !ok {
		return
	}

	data := &settingsData{
		Signup:   e.Settings.Signup,
		Defaults: e.Settings.Defaults,
		Rules:    e.Settings.Rules,
		Branding: e.Settings.Branding,
		Commands: e.Runner.Commands,
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

	e.mux.Lock()
	defer e.mux.Unlock()

	runner := &types.Runner{Commands: req.Commands}
	err = e.Store.Config.SaveRunner(runner)
	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
		return
	}

	settings := &types.Settings{}
	err = copier.Copy(settings, e.Settings)
	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
		return
	}

	settings.Signup = req.Signup
	settings.Defaults = req.Defaults
	settings.Rules = req.Rules
	settings.Branding = req.Branding

	err = e.Store.Config.SaveSettings(settings)
	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
		return
	}

	e.Runner = runner
	e.Settings = settings
}
