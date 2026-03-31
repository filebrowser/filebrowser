import { useLayoutStore } from "@/stores/layout";
import { useUploadStore } from "@/stores/upload";
import url from "@/utils/url";
import { files as api } from "@/api";

interface UploadEntryWithChild extends UploadEntry {
  children?: UploadEntryWithChild[];
  originalIndex: number;
}

/**
 * Convert UploadList into a forest (array of root nodes).
 * This properly handles uploads with multiple top-level files/folders
 * instead of assuming a single root.
 * @param flatArray
 * @param basePath
 */
function flatToForest(
  flatArray: UploadList,
  basePath: string
): UploadEntryWithChild[] {
  const nodeMap: Record<string, UploadEntryWithChild> = {};

  // First pass: create all nodes
  flatArray.forEach((item, index) => {
    const fullPathOrTo = item.fullPath || item.to?.replace(basePath, "");
    nodeMap[fullPathOrTo!] = {
      fullPath: fullPathOrTo,
      isDir: item.isDir,
      name: item.name,
      size: item.size,
      originalIndex: index,
      ...(item.isDir && { children: [] }),
      ...(item.file && { file: item.file }),
    };
  });

  const roots: UploadEntryWithChild[] = [];

  // Second pass: build hierarchy
  flatArray.forEach((item) => {
    const fullPathOrTo = item.fullPath || item.to?.replace(basePath, "");

    const node = nodeMap[fullPathOrTo!];
    const lastSlash = fullPathOrTo!.lastIndexOf("/");

    if (lastSlash === -1) {
      roots.push(node);
    } else {
      const parentPath = fullPathOrTo!.substring(0, lastSlash);
      const parent = nodeMap[parentPath];
      if (parent?.children) {
        parent.children.push(node);
      }
    }
  });

  return roots;
}

/**
 * Return conflict files from
 * @param files  - flat upload list to check
 * @param basePath   - server destination path (e.g. "/files/uploads/")
 */
export async function checkConflict(
  files: UploadList,
  basePath: string
): Promise<ConflictingResource[]> {
  console.log("Starting deepCheck conflict");
  console.debug(files.length + " possible conflict found:");
  console.debug(files);

  const forest = flatToForest(files, basePath);
  if (forest.length === 0) return [];

  const conflicts: ConflictingResource[] = [];

  /**
   * Check a list of sibling nodes against the server listing at the given path.
   * For directories that exist on the server, recurse into their children.
   * @param nodes   - sibling nodes to check at this level
   * @param serverPath - the server directory path to fetch (must end with "/")
   */
  async function recursiveCheckConflict(
    nodes: UploadEntryWithChild[],
    serverPath: string
  ): Promise<void> {
    let serverItems: ResourceItem[] = [];

    try {
      const res = await api.fetch(serverPath);
      serverItems = res.items || [];
    } catch {
      // Directory doesn't exist on server, no conflicts possible
      console.error(
        `Failed to fetch server listing for ${serverPath}. ` +
          `Assuming directory doesn't exist and skipping conflict check for this branch.`
      );
      return;
    }

    for (const node of nodes) {
      if (node.isDir && node.children) {
        // Check if this directory exists on the server before recursing
        const dirExists = serverItems.some(
          (item) =>
            item.url.replaceAll("/", "") ===
            `${serverPath}${node.name}`.replaceAll("/", "")
        );

        if (dirExists) {
          await recursiveCheckConflict(
            node.children,
            serverPath + encodeURIComponent(node.name) + "/"
          );
        }
      } else {
        // File – check for a conflict against the server listing
        const cleanFullPath = `${basePath}${node.fullPath!}`.replaceAll(
          "/",
          ""
        );
        const serverItem =
          serverItems.find(
            (item) => item.url.replaceAll("/", "") === cleanFullPath
          ) ?? null;

        if (serverItem) {
          conflicts.push({
            index: node.originalIndex,
            name: serverItem.path,
            origin: {
              lastModified: node.file?.lastModified,
              size: node.size,
            },
            dest: {
              lastModified: serverItem.modified,
              size: serverItem.size,
            },
            checked: ["origin"],
          });
        }
      }
    }
  }

  // Check all root nodes against the base destination
  await recursiveCheckConflict(forest, basePath);

  console.debug(conflicts.length + " conflicts found:");
  console.debug(conflicts);

  conflicts.sort((a, b) => a.index - b.index);

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
