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
	r.Refresh()
	return r
}

// Queues returns the shared QueueStore. May be nil when persistence
// failed at boot — callers should nil-check.
func (r *Registry) Queues() *QueueStore { return r.queues }

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
				} else if ev.Status.Running && r.queues != nil {
					r.queues.MarkProgress(machineID, int(ev.Status.LineCurrent), int(ev.Status.LineTotal))
				}
			}
		}
	}
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
