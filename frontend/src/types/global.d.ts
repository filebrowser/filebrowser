export {};

declare global {
  interface Window {
    FileBrowser: any;
    grecaptcha: any;
  }

  interface HTMLElement {
    // TODO: no idea what the exact type is
    __vue__: any;
  }
}
