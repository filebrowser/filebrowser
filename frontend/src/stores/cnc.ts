// Live CNC status — drives the global header pill (Z-12) and any
// future surfaces (sidebar entry, machine tracker in the 3D viewer).
//
// Two transport modes:
//   1. WebSocket /api/cnc/stream pushes line + status events; preferred.
//   2. Falls back to a 2s GET /api/cnc/status poll if the WS fails or
//      hasn't connected yet — keeps the pill correct on networks where
//      WS is blocked, and during the brief gap between page navigation
//      and the WS subscribe.

import { defineStore } from "pinia";
import { cnc as cncApi } from "@/api";
import type { CncMetric, CncStateSnapshot, CncStatus } from "@/api/cnc";
import { baseURL } from "@/utils/constants";

const POLL_INTERVAL_MS = 2000;
const RECONNECT_DELAY_MS = 3000;

// LogEntry is one line on the activity feed. Backend emits these via
// the WS `log` event; we also synthesize them client-side for line +
// status transitions so the panel reads chronologically.
export interface LogEntry {
  ts: number;
  level: string;
  msg: string;
}

const LOG_BUFFER_MAX = 200;
const LOG_PERSIST_KEY = "cncActivityLog";

// loadPersistedLog restores the activity feed across page reloads.
// Operators were losing all the dial / line / error entries on F5,
// which made postmortems on a flaky job pointless.
function loadPersistedLog(): LogEntry[] {
  try {
    const raw = localStorage.getItem(LOG_PERSIST_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed
      .filter(
        (e): e is LogEntry =>
          e &&
          typeof e.ts === "number" &&
          typeof e.level === "string" &&
          typeof e.msg === "string"
      )
      .slice(0, LOG_BUFFER_MAX);
  } catch {
    return [];
  }
}

let persistTimer: ReturnType<typeof setTimeout> | null = null;
function schedulePersist(entries: LogEntry[]) {
  if (persistTimer) return;
  // Coalesce — if a burst of events fires (e.g. 100-line beacon
  // batch on a long stream), one write per ~500 ms is plenty.
  persistTimer = setTimeout(() => {
    persistTimer = null;
    try {
      localStorage.setItem(LOG_PERSIST_KEY, JSON.stringify(entries));
    } catch {
      /* quota / private mode — drop silently */
    }
  }, 500);
}

interface CncState {
  running: boolean;
  filePath: string;
  fileURL: string;
  lineCurrent: number;
  lineTotal: number;
  haasOk: boolean;
  haasLastError: string;
  recoveryPending: boolean;
  recoveryFilePath: string;
  // Last raw status (handy if a future surface wants more fields).
  raw: CncStatus | null;
  // Internal: tells the pill component when to show "?", "running", "idle".
  initialized: boolean;
  // Live telemetry snapshot, populated from WS "metric" events with
  // an initial seed via /api/cnc/state. Keys mirror the backend's
  // metric-spec list (mode, spindle_actual, pos_x, …).
  metrics: CncStateSnapshot;
  metricsSeeded: boolean;
  // Rolling activity log — backend `log` events plus client-synthesized
  // entries for status / line milestones. Newest first.
  log: LogEntry[];
}

let pollTimer: ReturnType<typeof setInterval> | null = null;
let ws: WebSocket | null = null;
let wsRetry: ReturnType<typeof setTimeout> | null = null;
let started = false;

export const useCncStore = defineStore("cnc", {
  state: (): CncState => ({
    running: false,
    filePath: "",
    fileURL: "",
    lineCurrent: 0,
    lineTotal: 0,
    haasOk: true,
    haasLastError: "",
    recoveryPending: false,
    recoveryFilePath: "",
    raw: null,
    initialized: false,
    metrics: {},
    metricsSeeded: false,
    log: loadPersistedLog(),
  }),
  actions: {
    applyStatus(s: CncStatus) {
      this.running = !!s.running;
      this.filePath = s.file_path ?? "";
      this.fileURL = s.file_url ?? "";
      this.lineCurrent = s.line_current ?? 0;
      this.lineTotal = s.line_total ?? 0;
      this.haasOk = s.haas_ok !== false;
      this.haasLastError = s.haas_last_error ?? "";
      this.recoveryPending = !!s.recovery_pending;
      this.recoveryFilePath = s.recovery_file_path ?? "";
      this.raw = s;
      this.initialized = true;
    },

    async ackRecovery() {
      await cncApi.ackRecovery();
      this.recoveryPending = false;
      this.recoveryFilePath = "";
      // Refresh from the server so the rest of the state matches reality.
      this.pollOnce();
    },

    // Seed the metrics snapshot once via /api/cnc/state so consumers
    // (e.g. Machine.vue) have something to render before the first WS
    // "metric" event lands. Idempotent — re-callers just refresh the
    // map.
    async seedMetrics() {
      try {
        this.metrics = await cncApi.getState();
        this.metricsSeeded = true;
      } catch {
        /* leave previous map; WS events will fill it in */
      }
    },

    applyMetric(m: CncMetric) {
      this.metrics = { ...this.metrics, [m.key]: m };
      this.metricsSeeded = true;
    },

    pushLog(level: string, msg: string) {
      this.log.unshift({ ts: Date.now(), level, msg });
      if (this.log.length > LOG_BUFFER_MAX) {
        this.log.length = LOG_BUFFER_MAX;
      }
      schedulePersist(this.log);
    },

    clearLog() {
      this.log = [];
      try {
        localStorage.removeItem(LOG_PERSIST_KEY);
      } catch {
        /* ignore */
      }
    },

    async pollOnce() {
      try {
        const s = await cncApi.getStatus();
        this.applyStatus(s);
      } catch {
        // Silent: if /api/cnc/status fails (typically because the user
        // is on a public share view without auth) we just leave the
        // pill hidden. Don't spam the toast layer for a passive query.
      }
    },

    // Start the poll loop + try to upgrade to WS. Idempotent — safe to
    // call from every layout that mounts.
    start() {
      if (started) return;
      started = true;
      this.pollOnce();
      pollTimer = setInterval(() => this.pollOnce(), POLL_INTERVAL_MS);
      this.connectWS();
    },

    stop() {
      started = false;
      if (pollTimer) {
        clearInterval(pollTimer);
        pollTimer = null;
      }
      if (wsRetry) {
        clearTimeout(wsRetry);
        wsRetry = null;
      }
      if (ws) {
        try {
          ws.close();
        } catch {
          /* ignore */
        }
        ws = null;
      }
    },

    connectWS() {
      // Build absolute ws:// or wss:// URL matching the current origin.
      const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
      const url = `${proto}//${window.location.host}${baseURL}/api/cnc/stream`;
      try {
        ws = new WebSocket(url);
      } catch {
        this.scheduleReconnect();
        return;
      }

      ws.addEventListener("message", (e) => {
        try {
          const ev = JSON.parse(e.data);
          if (ev.type === "status" && ev.status) {
            // Capture the current counters BEFORE applyStatus overwrites
            // them. When a job ends the server's status payload zeros
            // line_current/total, so reading them after applyStatus
            // would always log "0/0".
            const wasRunning = this.running;
            const prevLine = this.lineCurrent;
            const prevTotal = this.lineTotal;
            this.applyStatus(ev.status);
            if (this.running && !wasRunning) {
              this.pushLog("info", `running: ${this.filePath}`);
            } else if (!this.running && wasRunning) {
              this.pushLog("info", `idle (last: ${prevLine}/${prevTotal})`);
            }
            if (ev.status.haas_last_error) {
              // Surface server-side errors on the feed even if the
              // status pill is off-screen.
              this.pushLog("error", ev.status.haas_last_error);
            }
          } else if (ev.type === "line" && typeof ev.n === "number") {
            this.lineCurrent = ev.n;
            this.running = true;
          } else if (ev.type === "metric" && ev.metric) {
            this.applyMetric(ev.metric as CncMetric);
          } else if (ev.type === "log" && typeof ev.msg === "string") {
            this.pushLog(ev.level || "info", ev.msg);
          }
        } catch {
          /* ignore malformed frames */
        }
      });

      ws.addEventListener("close", () => {
        ws = null;
        if (started) this.scheduleReconnect();
      });

      ws.addEventListener("error", () => {
        // Let the close handler schedule reconnect; closing is idempotent.
        try {
          ws?.close();
        } catch {
          /* ignore */
        }
      });
    },

    scheduleReconnect() {
      if (wsRetry) return;
      wsRetry = setTimeout(() => {
        wsRetry = null;
        if (started) this.connectWS();
      }, RECONNECT_DELAY_MS);
    },
  },
});
