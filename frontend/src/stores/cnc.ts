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
import type {
  CncMachine,
  CncMetric,
  CncStateSnapshot,
  CncStatus,
  QueueItem,
  SendMethod,
} from "@/api/cnc";
import { baseURL } from "@/utils/constants";

// Persist the operator's selected machine across page reloads. When
// the same machine is still in the registry after restart, we restore
// it; otherwise we fall back to the server's default.
const CURRENT_MACHINE_KEY = "cncCurrentMachineId";

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
  // ── Multi-machine routing ────────────────────────────────────────
  // currentMachineId is what every API call key ed off of and what the
  // WS stream subscribes to. Empty string == "let the server pick the
  // default" (used on cold load before /api/cnc/machines responds).
  currentMachineId: string;
  // Full machine list from /api/cnc/machines. Cached so the switcher
  // dropdown and Send-destination dropdown can render without a
  // separate fetch on every component mount.
  machines: CncMachine[];
  defaultMachineId: string;
  machinesLoaded: boolean;

  running: boolean;
  filePath: string;
  fileURL: string;
  // Mode the operator picked when starting the job — "mem" (Receive
  // into NC memory then Cycle Start) or "dnc" (drip-feed). Echoed by
  // the backend on every status frame; surfaces in the activity log
  // and the in-flight progress strip.
  method: string;
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
  // Per-machine NC job queue. Replaced wholesale on every WS "queue"
  // event so the UI doesn't have to merge incremental mutations.
  queue: QueueItem[];
  queueLoaded: boolean;
}

let pollTimer: ReturnType<typeof setInterval> | null = null;
let ws: WebSocket | null = null;
let wsRetry: ReturnType<typeof setTimeout> | null = null;
let started = false;

