import { fetchURL, fetchJSON } from "./utils";

export interface CollaboraTestCheck {
  name: string;
  status: "success" | "warning" | "error";
  message: string;
}

export interface CollaboraTestResponse {
  ok: boolean;
  checks: CollaboraTestCheck[];
}

export interface ClamAVTestResponse {
  status: string;
  message: string;
}

export function get() {
  return fetchJSON<ISettings>(`/api/settings`, {});
}

export async function update(settings: ISettings) {
  await fetchURL(`/api/settings`, {
    method: "PUT",
    body: JSON.stringify(settings),
  });
}

export async function testCollabora(collabora: SettingsCollabora) {
  return fetchJSON<CollaboraTestResponse>(`/api/collabora/test`, {
    method: "POST",
    body: JSON.stringify({ collabora }),
  });
}


export async function testClamAV(config: SettingsClamAV) {
  return fetchJSON<ClamAVTestResponse>(`/api/settings/clamav/test`, {
    method: "POST",
    body: JSON.stringify(config),
  });
}
