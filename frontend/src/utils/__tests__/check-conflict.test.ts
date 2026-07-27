import { describe, it, expect, vi, beforeEach } from "vitest";
import { checkConflict } from "@/utils/upload";
import { files as api } from "@/api";

vi.mock("@/api", () => ({
  files: {
    fetch: vi.fn(),
    fetchAll: vi.fn(),
  },
}));

vi.mock("@/api/utils", () => ({
  removePrefix: (value: string) => value.replace(/^\/files/, ""),
}));

// upload.ts imports these at module load; they reach window-bound constants
// which don't exist in the node test environment. checkConflict never uses
// them, so empty stubs keep the import graph from blowing up.
vi.mock("@/stores/layout", () => ({ useLayoutStore: vi.fn() }));
vi.mock("@/stores/upload", () => ({ useUploadStore: vi.fn() }));
vi.mock("@/utils/url", () => ({ default: {} }));

type ServerEntry = {
  path: string;
  name: string;
  size: number;
  modified: string;
  isDir: boolean;
};

// The destination's direct listing, used for copy/move and for flat uploads.
function mockListing(entries: ServerEntry[]) {
  vi.mocked(api.fetch).mockResolvedValue({ items: entries } as Resource);
}

// The recursive walk of the destination, used only for nested (folder) uploads.
function mockRecursiveListing(entries: ServerEntry[]) {
  vi.mocked(api.fetchAll).mockResolvedValue(entries);
}

// A move/copy/drag item carries name (raw) and to (URL-encoded) but no
// fullPath - mirroring what Move.vue / Copy.vue / ListingItem.vue build.
function moveItem(name: string, dest: string, size = 12) {
  return {
    from: `/files/source/${encodeURIComponent(name)}`,
    to: dest + encodeURIComponent(name),
    name,
    size,
    isDir: false,
    overwrite: false,
    rename: false,
  };
}

