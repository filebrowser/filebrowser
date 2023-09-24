interface Uploads {
  [key: number]: Upload;
}

interface Upload {
  id: number;
  file: UploadEntry;
  type?: ResourceType;
}

interface UploadItem {
  id: number;
  url?: string;
  path: string;
  file: UploadEntry;
  dir?: boolean;
  overwrite?: boolean;
  type?: ResourceType;
}

interface UploadEntry {
  name: string;
  size: number;
  isDir: boolean;
  fullPath?: string;
  file?: File;
}

type UploadList = UploadEntry[];

type Progress = number | boolean;
