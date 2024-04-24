// Based on code provided by Amir Fo
// https://stackoverflow.com/a/74528564
export function copy(text: string) {
  return new Promise<void>((resolve, reject) => {
    if (
      typeof navigator !== "undefined" &&
      typeof navigator.clipboard !== "undefined" &&
      // @ts-ignore
      navigator.permissions !== "undefined"
    ) {
      navigator.permissions
        // @ts-ignore
        .query({ name: "clipboard-write" })
        .then((permission) => {
          if (permission.state === "granted" || permission.state === "prompt") {
            const type = "text/plain";
            const blob = new Blob([text], { type });
            const data = [new ClipboardItem({ [type]: blob })];
            navigator.clipboard.write(data).then(resolve).catch(reject);
          } else {
            reject(new Error("Permission not granted!"));
          }
        })
        .catch((e) => {
          // Firefox doesn't support clipboard-write permission
          if (navigator.userAgent.indexOf("Firefox") != -1) {
            navigator.clipboard.writeText(text).then(resolve).catch(reject);
          } else {
            reject(e);
          }
        });
    } else if (
      document.queryCommandSupported &&
      document.queryCommandSupported("copy")
    ) {
      const textarea = document.createElement("textarea");
      textarea.textContent = text;
      textarea.setAttribute("readonly", "");
      textarea.style.fontSize = "12pt";
      textarea.style.position = "fixed";
      textarea.style.width = "2em";
      textarea.style.height = "2em";
      textarea.style.padding = "0";
      textarea.style.margin = "0";
      textarea.style.border = "none";
      textarea.style.outline = "none";
      textarea.style.boxShadow = "none";
      textarea.style.background = "transparent";
      document.body.appendChild(textarea);
      textarea.focus();
      textarea.select();
      try {
        document.execCommand("copy");
        document.body.removeChild(textarea);
        resolve();
      } catch (e) {
        document.body.removeChild(textarea);
        reject(e);
      }
    } else {
      reject(
        new Error("None of copying methods are supported by this browser!")
      );
    }
  });
}
