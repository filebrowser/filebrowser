export interface IFile {
  index?: number;
  name: string;
  modified: string;
  path: string;
  subtitles: any[];
  isDir: boolean;
  size: number;
  fullPath: string;
  type: FileType;
  items: IFile[];
  token?: string;
  hash: string;
  url?: string;
}

export type FileType =
  | "video"
  | "audio"
  | "image"
  | "pdf"
  | "text"
  | "blob"
  | "textImmutable";

type req = {
  path: string;
  name: string;
  size: number;
  extension: string;
  modified: string;
  mode: number;
  isDir: boolean;
  isSymlink: boolean;
  type: string;
  url: string;
  hash: string;
};

export interface Uploads {
  [key: string]: Upload;
}

export interface Upload {
  id: number;
  file: IFile;
  type: string;
}

export interface Item {
  id: number;
  url?: string;
  path: string;
  file: IFile;
  dir?: boolean;
  overwrite?: boolean;
  type?: FileType;
}
