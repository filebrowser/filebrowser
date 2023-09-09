type ApiUrl = string; // Can also be set as a path eg: "path1" | "path2"

type resourcePath = string;

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

interface tusSettings {
  retryCount: number;
  chunkSize: number;
}

type algo = any;

type inline = any;

interface share {
  expire: any;
  hash: string;
  path: string;
  userID: number;
  token: string;
}

interface settings {
  any;
}

type searchParams = any;
