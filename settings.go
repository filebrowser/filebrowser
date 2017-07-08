package filemanager

import (
	"encoding/json"
	"errors"
	"net/http"
)

func commandsHandler(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	switch r.Method {
	case http.MethodGet:
		return commandsGetHandler(c, w, r)
	case http.MethodPut:
		return commandsPutHandler(c, w, r)
	}

	return http.StatusMethodNotAllowed, nil
}

func commandsGetHandler(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	if !c.us.Admin {
		return http.StatusForbidden, nil
	}

	return renderJSON(w, c.fm.Commands)
}

func commandsPutHandler(c *requestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	if !c.us.Admin {
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

	if err := c.fm.db.Set("config", "commands", commands); err != nil {
		return http.StatusInternalServerError, err
	}

	c.fm.Commands = commands
	return http.StatusOK, nil
}
