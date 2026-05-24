package fbhttp

import (
	"strings"

	"github.com/filebrowser/filebrowser/v2/settings"
)

func effectiveCollabora(server *settings.Server, set *settings.Settings) settings.Collabora {
	if set != nil && set.Collabora.Configured {
		cfg := set.Collabora
		cfg.URL = strings.TrimRight(strings.TrimSpace(cfg.URL), "/")
		cfg.PublicURL = strings.TrimRight(strings.TrimSpace(cfg.PublicURL), "/")
		cfg.InternalURL = strings.TrimRight(strings.TrimSpace(cfg.InternalURL), "/")
		cfg.WOPISecret = strings.TrimSpace(cfg.WOPISecret)
		cfg.TokenTTL = strings.TrimSpace(cfg.TokenTTL)
		return cfg
	}

	cfg := settings.Collabora{
		Configured:  true,
		Enabled:     server.CollaboraEnabled,
		URL:         strings.TrimRight(strings.TrimSpace(server.CollaboraURL), "/"),
		PublicURL:   strings.TrimRight(strings.TrimSpace(server.CollaboraPublicURL), "/"),
		InternalURL: strings.TrimRight(strings.TrimSpace(server.CollaboraInternalURL), "/"),
		WOPISecret:  strings.TrimSpace(server.CollaboraWOPISecret),
		TokenTTL:    strings.TrimSpace(server.CollaboraTokenTTL),
	}
	if cfg.TokenTTL == "" {
		cfg.TokenTTL = "2h"
	}
	return cfg
}

func (d *data) collaboraConfig() settings.Collabora {
	return effectiveCollabora(d.server, d.settings)
}
