import { fetchURL, fetchJSON } from "./utils";

export function get() {
  return fetchJSON<ISettings>(`/api/settings`, {});
}

export async function update(settings: ISettings) {
  await fetchURL(`/api/settings`, {
    method: "PUT",
    body: JSON.stringify(settings),
  });
}


export async function testClamAV(config: SettingsClamAV) {
  return fetchJSON<ClamAVTestResponse>(`/api/settings/clamav/test`, {
    method: "POST",
    body: JSON.stringify(config),
  });
}
