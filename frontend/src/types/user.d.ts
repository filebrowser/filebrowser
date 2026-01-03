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
  viewMode: ViewModeType;
  sorting?: Sorting;
  aceEditorTheme: string;
  quotaLimit: number;
  quotaUnit: string;
  enforceQuota: boolean;
  quotaUsed: number;
}

type ViewModeType = "list" | "mosaic" | "mosaic gallery";

interface IUserForm {
  id?: number;
  username?: string;
  password?: string;
  scope?: string;
  locale?: string;
  perm?: Permissions;
  commands?: string[];
  rules?: IRule[];
  lockPassword?: boolean;
  hideDotfiles?: boolean;
  singleClick?: boolean;
  dateFormat?: boolean;
  quotaLimit?: number;
  quotaUnit?: string;
  enforceQuota?: boolean;
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

interface Sorting {
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

interface IQuotaInfo {
  limit: number;
  used: number;
  unit: string;
  enforce: boolean;
  percentage: number;
}

type UserTheme = "light" | "dark" | "";
