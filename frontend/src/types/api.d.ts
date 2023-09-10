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

export interface Share {
  expire: any;
  hash: string;
  path: string;
  userID: number;
  token: string;
}

interface settings {
  any;
}

export interface SearchParams {
  [key: string]: string;
}
