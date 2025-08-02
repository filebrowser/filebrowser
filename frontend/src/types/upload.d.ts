type Upload = {
  path: string;
  name: string;
  file: File | null;
  type: ResourceType;
  overwrite: boolean;
  totalBytes: number;
  sentBytes: number;
};

interface UploadEntry {
  name: string;
  size: number;
  isDir: boolean;
  fullPath?: string;
  file?: File;
}

type UploadList = UploadEntry[];
