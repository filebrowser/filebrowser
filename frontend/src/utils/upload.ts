import { useUploadStore } from "@/stores/upload";
import type { IFile, Item, FileType } from "@/types";
import url from "@/utils/url";

export function checkConflict(files: IFile[], items: Item[]) {
  if (typeof items === "undefined" || items === null) {
    items = [];
  }

  const folder_upload = files[0].fullPath !== undefined;

  let conflict = false;
  for (let i = 0; i < files.length; i++) {
    const file = files[i];
    let name = file.name;

    if (folder_upload) {
      const dirs = file.fullPath.split("/");
      if (dirs.length > 1) {
        name = dirs[0];
      }
    }

    const res = items.findIndex(function hasConflict(element) {
      // @ts-ignore Don't know what this does
      return element.name === this;
    }, name);

    if (res >= 0) {
      conflict = true;
      break;
    }
  }

  return conflict;
}

export function scanFiles(dt: { [key: string]: any; item: Item }) {
  return new Promise((resolve) => {
    let reading = 0;
    const contents: any[] = [];

    if (dt.items !== undefined) {
      for (const item of dt.items) {
        if (
          item.kind === "file" &&
          typeof item.webkitGetAsEntry === "function"
        ) {
          const entry = item.webkitGetAsEntry();
          readEntry(entry);
        }
      }
    } else {
      resolve(dt.files);
    }

    function readEntry(entry: any, directory = "") {
      if (entry.isFile) {
        reading++;
        entry.file((file: IFile) => {
          reading--;

          file.fullPath = `${directory}${file.name}`;
          contents.push(file);

          if (reading === 0) {
            resolve(contents);
          }
        });
      } else if (entry.isDirectory) {
        const dir = {
          isDir: true,
          size: 0,
          fullPath: `${directory}${entry.name}`,
          name: entry.name,
        };

        contents.push(dir);

        readReaderContent(entry.createReader(), `${directory}${entry.name}`);
      }
    }

    function readReaderContent(reader: any, directory: string) {
      reading++;

      reader.readEntries(function (entries: any[]) {
        reading--;
        if (entries.length > 0) {
          for (const entry of entries) {
            readEntry(entry, `${directory}/`);
          }

          readReaderContent(reader, `${directory}/`);
        }

        if (reading === 0) {
          resolve(contents);
        }
      });
    }
  });
}

function detectType(mimetype: string): FileType {
  if (mimetype.startsWith("video")) return "video";
  if (mimetype.startsWith("audio")) return "audio";
  if (mimetype.startsWith("image")) return "image";
  if (mimetype.startsWith("pdf")) return "pdf";
  if (mimetype.startsWith("text")) return "text";
  return "blob";
}

export function handleFiles(files: IFile[], base: string, overwrite = false) {
  const uploadStore = useUploadStore();

  for (let i = 0; i < files.length; i++) {
    const id = uploadStore.id;
    let path = base;
    const file = files[i];

    if (file.fullPath !== undefined) {
      path += url.encodePath(file.fullPath);
    } else {
      path += url.encodeRFC5987ValueChars(file.name);
    }

    if (file.isDir) {
      path += "/";
    }

    const item: Item = {
      id,
      path,
      file,
      overwrite,
      ...(!file.isDir && { type: detectType(file.type) }),
    };

    uploadStore.upload(item);
  }
}
