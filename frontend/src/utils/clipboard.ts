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
      const textarea = createTemporaryTextarea(text);
      const body = document.activeElement || document.body;
      try {
        body.appendChild(textarea);
        textarea.focus();
        textarea.select();
        document.execCommand("copy");
        resolve();
      } catch (e) {
        reject(e);
      } finally {
        body.removeChild(textarea);
      }
    } else {
      reject(
        new Error("None of copying methods are supported by this browser!")
      );
    }
  });
}

const styles = {
  fontSize: "12pt",
  position: "fixed",
  top: 0,
  left: 0,
  width: "2em",
  height: "2em",
  padding: 0,
  margin: 0,
  border: "none",
  outline: "none",
  boxShadow: "none",
  background: "transparent"
};

const createTemporaryTextarea = (text:string) => {
  const textarea = document.createElement("textarea");
  textarea.value = text;
  textarea.setAttribute("readonly", "");
  Object.assign(textarea.style, styles);
  return textarea;
};