package cnc

// Live event broadcast for /api/cnc/stream (Z-8). One in-process pub/sub
// — every WS handler gets its own buffered channel; if a subscriber falls
// behind we drop their oldest events instead of blocking the run loop.
// The streamer is the only producer, so there's no fan-in coordination.

// Event is one item on the WS feed. Type is one of:
//   - "line"   — streamer just wrote one line. n is 1-based.
//   - "status" — streamer state changed (started, stopped, errored).
//   - "metric" — aggregator updated a polled telemetry metric. The
//                Metric field carries the new snapshot for that key,
//                so subscribers can apply it directly without a
//                round-trip to /api/cnc/state.
//   - "log"    — informational; level + msg.
//   - "queue"  — the machine's NC job queue changed (add / remove /
//                reorder / state transition). Queue carries the
//                whole post-mutation snapshot so subscribers can
//                replace their local copy without a refetch.
//   - "dprnt"  — a Haas DPRNT[…] macro line emitted over RS-232 during
//                a streaming job. Text holds the line as received,
//                terminator stripped. Only emitted when the active
//                Machine has DPRNTCapture enabled.
type Event struct {
	Type   string      `json:"type"`
	N      int64       `json:"n,omitempty"`
	Text   string      `json:"text,omitempty"`
	Status *Status     `json:"status,omitempty"`
	Metric *Metric     `json:"metric,omitempty"`
	Level  string      `json:"level,omitempty"`
	Msg    string      `json:"msg,omitempty"`
	Queue  []QueueItem `json:"queue,omitempty"`
}

// subscriberBufferSize bounds how many events a slow subscriber can fall
// behind by before we drop their backlog. The WS path is fast (just
// JSON-encode + write), so 64 is generous; if the network stalls this
// gives the writer a few hundred ms to recover before we forfeit them.
const subscriberBufferSize = 64

type subscriber struct {
	ch chan Event
}

// Subscribe registers a new WS listener. Returned chan receives events
// until Unsubscribe is called or the streamer goes away. Drops oldest
// events on overflow rather than blocking — line streaming must never
// stall on a subscriber.
func (s *Streamer) Subscribe() <-chan Event {
	sub := &subscriber{ch: make(chan Event, subscriberBufferSize)}
	s.subsMu.Lock()
	s.subs = append(s.subs, sub)
	s.subsMu.Unlock()
	return sub.ch
}

// SubscriberCount returns the number of live WS listeners. Diagnostic
// only — polling gating is driven by Aggregator.Wake() now (operators
// touching state-bearing endpoints) rather than by raw subscriber
// count, since a header-pill subscriber on every authenticated layout
// would otherwise wake the bridge for every routine page view.
func (s *Streamer) SubscriberCount() int {
	s.subsMu.Lock()
	defer s.subsMu.Unlock()
	return len(s.subs)
}

// Unsubscribe removes a listener. The channel is closed so callers can
// range over it without leaking.
func (s *Streamer) Unsubscribe(ch <-chan Event) {
	s.subsMu.Lock()
	defer s.subsMu.Unlock()
	for i, sub := range s.subs {
		if sub.ch == ch {
			s.subs = append(s.subs[:i], s.subs[i+1:]...)
			close(sub.ch)
			return
		}
	}
}

// EmitQueueSnapshot pushes a "queue" event to every subscriber of
// this streamer. Exported so the HTTP layer can fan out after a queue
// mutation without taking a direct dep on the unexported emit().
func (s *Streamer) EmitQueueSnapshot(items []QueueItem) {
	s.emit(Event{Type: "queue", Queue: items})
}

func (s *Streamer) emit(ev Event) {
	s.subsMu.Lock()
	subs := s.subs
	s.subsMu.Unlock()
	for _, sub := range subs {
		select {
		case sub.ch <- ev:
		default:
			// Slow subscriber — drop their oldest, push the new one.
			// Saves memory pressure and keeps everyone moving.
			select {
			case <-sub.ch:
			default:
			}
			select {
			case sub.ch <- ev:
			default:
			}
		}
	}
}

// subsMu and subs are declared on Streamer in streamer.go; this file
// is the user of those fields, not the owner.
