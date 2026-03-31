package fbhttp

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
)

var brandingThemesHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	themesPath := d.settings.Branding.ThemesPath
	if themesPath == "" {
		return renderJSON(w, r, []string{})
	}

	entries, err := os.ReadDir(themesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return renderJSON(w, r, []string{})
		}
		return http.StatusInternalServerError, err
	}

	themes := []string{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Only include directories that contain a custom.css or img/ subdirectory
		cssPath := filepath.Join(themesPath, entry.Name(), "custom.css")
		imgPath := filepath.Join(themesPath, entry.Name(), "img")
		_, cssErr := os.Stat(cssPath)
		_, imgErr := os.Stat(imgPath)
		if cssErr == nil || imgErr == nil {
			themes = append(themes, entry.Name())
		}
	}

	sort.Strings(themes)
	return renderJSON(w, r, themes)
})
