package cnc

// Registry holds per-Machine Streamer + Aggregator pairs. The HTTP
// layer looks up which pair to use by Machine.ID (from
// ?machine_id=...); when no ID is supplied, the first machine in
// settings is the default — preserves single-machine behavior.
//
// Refresh() is safe to call after a settings change to pick up new
// machines / drop removed ones. New machines get a fresh Streamer
// + Aggregator and the aggregator starts immediately. Removed
// machines have their goroutines torn down.

import (
	"context"
	"fmt"
	"sync"

	"github.com/filebrowser/filebrowser/v2/settings"
)

// Registry coordinates the per-machine instances.
type Registry struct {
	settings settingsReader

	mu          sync.RWMutex
	streamers   map[string]*Streamer
	aggregators map[string]*Aggregator
	queues      *QueueStore
	notifier    *Notifier
	library     *LibraryStore
	bgCtx       context.Context
	bgCancel    context.CancelFunc
}

// NewRegistry instantiates Streamer + Aggregator pairs for every
// configured Machine and starts the aggregators. The returned Registry
// is the single source of truth the HTTP layer consults.
func NewRegistry(s settingsReader) *Registry {
	ctx, cancel := context.WithCancel(context.Background())
	r := &Registry{
		settings:    s,
		streamers:   make(map[string]*Streamer),
		aggregators: make(map[string]*Aggregator),
		bgCtx:       ctx,
		bgCancel:    cancel,
	}
	// Best-effort queue load. A read-only filesystem or permission
	// error here doesn't bring the install down — the install just
	// runs without persisted queues, mutations log a warning.
	if qs, err := NewQueueStore("", nil); err == nil {
		r.queues = qs
	}
	// Same shape for the tool library — load is best-effort, a bad
	// file just means the panel falls back to live-only Q-code data.
	if ls, err := NewLibraryStore(""); err == nil {
		r.library = ls
	}
	r.notifier = NewNotifier(s)
	r.Refresh()
	return r
}

// Notifier returns the shared Discord notifier. Always non-nil; the
// notifier itself no-ops when DiscordConfig isn't fully wired.
func (r *Registry) Notifier() *Notifier { return r.notifier }

// Queues returns the shared QueueStore. May be nil when persistence
// failed at boot — callers should nil-check.
func (r *Registry) Queues() *QueueStore { return r.queues }

// LibraryStore returns the shared tool-library store. May be nil
// when persistence failed at boot.
func (r *Registry) LibraryStore() *LibraryStore { return r.library }

