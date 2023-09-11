interface IUser {
  id: number;
  username: string;
  password: string;
  scope: string;
  locale: string;
  perm: Permissions;
  commands: string[];
  rules: IRule[];
  lockPassword: boolean;
  hideDotfiles: boolean;
  singleClick: boolean;
  dateFormat: boolean;
}

interface Permissions {
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

interface UserSorting {
  by: string;
  asc: boolean;
}

interface IRule {
  allow: boolean;
  path: string;
  regex: boolean;
  regexp: IRegexp;
}

interface IRegexp {
  raw: string;
}