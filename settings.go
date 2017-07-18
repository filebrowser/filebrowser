package filemanager

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mitchellh/mapstructure"
)

func commandsHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	switch r.Method {
	case http.MethodGet:
		return commandsGetHandler(c, w, r)
	case http.MethodPut:
		return commandsPutHandler(c, w, r)
	}

	return http.StatusMethodNotAllowed, nil
}

func commandsGetHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	if !c.User.Admin {
		return http.StatusForbidden, nil
	}

	return renderJSON(w, c.FM.Commands)
}

func commandsPutHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	if !c.User.Admin {
		return http.StatusForbidden, nil
	}

	if r.Body == nil {
		return http.StatusBadGateway, errors.New("Empty request body")
	}

	var commands map[string][]string

	// Parses the user and checks for error.
	err := json.NewDecoder(r.Body).Decode(&commands)
	if err != nil {
		return http.StatusBadRequest, errors.New("Invalid JSON")
	}

	if err := c.FM.db.Set("config", "commands", commands); err != nil {
		return http.StatusInternalServerError, err
	}

	c.FM.Commands = commands
	return http.StatusOK, nil
}

func pluginsHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	switch r.Method {
	case http.MethodGet:
		return pluginsGetHandler(c, w, r)
	case http.MethodPut:
		return pluginsPutHandler(c, w, r)
	}

	return http.StatusMethodNotAllowed, nil
}

func pluginsGetHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	if !c.User.Admin {
		return http.StatusForbidden, nil
	}

	return renderJSON(w, c.FM.Plugins)
}

func pluginsPutHandler(c *RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	if !c.User.Admin {
		return http.StatusForbidden, nil
	}

	if r.Body == nil {
		return http.StatusBadGateway, errors.New("Empty request body")
	}

	var raw map[string]map[string]interface{}

	// Parses the user and checks for error.
	err := json.NewDecoder(r.Body).Decode(&raw)
	if err != nil {
		return http.StatusBadRequest, err
	}

	for name, plugin := range raw {
		err = mapstructure.Decode(plugin, c.FM.Plugins[name])
		if err != nil {
			return http.StatusInternalServerError, err
		}

		err = c.FM.db.Set("plugins", name, c.FM.Plugins[name])
		if err != nil {
			return http.StatusInternalServerError, err
		}
	}

	return http.StatusOK, nil
}