// Refresh diffs settings.Cnc.Machines against the live registry,
// adding pairs for new IDs and stopping pairs for removed ones.
// Safe to call from a settings-changed handler; no-op if nothing
// changed.
func (r *Registry) Refresh() {
	set, err := r.settings.Get()
	if err != nil {
		return
	}
	wanted := make(map[string]struct{}, len(set.Cnc.Machines))
	for _, m := range set.Cnc.Machines {
		wanted[m.ID] = struct{}{}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Add missing.
	for _, m := range set.Cnc.Machines {
		if _, exists := r.streamers[m.ID]; exists {
			continue
		}
		st := New(r.settings, m.ID)
		ag := NewAggregator(st)
		ag.Start(r.bgCtx)
		r.streamers[m.ID] = st
		r.aggregators[m.ID] = ag
		// Spin a watcher that listens to this streamer's event feed
		// and auto-promotes a queue row to "running" whenever the
		// controller's program metric changes. Catches the SD-card /
		// USB / pre-existing-program case the spec calls out.
		if r.queues != nil {
			go r.watchQueueAutoMatch(r.bgCtx, m.ID, st)
		}
	}
	// Remove orphaned.
	for id, ag := range r.aggregators {
		if _, ok := wanted[id]; ok {
			continue
		}
		ag.Stop()
		delete(r.aggregators, id)
	}
	for id := range r.streamers {
		if _, ok := wanted[id]; ok {
			continue
		}
		// Streamer has no Stop on its lifecycle (only on the active
		// job); deleting the reference releases the GC-eligible
		// instance. If a job were running on a removed machine, that
		// goroutine continues until it finishes or hits its own
		// cancel — same as today's single-machine behavior.
		delete(r.streamers, id)
	}
}

// Stop tears down all aggregators. Called at process shutdown.
func (r *Registry) Stop() {
	r.bgCancel()
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, ag := range r.aggregators {
		ag.Stop()
	}
}

// watchQueueAutoMatch subscribes to a machine's WS event feed and
// promotes whichever queue row's O-number matches the controller's
// currently-executing program. Covers the operator-started-via-SD-
// card case the spec calls out — Q500 reflects whatever the panel is
// running, regardless of who sent the file.
//
// Runs for the lifetime of the registry; exits when the parent
// context is cancelled.
func (r *Registry) watchQueueAutoMatch(ctx context.Context, machineID string, st *Streamer) {
	feed := st.Subscribe()
	defer st.Unsubscribe(feed)
	var lastProgram string
	var wasRunning bool
	var lastAttachedFile string
	machineLabel := r.machineLabel(machineID)
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-feed:
			if !ok {
				return
			}
			switch ev.Type {
			case "metric":
				if ev.Metric == nil || ev.Metric.Key != "status_combined" {
					continue
				}
				prog := extractProgram(ev.Metric)
				if prog == "" || prog == lastProgram {
					continue
				}
				lastProgram = prog
				if r.queues != nil {
					r.queues.PromoteByONumber(machineID, prog)
					st.EmitQueueSnapshot(r.queues.List(machineID))
					// Auto-attach: when the controller reports a new
					// O-number, mark the matching queue item's file as
					// attached so /machine can follow along even when the
					// program came from SD card / USB / Ethernet drop.
					// Manual attaches win — AttachAuto is a no-op if a
					// manual attach is already active. The downstream
					// status event triggers the existing attached-file
					// Discord notification (source = "auto") so no extra
					// notify call is needed here.
					if it := r.queues.FindByONumber(machineID, prog); it != nil {
						if st.AttachAuto(it.FilePath) {
							st.EmitStatus()
						}
					}
				}
			case "status":
				if ev.Status == nil {
					continue
				}
				// Streamer reports idle → the local-send job has ended.
				// Demote any sending/running row. The next status_combined
				// metric will re-promote if the controller is actually
				// running something else.
				if !ev.Status.Running && r.queues != nil {
					r.queues.ClearInFlight(machineID)
					st.EmitQueueSnapshot(r.queues.List(machineID))
					lastProgram = ""
					// Auto-attaches are tied to the running program;
					// clear them on idle so a stale entry doesn't keep
					// /machine pointed at the wrong file. Manual
					// attaches stay until the operator detaches.
					if st.DetachAuto() {
						st.EmitStatus()
					}
				} else if ev.Status.Running && r.queues != nil {
					r.queues.MarkProgress(machineID, int(ev.Status.LineCurrent), int(ev.Status.LineTotal))
				}
				// Discord notifications — operator-initiated job starts
				// trip the operation_starts category; idle transitions
				// trip machine_info; manual attach also trips
				// operation_starts since it's an explicit "follow this
				// program" action.
				if ev.Status.Running && !wasRunning {
					r.notifyAsync(NotifyCategoryOperationStarts,
						fmt.Sprintf("▶️ %s — started %s (%d lines)",
							machineLabel, ev.Status.FilePath, ev.Status.LineTotal))
				} else if !ev.Status.Running && wasRunning {
					r.notifyAsync(NotifyCategoryMachineInfo,
						fmt.Sprintf("⏹️ %s — job ended at line %d",
							machineLabel, ev.Status.LineCurrent))
				}
				wasRunning = ev.Status.Running
				if ev.Status.AttachedFile != "" && ev.Status.AttachedFile != lastAttachedFile {
					r.notifyAsync(NotifyCategoryOperationStarts,
						fmt.Sprintf("🔗 %s — attached to %s (%s)",
							machineLabel, ev.Status.AttachedFile, ev.Status.AttachedSource))
					// Attached state changing means the operator just
					// initiated a follow-along (manually, post-send, or
					// O-number match). Wake the aggregator so position
					// + current_block metrics keep flowing for the
					// duration of the panel-cycle execution.
					if ag, _ := r.Aggregator(machineID); ag != nil {
						ag.Wake(0)
					}
				}
				lastAttachedFile = ev.Status.AttachedFile
				if ev.Status.HaasLastError != "" {
					r.notifyAsync(NotifyCategoryFailures,
						fmt.Sprintf("⚠️ %s — %s",
							machineLabel, ev.Status.HaasLastError))
				}
			case "log":
				if ev.Level == "error" {
					r.notifyAsync(NotifyCategoryFailures,
						fmt.Sprintf("⚠️ %s — %s", machineLabel, ev.Msg))
				}
			}
		}
	}
}

