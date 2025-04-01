import { fetchURL, removePrefix } from "@/api/utils.ts";

export async function fetchUrlFile(
  savePath: string,
  saveName: string,
  fetchUrl: string,
  opts: ApiOpts
) {
  const res = await fetchURL(`/api/download/`, {
    method: "POST",
    body: JSON.stringify({
      url: fetchUrl,
      pathname: removePrefix(savePath),
      filename: saveName,
    }),
    ...opts,
  });
  const taskID = await res.text();
  console.log("on create download task: ", taskID);
  return taskID;
}

export type DownloadTask = {
  filename: string;
  pathname: string;
  progress: number;
  savedSize: number;
  taskID: string;
  totalSize: number;
  url: string;
  status: string;
  error: string;
};

export async function queryDownloadTask(taskID: string): Promise<DownloadTask> {
  const res = await fetchURL(`/api/download/${taskID}`, {});
  return await res.json();
}