describe("checkConflict", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("detects a conflict for a plain filename", async () => {
    mockListing([
      {
        path: "/target/file.txt",
        name: "file.txt",
        size: 10,
        modified: "2026-06-04T00:00:00Z",
        isDir: false,
      },
    ]);

    const conflicts = await checkConflict(
      [moveItem("file.txt", "/files/target/")],
      "/files/target/"
    );

    expect(conflicts).toHaveLength(1);
    expect(conflicts[0].name).toBe("/target/file.txt");
  });

  // Regression for #6006: pressing Upload awaited a full server-side recursive
  // walk of the destination before a single byte was sent, so the UI sat there
  // doing nothing on a large destination. A flat upload can only collide with a
  // direct child, which the plain listing already covers.
  it("does not walk the destination tree for a flat upload", async () => {
    mockListing([]);

    await checkConflict(
      [{ name: "file.txt", size: 12, isDir: false }],
      "/files/target/"
    );

    expect(api.fetch).toHaveBeenCalledWith("/files/target/");
    expect(api.fetchAll).not.toHaveBeenCalled();
  });

  // ...but a folder upload lands entries below the destination, so it still
  // needs the recursive listing to see them.
  it("walks the destination tree for a nested upload", async () => {
    mockRecursiveListing([]);

    await checkConflict(
      [
        {
          name: "file.txt",
          size: 12,
          isDir: false,
          fullPath: "folder/file.txt",
        },
      ],
      "/files/target/"
    );

    expect(api.fetchAll).toHaveBeenCalledWith("/files/target/");
    expect(api.fetch).not.toHaveBeenCalled();
  });

  // Regression for #5957: names with encodable characters (spaces, "#",
  // non-ASCII) were keyed by the URL-encoded `to` value and never matched the
  // server's raw path, so the conflict modal was skipped and the backend
  // returned a bare 409 instead.
  it.each(["my file.txt", "résumé.pdf", "a#b.txt"])(
    "detects a conflict for %s (encodable characters)",
    async (name) => {
      mockListing([
        {
          path: `/target/${name}`,
          name,
          size: 10,
          modified: "2026-06-04T00:00:00Z",
          isDir: false,
        },
      ]);

      const conflicts = await checkConflict(
        [moveItem(name, "/files/target/")],
        "/files/target/"
      );

      expect(conflicts).toHaveLength(1);
      expect(conflicts[0].name).toBe(`/target/${name}`);
    }
  );

  it("reports no conflict when the destination has no matching name", async () => {
    mockListing([
      {
        path: "/target/other.txt",
        name: "other.txt",
        size: 10,
        modified: "2026-06-04T00:00:00Z",
        isDir: false,
      },
    ]);

    const conflicts = await checkConflict(
      [moveItem("my file.txt", "/files/target/")],
      "/files/target/"
    );

    expect(conflicts).toHaveLength(0);
  });

  it("detects nested conflicts for folder uploads via fullPath", async () => {
    mockRecursiveListing([
      {
        path: "/target/folder",
        name: "folder",
        size: 0,
        modified: "2026-06-04T00:00:00Z",
        isDir: true,
      },
      {
        path: "/target/folder/nested file.txt",
        name: "nested file.txt",
        size: 10,
        modified: "2026-06-04T00:00:00Z",
        isDir: false,
      },
    ]);

    const files = [
      { name: "folder", size: 0, isDir: true, fullPath: "folder" },
      {
        name: "nested file.txt",
        size: 12,
        isDir: false,
        fullPath: "folder/nested file.txt",
      },
    ];

    const conflicts = await checkConflict(files, "/files/target/");

    expect(conflicts).toHaveLength(1);
    expect(conflicts[0].name).toBe("/target/folder/nested file.txt");
  });

  // The "upload folder" file input pushes only files (with a relative
  // fullPath) and no directory entries. Conflict detection must still find a
  // nested file even though its parent folder is not in the upload list.
  it("detects nested conflicts when no directory entries are present", async () => {
    mockRecursiveListing([
      {
        path: "/target/folder/deep/file.txt",
        name: "file.txt",
        size: 10,
        modified: "2026-06-04T00:00:00Z",
        isDir: false,
      },
    ]);

    const files = [
      {
        name: "file.txt",
        size: 12,
        isDir: false,
        fullPath: "folder/deep/file.txt",
      },
    ];

    const conflicts = await checkConflict(files, "/files/target/");

    expect(conflicts).toHaveLength(1);
    expect(conflicts[0].name).toBe("/target/folder/deep/file.txt");
  });

  // Copy/move only needs the target directory's direct children. A recursive
  // walk can make the UI look frozen on large destinations (regression #6005).
  it("checks only the direct destination listing for copy/move", async () => {
    mockListing([
      {
        path: "/target/file.txt",
        name: "file.txt",
        size: 10,
        modified: "2026-06-04T00:00:00Z",
        isDir: false,
      },
      {
        path: "/target/folder",
        name: "folder",
        size: 0,
        modified: "2026-06-04T00:00:00Z",
        isDir: true,
      },
    ]);

    const items = [
      moveItem("file.txt", "/files/target/"),
      { ...moveItem("folder", "/files/target/", 0), isDir: true },
    ];

    const conflicts = await checkConflict(items, "/files/target/", true);

    expect(api.fetch).toHaveBeenCalledWith("/files/target/");
    expect(api.fetchAll).not.toHaveBeenCalled();
    expect(conflicts).toHaveLength(2);
    expect(conflicts.map((conflict) => conflict.name)).toEqual([
      "/target/file.txt",
      "/target/folder",
    ]);
  });

  // Uploads merge into an existing folder, so the directory itself must not be
  // reported - only the files inside it can conflict.
  it("ignores a directory conflict for uploads (default)", async () => {
    mockListing([
      {
        path: "/target/folder",
        name: "folder",
        size: 0,
        modified: "2026-06-04T00:00:00Z",
        isDir: true,
      },
    ]);

    const files = [
      { name: "folder", size: 0, isDir: true, fullPath: "folder" },
    ];

    const conflicts = await checkConflict(files, "/files/target/");

    expect(conflicts).toHaveLength(0);
  });

  it("returns no conflicts when the destination listing fails", async () => {
    vi.mocked(api.fetch).mockRejectedValue(new Error("404"));

    const conflicts = await checkConflict(
      [moveItem("file.txt", "/files/target/")],
      "/files/target/"
    );

    expect(conflicts).toHaveLength(0);
  });

  // Regression for #5980: a FileBrowser server running on Windows returns
  // backslash-separated paths, Without normalizing them, the prefix strip and
  // key lookup never match, so the conflict modal is skipped and the backend
  // returns a bare 409.
  it("detects a conflict for backslash-separated server paths (Windows)", async () => {
    mockListing([
      {
        path: "\\target\\file.txt",
        name: "file.txt",
        size: 10,
        modified: "2026-06-04T00:00:00Z",
        isDir: false,
      },
    ]);

    const conflicts = await checkConflict(
      [moveItem("file.txt", "/files/target/")],
      "/files/target/"
    );

    expect(conflicts).toHaveLength(1);
  });

  it("detects nested conflicts for backslash-separated server paths (Windows)", async () => {
    mockRecursiveListing([
      {
        path: "\\target\\folder\\nested file.txt",
        name: "nested file.txt",
        size: 10,
        modified: "2026-06-04T00:00:00Z",
        isDir: false,
      },
    ]);

    const files = [
      {
        name: "nested file.txt",
        size: 12,
        isDir: false,
        fullPath: "folder/nested file.txt",
      },
    ];

    const conflicts = await checkConflict(files, "/files/target/");

    expect(conflicts).toHaveLength(1);
  });

  it.each([
    ["forward slashes", "/target/My SubDir/file.txt"],
    ["Windows backslashes", "\\target\\My SubDir\\file.txt"],
  ])(
    "detects conflicts below an encoded destination path with %s",
    async (_label, serverPath) => {
      mockListing([
        {
          path: serverPath,
          name: "file.txt",
          size: 10,
          modified: "2026-06-04T00:00:00Z",
          isDir: false,
        },
      ]);

      const conflicts = await checkConflict(
        [moveItem("file.txt", "/files/target/My%20SubDir/")],
        "/files/target/My%20SubDir/"
      );

      expect(conflicts).toHaveLength(1);
      expect(conflicts[0].name).toBe("/target/My SubDir/file.txt");
    }
  );

  it("detects nested folder-upload conflicts below an encoded destination path", async () => {
    mockRecursiveListing([
      {
        path: "\\target\\My SubDir\\folder\\nested file.txt",
        name: "nested file.txt",
        size: 10,
        modified: "2026-06-04T00:00:00Z",
        isDir: false,
      },
    ]);

    const conflicts = await checkConflict(
      [
        {
          name: "nested file.txt",
          size: 12,
          isDir: false,
          fullPath: "folder/nested file.txt",
        },
      ],
      "/files/target/My%20SubDir/"
    );

    expect(conflicts).toHaveLength(1);
    expect(conflicts[0].name).toBe("/target/My SubDir/folder/nested file.txt");
  });

  it("leaves malformed destination path segments unchanged", async () => {
    mockListing([
      {
        path: "/target/bad%ZZ/file.txt",
        name: "file.txt",
        size: 10,
        modified: "2026-06-04T00:00:00Z",
        isDir: false,
      },
    ]);

    const conflicts = await checkConflict(
      [moveItem("file.txt", "/files/target/bad%ZZ/")],
      "/files/target/bad%ZZ/"
    );

    expect(conflicts).toHaveLength(1);
  });

  it("does not match files outside the decoded destination path", async () => {
    mockRecursiveListing([
      {
        path: "/target/My SubDir/other.txt",
        name: "other.txt",
        size: 10,
        modified: "2026-06-04T00:00:00Z",
        isDir: false,
      },
      {
        path: "/target/My Other SubDir/folder/file.txt",
        name: "file.txt",
        size: 10,
        modified: "2026-06-04T00:00:00Z",
        isDir: false,
      },
    ]);

    const conflicts = await checkConflict(
      [
        {
          name: "file.txt",
          size: 12,
          isDir: false,
          fullPath: "folder/file.txt",
        },
      ],
      "/files/target/My%20SubDir/"
    );

    expect(conflicts).toHaveLength(0);
  });
});
