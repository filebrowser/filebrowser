package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"

	fb "github.com/filebrowser/filebrowser/lib"
	"github.com/mitchellh/mapstructure"
)

type modifySettingsRequest struct {
	modifyRequest
	Data struct {
		CSS       string                 `json:"css"`
		Commands  map[string][]string    `json:"commands"`
		StaticGen map[string]interface{} `json:"staticGen"`
	} `json:"data"`
}

type option struct {
	Variable string      `json:"variable"`
	Name     string      `json:"name"`
	Value    interface{} `json:"value"`
}

func parsePutSettingsRequest(r *http.Request) (*modifySettingsRequest, error) {
	// Checks if the request body is empty.
	if r.Body == nil {
		return nil, fb.ErrEmptyRequest
	}

	// Parses the request body and checks if it's well formed.
	mod := &modifySettingsRequest{}
	err := json.NewDecoder(r.Body).Decode(mod)
	if err != nil {
		return nil, err
	}

	// Checks if the request type is right.
	if mod.What != "settings" {
		return nil, fb.ErrWrongDataType
	}

	return mod, nil
}

func settingsHandler(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
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
	CSS       string              `json:"css"`
	Commands  map[string][]string `json:"commands"`
	StaticGen []option            `json:"staticGen"`
}

func settingsGetHandler(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	if !c.User.Admin {
		return http.StatusForbidden, nil
	}

	result := &settingsGetRequest{
		Commands:  c.Commands,
		StaticGen: []option{},
		CSS:       c.CSS,
	}

	if c.StaticGen != nil {
		t := reflect.TypeOf(c.StaticGen).Elem()

		for i := 0; i < t.NumField(); i++ {
			if t.Field(i).Name[0] == bytes.ToLower([]byte{t.Field(i).Name[0]})[0] {
				continue
			}

			result.StaticGen = append(result.StaticGen, option{
				Variable: t.Field(i).Name,
				Name:     t.Field(i).Tag.Get("name"),
				Value:    reflect.ValueOf(c.StaticGen).Elem().FieldByName(t.Field(i).Name).Interface(),
			})
		}
	}

	return renderJSON(w, result)
}

func settingsPutHandler(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	if !c.User.Admin {
		return http.StatusForbidden, nil
	}

	mod, err := parsePutSettingsRequest(r)
	if err != nil {
		return http.StatusBadRequest, err
	}

	// Update the commands.
	if mod.Which == "commands" {
		if err := c.Store.Config.Save("commands", mod.Data.Commands); err != nil {
			return http.StatusInternalServerError, err
		}

		c.Commands = mod.Data.Commands
		return http.StatusOK, nil
	}

	// Update the global CSS.
	if mod.Which == "css" {
		if err := c.Store.Config.Save("css", mod.Data.CSS); err != nil {
			return http.StatusInternalServerError, err
		}

		c.CSS = mod.Data.CSS
		return http.StatusOK, nil
	}

	// Update the static generator options.
	if mod.Which == "staticGen" {
		err = mapstructure.Decode(mod.Data.StaticGen, c.StaticGen)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		err = c.Store.Config.Save("staticgen_"+c.StaticGen.Name(), c.StaticGen)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		return http.StatusOK, nil
	}

	return http.StatusMethodNotAllowed, nil
}
