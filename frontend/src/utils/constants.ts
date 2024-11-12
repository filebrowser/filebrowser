const name: string = window.FileBrowser.Name || "File Browser";
const disableExternal: boolean = window.FileBrowser.DisableExternal;
const disableUsedPercentage: boolean = window.FileBrowser.DisableUsedPercentage;
const baseURL: string = window.FileBrowser.BaseURL;
const staticURL: string = window.FileBrowser.StaticURL;
const recaptcha: string = window.FileBrowser.ReCaptcha;
const recaptchaKey: string = window.FileBrowser.ReCaptchaKey;
const signup: boolean = window.FileBrowser.Signup;
const version: string = window.FileBrowser.Version;
const logoURL = `${staticURL}/img/logo.svg`;
const tmpDir: string = window.FileBrowser.TmpDir;
const trashDir: string = window.FileBrowser.TrashDir;
const quotaExists: boolean = window.FileBrowser.QuotaExists;
const noAuth: boolean = window.FileBrowser.NoAuth;
const authMethod = window.FileBrowser.AuthMethod;
const authLogoutURL: string = window.FileBrowser.AuthLogoutURL;
const loginPage: boolean = window.FileBrowser.LoginPage;
const theme: UserTheme = window.FileBrowser.Theme;
const enableThumbs: boolean = window.FileBrowser.EnableThumbs;
const resizePreview: boolean = window.FileBrowser.ResizePreview;
const enableExec: boolean = window.FileBrowser.EnableExec;
const tusSettings = window.FileBrowser.TusSettings;
const origin = window.location.origin;
const tusEndpoint = `/api/tus`;

export {
  name,
  disableExternal,
  disableUsedPercentage,
  baseURL,
  logoURL,
  recaptcha,
  recaptchaKey,
  signup,
  version,
  tmpDir,
  trashDir,
  quotaExists,
  noAuth,
  authMethod,
  authLogoutURL,
  loginPage,
  theme,
  enableThumbs,
  resizePreview,
  enableExec,
  tusSettings,
  origin,
  tusEndpoint,
};
