package filemanager

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

type modifySettingsRequest struct {
	*modifyRequest
	Data struct {
		Commands map[string][]string               `json:"commands"`
		Plugins  map[string]map[string]interface{} `json:"plugins"`
	} `json:"data"`
}

type pluginOption struct {
	Variable string      `json:"variable"`
	Name     string      `json:"name"`
	Value    interface{} `json:"value"`
}

func parsePutSettingsRequest(r *http.Request) (*modifySettingsRequest, error) {
	// Checks if the request body is empty.
	if r.Body == nil {
		return nil, errEmptyRequest
	}

	// Parses the request body and checks if it's well formed.
	mod := &modifySettingsRequest{}
	err := json.NewDecoder(r.Body).Decode(mod)
	if err != nil {
		return nil, err
	}

	// Checks if the request type is right.
	if mod.What != "settings" {
		return nil, errWrongDataType
	}

	return mod, nil
}

func settingsHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	if r.URL.Path != "" && r.URL.Path != "/" {
		return http.StatusNotFound, nil
	}

	switch r.Method {
	case http.MethodGet:
		return settingsGetHandler(c, w, r)
	case http.MethodPut:
		return settingsPutHandler(c, w, r)
	}

	return http.StatusMethodNotAllowed, nil
}

type settingsGetRequest struct {
	Commands map[string][]string       `json:"commands"`
	Plugins  map[string][]pluginOption `json:"plugins"`
}

func settingsGetHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	if !c.User.Admin {
		return http.StatusForbidden, nil
	}

	result := &settingsGetRequest{
		Commands: c.Commands,
		Plugins:  map[string][]pluginOption{},
	}

	for name, p := range c.Plugins {
		result.Plugins[name] = []pluginOption{}

		t := reflect.TypeOf(p).Elem()
		for i := 0; i < t.NumField(); i++ {
			result.Plugins[name] = append(result.Plugins[name], pluginOption{
				Variable: t.Field(i).Name,
				Name:     t.Field(i).Tag.Get("name"),
				Value:    reflect.ValueOf(p).Elem().FieldByName(t.Field(i).Name).Interface(),
			})
		}
	}

	return renderJSON(w, result)
}

func settingsPutHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	if !c.User.Admin {
		return http.StatusForbidden, nil
	}

	mod, err := parsePutSettingsRequest(r)
	if err != nil {
		return http.StatusBadRequest, err
	}
	// Update the commands.
	if mod.Which == "commands" {
		if err := c.db.Set("config", "commands", mod.Data.Commands); err != nil {
			return http.StatusInternalServerError, err
		}

		c.Commands = mod.Data.Commands
		return http.StatusOK, nil
	}

	// Update the plugins.
	if mod.Which == "plugins" {
		for name, plugin := range mod.Data.Plugins {
			err = mapstructure.Decode(plugin, c.Plugins[name])
			if err != nil {
				return http.StatusInternalServerError, err
			}

			err = c.db.Set("plugins", name, c.Plugins[name])
			if err != nil {
				return http.StatusInternalServerError, err
			}
		}

		return http.StatusOK, nil
	}

	return http.StatusMethodNotAllowed, nil
}
