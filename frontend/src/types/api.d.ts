type ApiMethod = "GET" | "POST" | "PUT" | "DELETE" | "PATCH";

type ApiContent =
  | Blob
  | File
  | Pick<ReadableStreamDefaultReader<any>, "read">
  | "";

interface ApiOpts {
  method?: ApiMethod;
  headers?: object;
  body?: any;
}

interface TusSettings {
  retryCount: number;
  chunkSize: number;
}

type ChecksumAlgs = "md5" | "sha1" | "sha256" | "sha512";

type inline = any;

interface Share {
  expire: any;
  hash: string;
  path: string;
  userID: number;
  token: string;
}

interface settings {
  any;
}

interface SearchParams {
  [key: string]: string;
}
