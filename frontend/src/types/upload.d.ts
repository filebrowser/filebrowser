type Upload = {
  path: string;
  name: string;
  file: File | null;
  type: ResourceType;
  overwrite: boolean;
  skip: boolean;
  totalBytes: number;
  sentBytes: number;
  rawProgress: {
    sentBytes: number;
  };
};

interface UploadEntry {
  name: string;
  size: number;
  isDir: boolean;
  fullPath?: string;
  file?: File;
  overwrite?: boolean;
  skip?: boolean;
}

type UploadList = UploadEntry[];
