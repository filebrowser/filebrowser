import { describe, it, expect, vi, beforeEach } from "vitest";
import { checkConflict } from "@/utils/upload";
import { files as api } from "@/api";

vi.mock("@/api", () => ({
  files: {
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

// A move/copy/drag item carries `name` (raw) and `to` (URL-encoded) but no
// `fullPath` — mirroring what Move.vue / Copy.vue / ListingItem.vue build.
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
    vi.mocked(api.fetchAll).mockResolvedValue([
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

  // Regression for #5957: names with encodable characters (spaces, "#",
  // non-ASCII) were keyed by the URL-encoded `to` value and never matched the
  // server's raw path, so the conflict modal was skipped and the backend
  // returned a bare 409 instead.
  it.each(["my file.txt", "résumé.pdf", "a#b.txt"])(
    "detects a conflict for %s (encodable characters)",
    async (name) => {
      vi.mocked(api.fetchAll).mockResolvedValue([
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
    vi.mocked(api.fetchAll).mockResolvedValue([
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
    vi.mocked(api.fetchAll).mockResolvedValue([
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
    vi.mocked(api.fetchAll).mockResolvedValue([
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

  // Copy/move stats the whole destination, so a same-named directory is a
  // conflict in its own right (regression for the directory case of #5957).
  it("reports a directory conflict for copy/move (includeDirectories)", async () => {
    vi.mocked(api.fetchAll).mockResolvedValue([
      {
        path: "/target/folder",
        name: "folder",
        size: 0,
        modified: "2026-06-04T00:00:00Z",
        isDir: true,
      },
    ]);

    const items = [{ ...moveItem("folder", "/files/target/", 0), isDir: true }];

    const conflicts = await checkConflict(items, "/files/target/", true);

    expect(conflicts).toHaveLength(1);
    expect(conflicts[0].name).toBe("/target/folder");
  });

  // Uploads merge into an existing folder, so the directory itself must not be
  // reported — only the files inside it can conflict.
  it("ignores a directory conflict for uploads (default)", async () => {
    vi.mocked(api.fetchAll).mockResolvedValue([
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

  it("returns no conflicts when the recursive listing fails", async () => {
    vi.mocked(api.fetchAll).mockRejectedValue(new Error("404"));

    const conflicts = await checkConflict(
      [moveItem("file.txt", "/files/target/")],
      "/files/target/"
    );

    expect(conflicts).toHaveLength(0);
  });
});