// notifyAsync fires the Discord post off the watcher goroutine. The
// notifier itself rate-limits per category; we don't block the event
// loop on a network call.
func (r *Registry) notifyAsync(category, content string) {
	if r.notifier == nil {
		return
	}
	go func() {
		_ = r.notifier.Send(r.bgCtx, category, content)
	}()
}

// machineLabel renders "Name" or just the ID when no name is set.
// Avoids a settings.Get() per event by caching at watcher-spawn time
// — operators rarely rename mid-session.
func (r *Registry) machineLabel(id string) string {
	s, err := r.settings.Get()
	if err != nil {
		return id
	}
	for _, m := range s.Cnc.Machines {
		if m.ID == id {
			if m.Name != "" {
				return m.Name
			}
			return id
		}
	}
	return id
}

// extractProgram pulls the program ID (e.g. "O00057") out of a parsed
// status_combined metric. Q500 frames decode to either a struct with
// a "program" key, or a raw string we fall through to.
func extractProgram(m *Metric) string {
	if m == nil {
		return ""
	}
	if parsed, ok := m.Parsed.(map[string]any); ok {
		if v, ok := parsed["program"].(string); ok {
			return v
		}
	}
	return m.Value
}

// resolve returns the (streamer, aggregator) pair for the requested
// machine ID. Empty ID falls back to the configured default. Returns
// (nil, nil, "") if no machine matches.
func (r *Registry) resolve(machineID string) (*Streamer, *Aggregator, string) {
	if machineID == "" {
		set, err := r.settings.Get()
		if err != nil {
			return nil, nil, ""
		}
		machineID = set.Cnc.DefaultMachineID()
		if machineID == "" {
			return nil, nil, ""
		}
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	st, sok := r.streamers[machineID]
	ag, aok := r.aggregators[machineID]
	if !sok || !aok {
		return nil, nil, ""
	}
	return st, ag, machineID
}

// Streamer returns the Streamer for the given machine ID (or default
// if empty), and the resolved ID. Returns (nil, "") if no machine
// matches — handlers should map that to 404.
func (r *Registry) Streamer(machineID string) (*Streamer, string) {
	st, _, id := r.resolve(machineID)
	return st, id
}

// Aggregator returns the Aggregator for the given machine ID (or
// default if empty), and the resolved ID.
func (r *Registry) Aggregator(machineID string) (*Aggregator, string) {
	_, ag, id := r.resolve(machineID)
	return ag, id
}

// Machines returns the configured machine list at this moment. Snapshot;
// mutating the slice is safe (it's a copy).
func (r *Registry) Machines() []settings.Machine {
	set, err := r.settings.Get()
	if err != nil {
		return nil
	}
	out := make([]settings.Machine, len(set.Cnc.Machines))
	copy(out, set.Cnc.Machines)
	return out
}

// AnyHasPendingRecovery is true if at least one machine reports
// pendingRecovery. Header-pill / global banner can use this to nudge
// the operator without picking a specific machine.
func (r *Registry) AnyHasPendingRecovery() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, st := range r.streamers {
		st.mu.Lock()
		pending := st.pendingRecovery != nil
		st.mu.Unlock()
		if pending {
			return true
		}
	}
	return false
}
