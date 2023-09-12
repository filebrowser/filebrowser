interface ResourceBase {
  path: string;
  name: string;
  size: number;
  extension: string;
  modified: string; // ISO 8601 datetime
  mode: number;
  isDir: boolean;
  isSymlink: boolean;
  type: ResourceType;
  url: string;
  fullPath?: string;
}

interface Resource extends ResourceBase {
  items: ResourceItem[];
  numDirs: number;
  numFiles: number;
  sorting: Sorting;
  hash?: string;
  token?: string;
}

interface ResourceItem extends ResourceBase {
  index?: number;
  subtitles?: string[];
}

type ResourceType =
  | "video"
  | "audio"
  | "image"
  | "pdf"
  | "text"
  | "blob"
  | "textImmutable";
