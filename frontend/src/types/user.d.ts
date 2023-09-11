type UserKey = keyof IUser;

interface IUser {
  id: number;
  username: string;
  password: string;
  scope: string;
  locale: string;
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