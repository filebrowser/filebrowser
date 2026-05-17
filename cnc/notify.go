package cnc

// Discord notifications — admin wires a bot token + channel ID + a
// list of opt-in categories in Settings → External. When events on the
// streamer match an enabled category, the dispatcher posts a single
// channel message via the Discord HTTP API.
//
// Why bot, not webhook: the user wants explicit credentials they can
// rotate, and a bot can post to multiple channels with a single token
// if we extend later. Webhook is simpler but each one is single-
// channel and harder to audit.
//
// Anti-spam: per-category timestamps prevent a hot loop on the same
// category (a flapping alarm fires once per minute, not every event).

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/filebrowser/filebrowser/v2/settings"
)

const (
	NotifyCategoryMachineInfo     = "machine_info"
	NotifyCategoryFailures        = "failures"
	NotifyCategoryOperationStarts = "operation_starts"
)

// AllNotifyCategories is the canonical list — surfaced through the
// admin UI as checkboxes.
var AllNotifyCategories = []string{
	NotifyCategoryMachineInfo,
	NotifyCategoryFailures,
	NotifyCategoryOperationStarts,
}

// notifyMinInterval per category bounds the post rate so a flaky
// alarm or status flap doesn't spam the channel.
const notifyMinInterval = 60 * time.Second

// discordPostTimeout bounds an individual HTTPS POST so a dead bot
// doesn't wedge the dispatcher.
const discordPostTimeout = 5 * time.Second

// Notifier owns the per-config Discord client + last-post bookkeeping.
// Safe to share across goroutines (single mu).
type Notifier struct {
	settings settingsReader

	mu       sync.Mutex
	lastPost map[string]time.Time // keyed by category
}

// NewNotifier builds a notifier bound to the live settings reader so
// every Send picks up the current config without a restart.
func NewNotifier(s settingsReader) *Notifier {
	return &Notifier{
		settings: s,
		lastPost: map[string]time.Time{},
	}
}

// Send posts `content` to Discord when:
//   - the global Discord config is wired (token + channel + ≥1 category)
//   - `category` is in the admin's enabled list
//   - the same category hasn't fired within notifyMinInterval
//
// Returns nil on a no-op (category disabled / rate-limited) AND on
// successful post. The error path is reserved for actual transport /
// API failures the operator should know about.
//
// Best-effort by design — the dispatcher logs failures but never
// retries; a streamer event that misses the notify post is gone.
func (n *Notifier) Send(ctx context.Context, category, content string) error {
	if n == nil {
		return nil
	}
	cfg, err := n.config()
	if err != nil || !cfg.Enabled() || !cfg.CategoryEnabled(category) {
		return nil
	}
	if n.rateLimited(category) {
		return nil
	}
	if err := postDiscord(ctx, cfg, content); err != nil {
		return fmt.Errorf("discord post: %w", err)
	}
	n.markSent(category)
	return nil
}

// SendTest bypasses the category check + rate-limit and posts a
// "test message" string. Used by the admin UI's "Send test" button.
// Still requires the config to be enabled — no point firing on an
// empty channel ID.
func (n *Notifier) SendTest(ctx context.Context, content string) error {
	if n == nil {
		return fmt.Errorf("notifier not initialised")
	}
	cfg, err := n.config()
	if err != nil {
		return err
	}
	if !cfg.Enabled() {
		return fmt.Errorf("Discord notifications are not configured (need bot token + channel ID + at least one category)")
	}
	return postDiscord(ctx, cfg, content)
}

func (n *Notifier) config() (settings.DiscordConfig, error) {
	s, err := n.settings.Get()
	if err != nil {
		return settings.DiscordConfig{}, err
	}
	return s.Cnc.Discord, nil
}

func (n *Notifier) rateLimited(category string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	last := n.lastPost[category]
	if last.IsZero() {
		return false
	}
	return time.Since(last) < notifyMinInterval
}

func (n *Notifier) markSent(category string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.lastPost[category] = time.Now()
}

// postDiscord serializes a {"content": ...} payload and POSTs it to
// the channel endpoint. Authorisation is the Bot token from cfg.
func postDiscord(ctx context.Context, cfg settings.DiscordConfig, content string) error {
	ctx, cancel := context.WithTimeout(ctx, discordPostTimeout)
	defer cancel()
	body, err := json.Marshal(map[string]string{"content": content})
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", cfg.ChannelID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bot "+cfg.BotToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	// Pull a short snippet of the body for the operator-side error
	// message. Discord returns JSON like {"message":"Missing Access","code":50001}.
	snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	return fmt.Errorf("status %d: %s", resp.StatusCode, string(snippet))
}
