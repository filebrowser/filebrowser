export interface User {
  id: number;
  locale: string;
  perm: Permissions;
}

export interface Permissions {
  admin: boolean;
  copy: boolean;
  create: boolean;
  delete: boolean;
  download: boolean;
  execute: boolean;
  modify: boolean;
  move: boolean;
  rename: boolean;
  share: boolean;
  shell: boolean;
  upload: boolean;
}
