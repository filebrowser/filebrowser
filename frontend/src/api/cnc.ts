// /api/cnc/* — Haas Dashboard ↔ Zinc integration.
// See docs/INTEGRATION_WITH_HAAS_DASHBOARD.md.

import { fetchURL, fetchJSON } from "./utils";

export type CameraType = "auto" | "hls" | "mjpeg" | "iframe" | "none";

export interface CncMachine {
  id: string;
  name: string;
  // Controller family. "haas" is the only brand wired into the streamer
  // today; the field exists so per-brand send protocols can land later
  // without a settings migration.
  brand?: string;
  host: string;
  port: number;
  // Magazine slot count for tool-table reads. 0 / undefined = use the
  // backend default (30). Set to your actual pocket count so reads
  // don't probe unreachable upper slots.
  toolSlots?: number;
  cameraUrl?: string;
  cameraType?: CameraType;
  // When true the server refuses /api/cnc/start if any program tool
  // is missing/empty in the latest tool table. Off by default; the
  // wizard always soft-warns regardless.
  requirePreflight?: boolean;
  // X/Y/Z always render; A/B/C are optional 4th/5th-axis controllers.
  // Stored as uppercase letters; undefined / empty falls back to XYZ.
  axesEnabled?: string[];
  // Inches threshold for the ∆ CMD column on the dashboard. Falls back
  // to 0.001" when 0 / unset.
  positionToleranceIn?: number;
  // When true, the streamer scavenges DPRNT[…] macro output between
  // line writes during a job and emits it as WS "dprnt" events.
  // Off by default; enable on machines that use DPRNT for in-cycle
  // probing or telemetry output.
  dprntCapture?: boolean;
  // When true, /api/cnc/auto-send is enabled for this machine. The
  // pipeline runs preflight and starts the send in one round-trip
  // when everything checks out. Operators still press CYCLE START
  // — auto-send does NOT trigger the cycle.
  autoSendEnabled?: boolean;
}

export interface CncSettings {
  machines: CncMachine[];
  machineToken?: string;
  // Legacy mirror of machines[0] — server returns these for back-compat
  // with any old client. New code reads `machines` directly.
  haasHost?: string;
  haasPort?: number;
  cameraUrl?: string;
}

export interface CncMachinesList {
  machines: CncMachine[];
  default_id: string;
}

export function listMachines() {
  return fetchJSON<CncMachinesList>(`/api/cnc/machines`, {});
}

export type SendMethod = "mem" | "dnc";

export interface CncStatus {
  running: boolean;
  file_path?: string;
  file_url?: string;
  method?: SendMethod;
  line_current?: number;
  line_total?: number;
  started_at?: string;
  haas_ok: boolean;
  haas_last_error?: string;
  recovery_pending?: boolean;
  recovery_file_path?: string;
  // Attachment surfaces when no real job is running but the operator
  // (or future O-number auto-match) has marked a file as the active
  // program. The dashboard uses attached_file for follow-along
  // alongside the live line_current metric, with a UI badge so the
  // operator knows it might not actually be the file on the
  // controller.
  attached_file?: string;
  attached_source?: "manual" | "auto";
  attached_at?: string;
}

export function getSettings() {
  return fetchJSON<CncSettings>(`/api/cnc/settings`, {});
}

export async function updateSettings(settings: CncSettings) {
  await fetchURL(`/api/cnc/settings`, {
    method: "PUT",
    body: JSON.stringify(settings),
  });
}

export function regenerateToken() {
  return fetchJSON<{ machineToken: string }>(`/api/cnc/settings/token`, {
    method: "POST",
  });
}

export function getStatus(machineId?: string) {
  const q = machineId ? `?machine_id=${encodeURIComponent(machineId)}` : "";
  return fetchJSON<CncStatus>(`/api/cnc/status${q}`, {});
}

export function start(
  filePath: string,
  method: SendMethod = "mem",
  machineId?: string,
  queueId?: string
) {
  const body: Record<string, string> = { file_path: filePath, method };
  if (machineId) body.machine_id = machineId;
  if (queueId) body.queue_id = queueId;
  return fetchJSON<{ job_id: string }>(`/api/cnc/start`, {
    method: "POST",
    body: JSON.stringify(body),
  });
}

// AutoSendResponse — see http/cnc_autosend.go. `started` is true when
// the pipeline accepted the send and a streamer job is now running;
// false when blocked (the UI should fall back to the manual wizard
// and the `preflight` payload is the same shape /api/cnc/preflight
// returns).
export interface AutoSendResponse {
  started: boolean;
  job_id?: string;
  blocked_reason?: string;
  preflight?: Preflight;
}

