export {};

declare global {
    interface Window {
        FileBrowser: any;
        grecaptcha: any
    }
    interface HTMLAttributes extends HTMLAttributes {
        title: any
    }
}