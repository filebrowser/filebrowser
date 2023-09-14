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
  fullPath: string;
  isDir: boolean;
  name: string;
  size: number;
  file?: File;
}

type UploadList = UploadEntry[];

type Progress = number | boolean;
