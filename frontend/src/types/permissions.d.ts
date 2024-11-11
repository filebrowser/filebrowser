interface FilePermissions {
  owner: PermissionModes;
  group: PermissionModes;
  others: PermissionModes;
}

interface PermissionModes {
  read: boolean;
  write: boolean;
  execute: boolean;
}
