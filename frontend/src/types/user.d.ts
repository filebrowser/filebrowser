<<<<<<< HEAD
type UserKey = keyof IUser;

interface IUser {
=======
export interface IUser {
>>>>>>> kloon15/vue3
  id: number;
  username: string;
  password: string;
  scope: string;
  locale: string;
<<<<<<< HEAD
  lockPassword: boolean;
  viewMode: string;
  singleClick: boolean;
  perm: UserPerm;
  commands: any[];
  sorting: UserSorting;
  rules: any[];
  hideDotfiles: boolean;
  dateFormat: boolean;
}

interface UserPerm {
  admin: boolean;
  execute: boolean;
  create: boolean;
  rename: boolean;
  modify: boolean;
  delete: boolean;
  share: boolean;
  download: boolean;
}

interface UserSorting {
  by: string;
  asc: boolean;
}
=======
  perm: Permissions;
  commands: string[];
  rules: IRule[];
  lockPassword: boolean;
  hideDotfiles: boolean;
  singleClick: boolean;
  dateFormat: boolean;
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

export interface UserSorting {
  by: string;
  asc: boolean;
}

export interface IRule {
  allow: boolean;
  path: string;
  regex: boolean;
  regexp: IRegexp;
}

interface IRegexp {
  raw: string;
}
>>>>>>> kloon15/vue3
