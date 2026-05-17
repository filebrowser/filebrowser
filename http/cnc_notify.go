package fbhttp

// /api/cnc/notifications — admin-only management of the Discord
// notify config. GET masks the bot token (write-only field — admin
// rotates by pasting a new value). PUT replaces. POST .../test fires
// a one-off message so the admin can verify the bot has channel access
// without waiting for a real event.

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/filebrowser/filebrowser/v2/cnc"
	"github.com/filebrowser/filebrowser/v2/settings"
)

// botTokenMask is what GET returns in place of the real token. Frontend
// renders this verbatim so the admin sees "is something set" without
// the raw secret ever bouncing through the UI.
const botTokenMask = "********"

type cncNotificationsBody struct {
	BotToken         string   `json:"botToken,omitempty"`
	BotTokenIsMasked bool     `json:"botTokenIsMasked,omitempty"` // GET only
	ChannelID        string   `json:"channelId,omitempty"`
	Categories       []string `json:"categories,omitempty"`
	// Echo back the canonical category list so the frontend can build
	// the checkbox list without hardcoding strings.
	KnownCategories []string `json:"knownCategories,omitempty"` // GET only
}

func notificationsBody(d settings.DiscordConfig, mask bool) cncNotificationsBody {
	out := cncNotificationsBody{
		ChannelID:       d.ChannelID,
		Categories:      append([]string{}, d.Categories...),
		KnownCategories: append([]string{}, cnc.AllNotifyCategories...),
	}
	if d.BotToken != "" {
		if mask {
			out.BotToken = botTokenMask
			out.BotTokenIsMasked = true
		} else {
			out.BotToken = d.BotToken
		}
	}
	return out
}

var cncNotificationsGetHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	return renderJSON(w, r, notificationsBody(d.settings.Cnc.Discord, true))
})

// cncNotificationsPutHandler replaces the config. If the admin sends
// botToken=""  the existing token is preserved (otherwise an accidental
// "save with the form pre-filled with mask" wipes the token). To clear
// the token explicitly, send botToken="" AND categories=[].
func cncNotificationsPutHandler() handleFunc {
	return withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		req := &cncNotificationsBody{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			return http.StatusBadRequest, err
		}
		// Replace token only when the caller sent something other than
		// the masked placeholder. Empty AND non-mask = preserve.
		current := d.settings.Cnc.Discord
		newToken := current.BotToken
		switch req.BotToken {
		case "", botTokenMask:
			// keep existing
		default:
			newToken = strings.TrimSpace(req.BotToken)
		}
		// Validate categories against the known list — anything else
		// gets dropped silently so a future client can't sneak in a
		// junk tag that the dispatcher ignores anyway.
		known := map[string]bool{}
		for _, c := range cnc.AllNotifyCategories {
			known[c] = true
		}
		cleanCats := make([]string, 0, len(req.Categories))
		seen := map[string]bool{}
		for _, c := range req.Categories {
			c = strings.ToLower(strings.TrimSpace(c))
			if !known[c] || seen[c] {
				continue
			}
			seen[c] = true
			cleanCats = append(cleanCats, c)
		}
		d.settings.Cnc.Discord = settings.DiscordConfig{
			BotToken:   newToken,
			ChannelID:  strings.TrimSpace(req.ChannelID),
			Categories: cleanCats,
		}
		if err := d.store.Settings.Save(d.settings); err != nil {
			return http.StatusInternalServerError, err
		}
		return renderJSON(w, r, notificationsBody(d.settings.Cnc.Discord, true))
	})
}

// cncNotificationsTestHandler fires a one-off Discord post so the
// admin can verify the bot has channel-write access without waiting
// for a real event.
func cncNotificationsTestHandler(registry *cnc.Registry) handleFunc {
	return withAdmin(func(w http.ResponseWriter, r *http.Request, _ *data) (int, error) {
		notifier := registry.Notifier()
		if notifier == nil {
			return http.StatusServiceUnavailable, errors.New("notifier not initialised")
		}
		ctx, cancel := context.WithTimeout(r.Context(), 6*time.Second)
		defer cancel()
		if err := notifier.SendTest(ctx, "filebrowser-NC test message — Discord bot wired up ✅"); err != nil {
			return http.StatusBadGateway, err
		}
		return renderJSON(w, r, map[string]any{"ok": true})
	})
}
