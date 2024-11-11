import { fetchJSON } from "./utils";

export async function getQuota() {
  return await fetchJSON<IQuota>(`/api/quota`, {
    method: "GET",
  });
}
