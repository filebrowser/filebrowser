export type ApiUrl = string; // Can also be set as a path eg: "path1" | "path2"

export type ApiMethod = "GET" | "POST" | "PUT" | "DELETE" | "PATCH";

export type ApiContent =
  | Blob
  | File
  | Pick<ReadableStreamDefaultReader<any>, "read">
  | "";

export interface ApiOpts {
  method?: ApiMethod;
  headers?: object;
  body?: any;
}

export interface TusSettings {
  retryCount: number;
  chunkSize: number;
}

export type ChecksumAlgs = "md5" | "sha1" | "sha256" | "sha512";

type inline = any;

<<<<<<< HEAD
interface IShare {
  expire: any;
=======
export interface IShare {
>>>>>>> kloon15/vue3
  hash: string;
  path: string;
  expire?: any;
  userID?: number;
  token?: string;
}

interface settings {
  any;
}

export interface SearchParams {
  [key: string]: string;
}
