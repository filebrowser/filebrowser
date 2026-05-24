package fbhttp

import (
	"net/url"
	"strings"

	"github.com/filebrowser/filebrowser/v2/settings"
)

func contentSecurityPolicy(cfg settings.Collabora) string {
	if !cfg.Enabled || strings.TrimSpace(cfg.URL) == "" {
		return `default-src 'self'; style-src 'unsafe-inline';`
	}

	origins := collaboraCSPOrigins(cfg.URL)
	parts := []string{
		"default-src 'self'",
		"style-src 'self' 'unsafe-inline'",
		"script-src 'self'",
		"img-src 'self' data: blob:",
		"font-src 'self' data:",
		"frame-src 'self' " + strings.Join(origins, " "),
		"connect-src 'self' " + strings.Join(origins, " "),
	}

	return strings.Join(parts, "; ") + ";"
}

func collaboraCSPOrigins(raw string) []string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return []string{}
	}

	origin := u.Scheme + "://" + u.Host
	wsScheme := "ws"
	if u.Scheme == "https" {
		wsScheme = "wss"
	}

	return []string{origin, wsScheme + "://" + u.Host}
}
