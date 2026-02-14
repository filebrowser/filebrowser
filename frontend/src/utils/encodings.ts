export const availableEncodings = [
  "utf-8",
  "ibm866",
  "iso-8859-2",
  "iso-8859-3",
  "iso-8859-4",
  "iso-8859-5",
  "iso-8859-6",
  "iso-8859-7",
  "iso-8859-8",
  "iso-8859-8-i",
  "iso-8859-10",
  "iso-8859-13",
  "iso-8859-14",
  "iso-8859-15",
  "iso-8859-16",
  "koi8-r",
  "koi8-u",
  "macintosh",
  "windows-874",
  "windows-1250",
  "windows-1251",
  "windows-1252",
  "windows-1253",
  "windows-1254",
  "windows-1255",
  "windows-1256",
  "windows-1257",
  "windows-1258",
  "x-mac-cyrillic",
  "gbk",
  "gb18030",
  "big5",
  "euc-jp",
  "iso-2022-jp",
  "shift_jis",
  "euc-kr",
  "utf-16be",
  "utf-16le",
];

export function decode(content: ArrayBuffer, encoding: string): string {
  const decoder = new TextDecoder(encoding);
  return decoder.decode(content);
}

export function isEncodableResponse(url: string): boolean {
  const extensions = [".csv"];

  if (typeof TextDecoder === "undefined") {
    return false;
  }

  for (const extension of extensions) {
    if (url.endsWith(extension)) {
      return true;
    }
  }

  return false;
}

export async function makeRawResource(
  res: Response,
  url: string
): Promise<Resource> {
  const buffer = await res.arrayBuffer();
  return {
    items: [],
    numDirs: 0,
    numFiles: 0,
    sorting: {} as Sorting,
    index: 0,
    extension: getExtension(url),
    isDir: false,
    isSymlink: false,
    path: url,
    size: buffer.byteLength,
    modified: new Date().toISOString(),
    name: url.split("/").pop() || "",
    type: "text",
    mode: 0,
    url: `/files${url}`,
    rawContent: buffer,
    content: decode(buffer, "utf-8"),
  };
}

function getExtension(url: string): string {
  const lastDotIndex = url.lastIndexOf(".");
  if (lastDotIndex === -1) {
    return "";
  }
  return url.substring(lastDotIndex);
}
