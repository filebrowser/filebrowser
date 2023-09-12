interface Uploads {
  [key: string]: Upload;
}

interface Upload {
  id: number;
  file: Resource;
  type: string;
}

interface UploadItem {
  id: number;
  url?: string;
  path: string;
  file: Resource;
  dir?: boolean;
  overwrite?: boolean;
  type?: ResourceType;
}
