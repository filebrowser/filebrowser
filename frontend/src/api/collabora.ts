import { fetchJSON } from "./utils";

export interface CollaboraOpenResponse {
  url: string;
  fileID: string;
  canWrite: boolean;
  name: string;
}

const officeExtensions = new Set([
  ".doc",
  ".docx",
  ".docm",
  ".dot",
  ".dotx",
  ".odt",
  ".ott",
  ".rtf",
  ".xls",
  ".xlsx",
  ".xlsm",
  ".xlt",
  ".xltx",
  ".ods",
  ".ots",
  ".csv",
  ".ppt",
  ".pptx",
  ".pptm",
  ".pot",
  ".potx",
  ".odp",
  ".otp",
  ".vsd",
  ".vsdx",
  ".odg",
  ".pdf",
]);

export function isSupportedExtension(extension?: string): boolean {
  if (!extension) return false;
  return officeExtensions.has(extension.toLowerCase());
}

export async function openPath(path: string): Promise<CollaboraOpenResponse> {
  if (!path.startsWith("/")) {
    path = `/${path}`;
  }

  const params = new URLSearchParams({ path });
  return fetchJSON<CollaboraOpenResponse>(`/api/collabora/open?${params.toString()}`, {
    method: "GET",
  });
}