export function autoSend(
  filePath: string,
  method: SendMethod = "mem",
  machineId?: string,
  queueId?: string
) {
  const body: Record<string, string> = { file_path: filePath, method };
  if (machineId) body.machine_id = machineId;
  if (queueId) body.queue_id = queueId;
  return fetchJSON<AutoSendResponse>(`/api/cnc/auto-send`, {
    method: "POST",
    body: JSON.stringify(body),
  });
}

// Mark a filebrowser file as the program the controller is actually
// running, without the streamer pushing it. Useful when the operator
// loaded the program from SD card / Ethernet drop and wants the
// dashboard to follow along anyway.
export function attachFile(
  filePath: string,
  machineId?: string,
  source: "manual" | "auto" = "manual"
) {
  const body: Record<string, string> = { file_path: filePath, source };
  if (machineId) body.machine_id = machineId;
  return fetchJSON<CncStatus>(`/api/cnc/attach`, {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function detachFile(machineId?: string) {
  const q = machineId ? `?machine_id=${encodeURIComponent(machineId)}` : "";
  return fetchJSON<CncStatus>(`/api/cnc/attach${q}`, {
    method: "DELETE",
  });
}

// ── NC chapters (operation-header comment TOC) ────────────────────────────

export interface Chapter {
  line: number;
  comment: string;
}

export interface ChapterList {
  file_path: string;
  total: number;
  chapters: Chapter[];
}

export function getChapters(filePath: string) {
  const params = new URLSearchParams({ file_path: filePath });
  return fetchJSON<ChapterList>(`/api/cnc/chapters?${params}`, {});
}

// ── Queue (per-machine staging area) ──────────────────────────────────────

export type QueueState = "queued" | "sending" | "running";

export interface QueueItem {
  id: string;
  file_path: string;
  job_name?: string;
  onumber_hint?: string;
  size_bytes?: number;
  state: QueueState;
  method?: SendMethod;
  added_at: string;
  line_current?: number;
  line_total?: number;
}

export function listQueue(machineId?: string) {
  const q = machineId ? `?machine_id=${encodeURIComponent(machineId)}` : "";
  return fetchJSON<QueueItem[]>(`/api/cnc/queue${q}`, {});
}

export function addToQueue(filePath: string, machineId?: string) {
  const body: Record<string, string> = { file_path: filePath };
  if (machineId) body.machine_id = machineId;
  return fetchJSON<QueueItem>(`/api/cnc/queue`, {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function removeFromQueue(id: string, machineId?: string) {
  const q = machineId ? `?machine_id=${encodeURIComponent(machineId)}` : "";
  return fetchJSON<{ removed: boolean }>(`/api/cnc/queue/${id}${q}`, {
    method: "DELETE",
  });
}

export function reorderQueue(ids: string[], machineId?: string) {
  const q = machineId ? `?machine_id=${encodeURIComponent(machineId)}` : "";
  return fetchJSON<QueueItem[]>(`/api/cnc/queue${q}`, {
    method: "PATCH",
    body: JSON.stringify({ ids }),
  });
}

export function stop(machineId?: string) {
  const q = machineId ? `?machine_id=${encodeURIComponent(machineId)}` : "";
  return fetchJSON<{ stopped: boolean }>(`/api/cnc/stop${q}`, {
    method: "POST",
  });
}

export function ackRecovery(machineId?: string) {
  const q = machineId ? `?machine_id=${encodeURIComponent(machineId)}` : "";
  return fetchJSON<{ acknowledged: boolean }>(`/api/cnc/recovery/ack${q}`, {
    method: "POST",
  });
}

export interface CncCheckResult {
  bridge: {
    ok: boolean;
    latency_ms?: number;
    error?: string;
    address?: string;
  };
  controller: {
    ok: boolean;
    latency_ms?: number;
    error?: string;
    mode?: string;
  };
}

export function checkConnection(machineId?: string) {
  const q = machineId ? `?machine_id=${encodeURIComponent(machineId)}` : "";
  return fetchJSON<CncCheckResult>(`/api/cnc/check${q}`, { method: "POST" });
}

export interface CncSiblings {
  model_url?: string;
  model_name?: string;
  model_path?: string;
  drawing_url?: string;
  drawing_name?: string;
  drawing_path?: string;
}

export function getSiblings(filePath: string) {
  const q = encodeURIComponent(filePath);
  return fetchJSON<CncSiblings>(`/api/cnc/siblings?path=${q}`, {});
}

export interface ProbeToolsBaseResult {
  base: number;
  label: string;
  ok: number;
  empty: number;
  errors: number;
  first_error?: string;
  samples: { slot: number; var: number; value?: string; error?: string }[];
}

export interface ProbeToolsReport {
  slots_probed: number;
  duration_ms: number;
  bridge_address: string;
  bases: ProbeToolsBaseResult[];
  verdict: string;
  recommendation: string;
}

export function probeTools(slots = 30, machineId?: string) {
  const q = machineId ? `&machine_id=${encodeURIComponent(machineId)}` : "";
  return fetchJSON<ProbeToolsReport>(`/api/cnc/probe-tools?slots=${slots}${q}`, {
    method: "POST",
  });
}

// ── Tool-life discovery probe (POST /api/cnc/probe-tool-life) ─────────────

export interface ToolLifeProbeSample {
  macro: number;
  value?: string;
  number?: number;
  error?: string;
}

export interface MacroCluster {
  start: number;
  end: number;
  count: number;
}

export interface ToolLifeProbeReport {
  start: number;
  end: number;
  step: number;
  probed: number;
  ok: number;
  empty: number;
  non_zero: number;
  errors: number;
  duration_ms: number;
  bridge_address: string;
  samples: ToolLifeProbeSample[];
  // Contiguous runs of non-zero macros (gap ≤ step). Present when the
  // probe found any populated macro; omitted otherwise.
  clusters?: MacroCluster[];
  verdict: string;
  recommendation: string;
}

export function probeToolLife(opts: {
  start?: number;
  end?: number;
  step?: number;
  machineId?: string;
}) {
  const params = new URLSearchParams();
  if (opts.start) params.set("start", String(opts.start));
  if (opts.end) params.set("end", String(opts.end));
  if (opts.step) params.set("step", String(opts.step));
  if (opts.machineId) params.set("machine_id", opts.machineId);
  const qs = params.toString();
  return fetchJSON<ToolLifeProbeReport>(
    `/api/cnc/probe-tool-life${qs ? `?${qs}` : ""}`,
    { method: "POST" }
  );
}

// ── Tool table (live readout, persisted as JSON in user share) ─────────────

export interface ToolTableSlot {
  slot: number;
  length_geom?: number;
  length_wear?: number;
  diameter_geom?: number;
  diameter_wear?: number;
  effective_length?: number;
  effective_diameter?: number;
  empty?: boolean;
  errors?: Record<string, string>;
}

export interface ToolTable {
  machine_id: string;
  machine_name?: string;
  bridge_address: string;
  read_at: string;
  duration_ms: number;
  slots_requested: number;
  slots_read: number;
  slots: ToolTableSlot[];
}

export interface ToolTableEnvelope {
  table: ToolTable;
  persist_error?: string;
  // Set when the read was cut short (timeout / cancel). Partial table
  // is still returned + persisted; surface the reason in the panel.
  read_error?: string;
}

export interface ToolTableHistoryEntry {
  path: string;
  filename: string;
  modified_at: string;
  size_bytes: number;
  slots_requested?: number;
  slots_read?: number;
}

export interface ToolTableHistory {
  machine_id: string;
  folder: string;
  entries: ToolTableHistoryEntry[];
}

export function readToolTable(slots = 30, machineId?: string) {
  const params = new URLSearchParams({ slots: String(slots) });
  if (machineId) params.set("machine_id", machineId);
  return fetchJSON<ToolTableEnvelope>(`/api/cnc/tool-table?${params}`, {
    method: "POST",
  });
}

// 204 No Content means "no dump persisted yet" — valid empty state,
// not an error. fetchJSON would treat it as a non-200 throw so we hit
// the lower-level fetchURL.
export async function getLatestToolTable(
  machineId?: string
): Promise<ToolTable | null> {
  const q = machineId ? `?machine_id=${encodeURIComponent(machineId)}` : "";
  const res = await fetchURL(`/api/cnc/tool-table${q}`, {});
  if (res.status === 204) return null;
  const env = (await res.json()) as ToolTableEnvelope;
  return env.table ?? null;
}

export function getToolTableHistory(machineId?: string) {
  const q = machineId ? `?machine_id=${encodeURIComponent(machineId)}` : "";
  return fetchJSON<ToolTableHistory>(`/api/cnc/tool-table/history${q}`, {});
}

// ── Tool-table diff (history comparison) ─────────────────────────────────

export type SlotChange =
  | "unchanged"
  | "added"
  | "removed"
  | "drift_diameter"
  | "drift_length"
  | "drift_both"
  | "offline_then"
  | "offline_now";

export interface SlotDiff {
  slot: number;
  change: SlotChange;
  old_diameter?: number;
  new_diameter?: number;
  diameter_delta?: number;
  old_length?: number;
  new_length?: number;
  length_delta?: number;
  note?: string;
}

export interface DiffSummary {
  added: number;
  removed: number;
  drift_diameter: number;
  drift_length: number;
  drift_both: number;
  offline_then: number;
  offline_now: number;
  unchanged: number;
}

export interface ToolTableDiff {
  machine_id: string;
  old_read_at: string;
  new_read_at: string;
  diameter_tolerance: number;
  length_tolerance: number;
  summary: DiffSummary;
  slots: SlotDiff[];
}

export function diffToolTables(opts: {
  machineId?: string;
  oldFile?: string;
  newFile?: string;
  diaTol?: number;
  lenTol?: number;
}) {
  const params = new URLSearchParams();
  if (opts.machineId) params.set("machine_id", opts.machineId);
  if (opts.oldFile) params.set("old", opts.oldFile);
  if (opts.newFile) params.set("new", opts.newFile);
  if (opts.diaTol !== undefined) params.set("dia_tol", String(opts.diaTol));
  if (opts.lenTol !== undefined) params.set("len_tol", String(opts.lenTol));
  const qs = params.toString();
  return fetchJSON<ToolTableDiff>(
    `/api/cnc/tool-table/diff${qs ? `?${qs}` : ""}`,
    {}
  );
}

// ── Pre-flight tool check (NC ↔ tool table) ─────────────────────────────────

export type PreflightStatus = "ok" | "warn" | "empty" | "offline" | "missing";

export interface PreflightToolUsage {
  tool: number;
  reference_count: number;
  comment?: string;
  expected_diameter?: number;
  expected_corner_radius?: number;
  // Set when the program activates G41/G42 while this tool is the
  // most-recently-selected T-code. Surfaces a "uses G41" pill in the
  // wizard tool row; also flips the status to warn when the diameter
  // offset is 0 since cutter comp + D=0 produces silent wrong cuts.
  uses_cutter_comp?: boolean;
  in_table: boolean;
  loaded: boolean;
  empty_pocket: boolean;
  offline: boolean;
  actual_diameter?: number;
  diameter_delta?: number;
  status: PreflightStatus;
  status_reason?: string;
}

export interface PreflightSummary {
  ok: number;
  warn: number;
  empty: number;
  offline: number;
  missing: number;
}

export interface Preflight {
  file_path: string;
  machine_id: string;
  tools: PreflightToolUsage[];
  table_read_at?: string;
  table_missing?: boolean;
  summary: PreflightSummary;
  // Program-order first T-code. Surfaced so the wizard can show
  // "starting tool: T5" before Send. Absent when the program has no
  // T-codes (rare — typically a sub-program fragment).
  starting_tool?: number;
  // Controller's currently-selected tool from the aggregator's Q201
  // metric at the time the preflight ran. Absent when the metric is
  // stale or unavailable.
  current_spindle_tool?: number;
  // True when starting_tool and current_spindle_tool are both present
  // and differ. Pre-computed so the wizard doesn't re-derive.
  spindle_swap?: boolean;
}

export function getPreflight(filePath: string, machineId?: string) {
  const params = new URLSearchParams({ file_path: filePath });
  if (machineId) params.set("machine_id", machineId);
  return fetchJSON<Preflight>(`/api/cnc/preflight?${params}`, {});
}

export interface CncMetric {
  key: string;
  label: string;
  q_code: number;
  macro_var?: number;
  interval_s: number;
  raw?: string;
  value?: string;
  parsed?: string | number | boolean | Record<string, unknown> | null;
  last_update?: string;
  last_error?: string;
  ok: boolean;
  stale: boolean;
}

export type CncStateSnapshot = Record<string, CncMetric>;

export function getState(machineId?: string) {
  const q = machineId ? `?machine_id=${encodeURIComponent(machineId)}` : "";
  return fetchJSON<CncStateSnapshot>(`/api/cnc/state${q}`, {});
}

// ── Haas alarm/setting/parameter code catalog ───────────────────────────────

import type { CodeKind, CodeEntry } from "@/utils/codeRefs";
export type { CodeKind, CodeEntry } from "@/utils/codeRefs";

export interface CodeLookupResponse {
  ok: boolean;
  kind: CodeKind;
  entry: CodeEntry;
}

export function lookupCode(kind: CodeKind, number: number) {
  return fetchJSON<CodeLookupResponse>(
    `/api/cnc/codes/lookup?kind=${encodeURIComponent(kind)}&number=${number}`,
    {}
  );
}

export function searchCodes(q: string, kind?: CodeKind, limit?: number) {
  const params = new URLSearchParams();
  if (q) params.set("q", q);
  if (kind) params.set("kind", kind);
  if (limit) params.set("limit", String(limit));
  return fetchJSON<{ count: number; results: CodeEntry[] }>(
    `/api/cnc/codes/search?${params.toString()}`,
    {}
  );
}
