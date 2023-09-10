export type IToastSuccess = (message: string) => void;
export type IToastError = (
  error: Error | string,
  displayReport?: boolean
) => void;
