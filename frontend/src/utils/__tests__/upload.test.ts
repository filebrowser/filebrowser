import { describe, it, expect } from "vitest";

/**
 * Reproduces the path-building logic from scanFiles() in upload.ts (lines 118-138).
 *
 * readReaderContent() is called when traversing a dropped folder. The browser's
 * FileSystemDirectoryReader.readEntries() returns entries in batches — you must
 * call it repeatedly until it returns an empty array. Each recursive call in the
 * current code appends "/" to the directory, so the second batch and beyond get
 * double (or triple, etc.) slashes in the constructed fullPath.
 */

type Entry = { name: string; isFile: boolean; isDirectory: boolean };

function simulateScanFiles(
  dirName: string,
  entryBatches: Entry[][]
): string[] {
  const paths: string[] = [];
  let batchIndex = 0;

  // Mirrors readEntry() for files — records fullPath as `${directory}${file.name}`
  function readEntry(entry: Entry, directory = ""): void {
    if (entry.isFile) {
      paths.push(`${directory}${entry.name}`);
    }
  }

  // Mirrors readReaderContent() from upload.ts lines 118-138
  function readReaderContent(directory: string): void {
    const entries = batchIndex < entryBatches.length ? entryBatches[batchIndex] : [];
    batchIndex++;

    if (entries.length > 0) {
      const dirWithSlash = directory.endsWith("/")
        ? directory
        : `${directory}/`;
      for (const entry of entries) {
        readEntry(entry, dirWithSlash);
      }
      readReaderContent(dirWithSlash);
    }
  }

  // Initial call mirrors readEntry() for a directory — upload.ts line 111-114
  readReaderContent(dirName);
  return paths;
}

describe("scanFiles path construction", () => {
  it("should not produce double slashes when readEntries returns multiple batches", () => {
    // Two batches: simulates a large directory where the browser splits
    // readEntries() results across multiple calls
    const paths = simulateScanFiles("TestFolder", [
      [
        { name: "file1.xlsx", isFile: true, isDirectory: false },
        { name: "file2.xlsx", isFile: true, isDirectory: false },
      ],
      [
        { name: "file3.xlsx", isFile: true, isDirectory: false },
      ],
    ]);

    expect(paths).toHaveLength(3);

    for (const p of paths) {
      expect(p, `path "${p}" contains double slash`).not.toContain("//");
    }
  });

  it("single batch should work fine (no regression)", () => {
    const paths = simulateScanFiles("TestFolder", [
      [
        { name: "file1.xlsx", isFile: true, isDirectory: false },
      ],
    ]);

    expect(paths).toEqual(["TestFolder/file1.xlsx"]);
  });
});
