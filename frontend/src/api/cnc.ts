// /api/cnc/* — Haas Dashboard ↔ Zinc integration.
// See docs/INTEGRATION_WITH_HAAS_DASHBOARD.md.

import { fetchURL, fetchJSON } from "./utils";

export interface CncSettings {
  haasHost: string;
  haasPort: number;
  cameraUrl: string;
  machineToken?: string;
}

export interface CncStatus {
  running: boolean;
  file_path?: string;
  file_url?: string;
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

export function getStatus() {
  return fetchJSON<CncStatus>(`/api/cnc/status`, {});
}

export function start(filePath: string) {
  return fetchJSON<{ job_id: string }>(`/api/cnc/start`, {
    method: "POST",
    body: JSON.stringify({ file_path: filePath }),
  });
}

export function stop() {
  return fetchJSON<{ stopped: boolean }>(`/api/cnc/stop`, {
    method: "POST",
  });
}

export function ackRecovery() {
  return fetchJSON<{ acknowledged: boolean }>(`/api/cnc/recovery/ack`, {
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

export function checkConnection() {
  return fetchJSON<CncCheckResult>(`/api/cnc/check`, { method: "POST" });
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

export function getState() {
  return fetchJSON<CncStateSnapshot>(`/api/cnc/state`, {});
}
