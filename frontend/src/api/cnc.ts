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
  machineId?: string
) {
  const body: Record<string, string> = { file_path: filePath, method };
  if (machineId) body.machine_id = machineId;
  return fetchJSON<{ job_id: string }>(`/api/cnc/start`, {
    method: "POST",
    body: JSON.stringify(body),
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

// ── Pre-flight tool check (NC ↔ tool table) ─────────────────────────────────

export type PreflightStatus = "ok" | "warn" | "empty" | "offline" | "missing";

export interface PreflightToolUsage {
  tool: number;
  reference_count: number;
  comment?: string;
  expected_diameter?: number;
  expected_corner_radius?: number;
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