export const useCncStore = defineStore("cnc", {
  state: (): CncState => ({
    currentMachineId: localStorage.getItem(CURRENT_MACHINE_KEY) || "",
    machines: [],
    defaultMachineId: "",
    machinesLoaded: false,
    running: false,
    filePath: "",
    fileURL: "",
    method: "",
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
    queue: [],
    queueLoaded: false,
  }),
  getters: {
    // currentMachine returns the CncMachine entry for the active id,
    // or undefined while the list is still loading. Components that
    // want to show "Haas VF-2 · 192.168.20.200" use this getter; raw
    // calls into the store should pass currentMachineId directly.
    currentMachine(state): CncMachine | undefined {
      return state.machines.find((m) => m.id === state.currentMachineId);
    },
  },
  actions: {
    applyStatus(s: CncStatus) {
      this.running = !!s.running;
      this.filePath = s.file_path ?? "";
      this.fileURL = s.file_url ?? "";
      this.method = s.method ?? "";
      this.lineCurrent = s.line_current ?? 0;
      this.lineTotal = s.line_total ?? 0;
      this.haasOk = s.haas_ok !== false;
      this.haasLastError = s.haas_last_error ?? "";
      this.recoveryPending = !!s.recovery_pending;
      this.recoveryFilePath = s.recovery_file_path ?? "";
      this.raw = s;
      this.initialized = true;
    },

    // Pull the machine list from /api/cnc/machines. Restores the
    // operator's last-selected machine if it still exists, falls back
    // to the server's default_id otherwise. Idempotent — components
    // that mount during the same session re-call it cheaply.
    async loadMachines() {
      try {
        const list = await cncApi.listMachines();
        this.machines = list.machines || [];
        this.defaultMachineId = list.default_id || "";
        // Reconcile the persisted selection with what's now in the
        // registry — if the operator's last machine was removed in
        // settings, drop back to the default to avoid 404s on every
        // call.
        const stillExists = this.machines.some(
          (m) => m.id === this.currentMachineId
        );
        if (!this.currentMachineId || !stillExists) {
          this.currentMachineId = this.defaultMachineId;
          try {
            localStorage.setItem(CURRENT_MACHINE_KEY, this.currentMachineId);
          } catch {
            /* quota — ignore */
          }
        }
        this.machinesLoaded = true;
      } catch {
        /* leave existing list; subsequent calls will retry */
      }
    },

    // Switch to a different machine. Tears down the WS subscription,
    // clears in-flight job state (so the previous machine's running
    // counters don't show against the new machine), reseeds metrics,
    // and reconnects the WS to the new machine_id.
    async setCurrentMachine(id: string) {
      if (!id || id === this.currentMachineId) return;
      this.currentMachineId = id;
      try {
        localStorage.setItem(CURRENT_MACHINE_KEY, id);
      } catch {
        /* ignore */
      }
      // Reset transient per-machine state. Activity log is shared
      // across machines for now (operators seldom switch mid-job; if
      // that turns out to be wrong we can move log persistence to a
      // per-machine key later).
      this.running = false;
      this.filePath = "";
      this.fileURL = "";
      this.method = "";
      this.lineCurrent = 0;
      this.lineTotal = 0;
      this.haasOk = true;
      this.haasLastError = "";
      this.recoveryPending = false;
      this.recoveryFilePath = "";
      this.raw = null;
      this.metrics = {};
      this.metricsSeeded = false;
      this.queue = [];
      this.queueLoaded = false;
      // Drop the existing WS and reconnect with the new machine_id.
      // pollOnce + seedMetrics fire fresh state for the new machine
      // so the dashboard repaints without waiting for the WS to
      // reattach.
      if (ws) {
        try {
          ws.close();
        } catch {
          /* ignore */
        }
        ws = null;
      }
      if (wsRetry) {
        clearTimeout(wsRetry);
        wsRetry = null;
      }
      await Promise.all([this.pollOnce(), this.seedMetrics(), this.loadQueue()]);
      if (started) this.connectWS();
    },

    async ackRecovery() {
      await cncApi.ackRecovery(this.currentMachineId || undefined);
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
        this.metrics = await cncApi.getState(
          this.currentMachineId || undefined
        );
        this.metricsSeeded = true;
      } catch {
        /* leave previous map; WS events will fill it in */
      }
    },

    applyMetric(m: CncMetric) {
      this.metrics = { ...this.metrics, [m.key]: m };
      this.metricsSeeded = true;
    },

    // ── Queue ────────────────────────────────────────────────────────
    async loadQueue() {
      try {
        this.queue = await cncApi.listQueue(
          this.currentMachineId || undefined
        );
        this.queueLoaded = true;
      } catch {
        /* leave existing queue; WS event will re-seed */
      }
    },

    async addToQueue(filePath: string) {
      const item = await cncApi.addToQueue(
        filePath,
        this.currentMachineId || undefined
      );
      // Optimistically append — the WS "queue" event will replace
      // with the canonical server-side ordering moments later.
      this.queue = [...this.queue, item];
      return item;
    },

    async removeFromQueue(id: string) {
      await cncApi.removeFromQueue(id, this.currentMachineId || undefined);
      this.queue = this.queue.filter((q) => q.id !== id);
    },

    async reorderQueue(ids: string[]) {
      const next = await cncApi.reorderQueue(
        ids,
        this.currentMachineId || undefined
      );
      this.queue = next;
    },

    async sendFromQueue(item: QueueItem, method: SendMethod) {
      await cncApi.start(
        item.file_path,
        method,
        this.currentMachineId || undefined,
        item.id
      );
    },

    // autoSendFromQueue runs /api/cnc/auto-send. On a clean preflight
    // the streamer starts immediately and the WS picks it up; on a
    // block the response carries blocked_reason + the preflight payload
    // so the caller can fall back to the wizard with both already in
    // hand. Throws on transport / server error.
    async autoSendFromQueue(item: QueueItem, method: SendMethod) {
      return cncApi.autoSend(
        item.file_path,
        method,
        this.currentMachineId || undefined,
        item.id
      );
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
        const s = await cncApi.getStatus(this.currentMachineId || undefined);
        this.applyStatus(s);
      } catch {
        // Silent: if /api/cnc/status fails (typically because the user
        // is on a public share view without auth) we just leave the
        // pill hidden. Don't spam the toast layer for a passive query.
      }
    },

    // Start the poll loop + try to upgrade to WS. Idempotent — safe to
    // call from every layout that mounts. loadMachines() runs first
    // so the persisted currentMachineId is reconciled against the
    // current registry (e.g. a machine deleted in settings since the
    // last session) before the first poll fires; otherwise we'd 404
    // every call until the operator manually picked a new machine.
    async start() {
      if (started) return;
      started = true;
      await this.loadMachines();
      this.pollOnce();
      this.loadQueue();
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
      // The stream is per-machine; setCurrentMachine() tears down +
      // reconnects so events are always for the active selection.
      const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
      const q = this.currentMachineId
        ? `?machine_id=${encodeURIComponent(this.currentMachineId)}`
        : "";
      const url = `${proto}//${window.location.host}${baseURL}/api/cnc/stream${q}`;
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
              const tag = this.method ? ` [${this.method}]` : "";
              this.pushLog("info", `running${tag}: ${this.filePath}`);
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
          } else if (ev.type === "queue" && Array.isArray(ev.queue)) {
            this.queue = ev.queue as QueueItem[];
            this.queueLoaded = true;
          } else if (ev.type === "dprnt" && typeof ev.text === "string") {
            // DPRNT macro output during a stream. Surface in the
            // activity log as an info entry prefixed so operators can
            // filter — Haas DPRNT is a niche feature used for in-cycle
            // probing output and homemade telemetry.
            this.pushLog("info", `DPRNT: ${ev.text}`);
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
