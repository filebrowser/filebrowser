import { useLayoutStore } from "@/stores/layout";
import { useUploadStore } from "@/stores/upload";
import url from "@/utils/url";
import { files as api } from "@/api";
import { removePrefix } from "@/api/utils";

/**
 * The path used to detect conflicts against the server's recursive listing.
 *
 * It MUST be the raw, un-encoded path relative to the destination:
 *  - `fullPath` is set for folder uploads and drag & drop (e.g. "sub/file.txt").
 *  - `name` is the leaf for every other case (copy/move/paste/single upload),
 *    which is always a flat top-level entry.
 *
 * We never key on `item.to`: it is URL-encoded (`dest + encodeURIComponent(name)`)
 * and would miss conflicts for any name with encodable characters (spaces, "#",
 * non-ASCII, ...), surfacing a raw 409 error instead of the conflict modal.
 * @param item
 */
function conflictKey(item: UploadEntry): string {
  return (item.fullPath || item.name).replace(/^\/+/, "");
}

type ServerConflictEntry = {
  path: string;
  name: string;
  size: number;
  modified: string;
};

async function fetchConflictEntries(
  basePath: string,
  includeDirectories: boolean
): Promise<ServerConflictEntry[]> {
  if (!includeDirectories) {
    return await api.fetchAll(basePath);
  }

  const destination = await api.fetch(basePath);
  return destination.items ?? [];
}

function conflictPath(entry: ServerConflictEntry): string {
  return entry.path.replace(/\\/g, "/");
}

function buildConflictMap(
  serverEntries: ServerConflictEntry[],
  basePath: string,
  includeDirectories: boolean
): Map<string, ServerConflictEntry> {
  const serverMap = new Map<string, ServerConflictEntry>();
  const normBase = removePrefix(basePath).replace(/\/+$/, "");
  for (const entry of serverEntries) {
    // A Windows server may return OS-native backslash separators; normalize to
    // forward slashes so the prefix strip and key lookup line up.
    const path = conflictPath(entry);
    const key = includeDirectories
      ? entry.name
      : path.startsWith(normBase)
        ? path.slice(normBase.length)
        : path;
    serverMap.set(key.replace(/^\/+/, ""), entry);
  }

  return serverMap;
}

/**
 * Return the entries from `files` that already exist under `basePath` on the
 * server, so the caller can prompt the user to overwrite/rename/skip.
 *
 * Directory handling differs by action, hence `includeDirectories`:
 *  - Upload (false): the destination tree is fetched recursively so nested
 *    file uploads can be checked; existing folders are silently merged.
 *  - Copy/move (true): only the destination directory itself is fetched. These
 *    operations move flat top-level selections, so a same-named direct child is
 *    the only preflight conflict the backend will reject.
 *
 * @param files              - flat upload list to check
 * @param basePath           - server destination path (e.g. "/files/uploads/")
 * @param includeDirectories - report directories as conflicts (copy/move)
 */
export async function checkConflict(
  files: UploadList,
  basePath: string,
  includeDirectories = false
): Promise<ConflictingResource[]> {
  if (files.length === 0) return [];

  let serverEntries: ServerConflictEntry[];
  try {
    serverEntries = await fetchConflictEntries(basePath, includeDirectories);
  } catch {
    // The destination doesn't exist yet, so nothing can conflict.
    return [];
  }

  const serverMap = buildConflictMap(
    serverEntries,
    basePath,
    includeDirectories
  );

  const conflicts: ConflictingResource[] = [];
  files.forEach((file, index) => {
    if (file.isDir && !includeDirectories) return; // see directory note above

    const server = serverMap.get(conflictKey(file));
    if (!server) return;

    conflicts.push({
      index,
      name: conflictPath(server),
      origin: { lastModified: file.file?.lastModified, size: file.size },
      dest: { lastModified: server.modified, size: server.size },
      checked: ["origin"],
      isSmallerOnServer: file.size > server.size,
    });
  });

  return conflicts;
}

export function scanFiles(dt: DataTransfer): Promise<UploadList | FileList> {
  return new Promise((resolve) => {
    let reading = 0;
    const contents: UploadList = [];

    if (dt.items) {
      // ts didn't like the for of loop even tho
      // it is the official example on MDN
      // for (const item of dt.items) {
      for (let i = 0; i < dt.items.length; i++) {
        const item = dt.items[i];
        if (
          item.kind === "file" &&
          typeof item.webkitGetAsEntry === "function"
        ) {
          const entry = item.webkitGetAsEntry();
          entry && readEntry(entry);
        }
      }
    } else {
      resolve(dt.files);
    }

    function readEntry(entry: FileSystemEntry, directory = ""): void {
      if (entry.isFile) {
        reading++;
        (entry as FileSystemFileEntry).file((file) => {
          reading--;

          contents.push({
            file,
            name: file.name,
            size: file.size,
            isDir: false,
            fullPath: `${directory}${file.name}`,
          });

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

        readReaderContent(
          (entry as FileSystemDirectoryEntry).createReader(),
          `${directory}${entry.name}`
        );
      }
    }

    function readReaderContent(
      reader: FileSystemDirectoryReader,
      directory: string
    ): void {
      reading++;

      reader.readEntries((entries) => {
        reading--;
        if (entries.length > 0) {
          const dirWithSlash = directory.endsWith("/")
            ? directory
            : `${directory}/`;
          for (const entry of entries) {
            readEntry(entry, dirWithSlash);
          }

          readReaderContent(reader, dirWithSlash);
        }

        if (reading === 0) {
          resolve(contents);
        }
      });
    }
  });
}

function detectType(mimetype: string): ResourceType {
  if (mimetype.startsWith("video")) return "video";
  if (mimetype.startsWith("audio")) return "audio";
  if (mimetype.startsWith("image")) return "image";
  if (mimetype.startsWith("pdf")) return "pdf";
  if (mimetype.startsWith("text")) return "text";
  return "blob";
}

export function handleFiles(
  files: UploadList,
  base: string,
  overwrite = false
) {
  const uploadStore = useUploadStore();
  const layoutStore = useLayoutStore();

  layoutStore.closeHovers();

  for (const file of files) {
    let path = base;

    if (file.fullPath !== undefined) {
      path += url.encodePath(file.fullPath);
    } else {
      path += url.encodeRFC5987ValueChars(file.name);
    }

    if (file.isDir) {
      path += "/";
    }

    const type = file.isDir ? "dir" : detectType((file.file as File).type);

    uploadStore.upload(
      path,
      file.name,
      file.file ?? null,
      file.overwrite || overwrite,
      type
    );
  }
}
