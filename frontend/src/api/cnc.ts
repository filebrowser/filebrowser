// /api/cnc/* — Haas Dashboard ↔ Zinc integration.
// See docs/INTEGRATION_WITH_HAAS_DASHBOARD.md.

import { fetchURL, fetchJSON } from "./utils";

export interface CncSettings {
  haasHost: string;
  haasPort: number;
  cameraUrl: string;
  haasDashboardUrl: string;
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
