import { beforeEach, describe, expect, it, vi } from "vitest";
import FileListing from "@/views/files/FileListing.vue";
import UploadPrompt from "@/components/prompts/Upload.vue";

const harness = vi.hoisted(() => ({
  checkConflict: vi.fn(),
  scanFiles: vi.fn(),
  handleFiles: vi.fn(),
  queueUpload: vi.fn(),
  showHover: vi.fn(),
  closeHovers: vi.fn(),
  fileStore: {
    req: { items: [] },
    selected: [] as number[],
    selectedCount: 0,
    multiple: false,
    preselect: "",
  },
}));

vi.mock("@/stores/auth", () => ({
  useAuthStore: () => ({ user: null }),
}));
vi.mock("@/stores/clipboard", () => ({
  useClipboardStore: () => ({ items: [], $patch: vi.fn() }),
}));
vi.mock("@/stores/file", () => ({
  useFileStore: () => harness.fileStore,
}));
vi.mock("@/stores/layout", () => ({
  useLayoutStore: () => ({
    currentPrompt: null,
    showHover: harness.showHover,
    closeHovers: harness.closeHovers,
  }),
}));
vi.mock("@/stores/upload", () => ({
  useUploadStore: () => ({ upload: harness.queueUpload }),
}));
vi.mock("@/api", () => ({
  users: {},
  files: {},
}));
vi.mock("@/utils/constants", () => ({ enableExec: false }));
vi.mock("@/utils/auth", () => ({}));
vi.mock("@/router", () => ({ default: {} }));
vi.mock("@/i18n", () => ({ default: {} }));
vi.mock("@/utils/upload", () => ({
  checkConflict: harness.checkConflict,
  scanFiles: harness.scanFiles,
  handleFiles: harness.handleFiles,
}));
vi.mock("@/utils/css", () => ({ default: vi.fn() }));
// Reaches document.querySelector, which the node test environment lacks.
vi.mock("@/utils/buttons", () => ({
  default: { loading: vi.fn(), done: vi.fn(), success: vi.fn() },
}));
vi.mock("@/components/header/HeaderBar.vue", () => ({ default: {} }));
vi.mock("@/components/header/Action.vue", () => ({ default: {} }));
vi.mock("@/components/Search.vue", () => ({ default: {} }));
vi.mock("@/components/files/ListingItem.vue", () => ({ default: {} }));
vi.mock("@/components/ContextMenu.vue", () => ({ default: {} }));
vi.mock("vue-router", () => ({
  useRoute: () => ({ path: "/files/target/" }),
  onBeforeRouteUpdate: vi.fn(),
}));
vi.mock("vue-i18n", async (importOriginal) => ({
  ...(await importOriginal<typeof import("vue-i18n")>()),
  useI18n: () => ({ t: (key: string) => key }),
}));
vi.mock("pinia", async (importOriginal) => ({
  ...(await importOriginal<typeof import("pinia")>()),
  storeToRefs: () => ({ req: { value: harness.fileStore.req } }),
}));
vi.mock("vue", async (importOriginal) => ({
  ...(await importOriginal<typeof import("vue")>()),
  inject: () => vi.fn(),
  watch: vi.fn(),
  onMounted: vi.fn(),
  onBeforeUnmount: vi.fn(),
  useSSRContext: () => ({ modules: new Set<string>() }),
}));

const file = (name: string): UploadEntry => ({
  file: { name, size: 12, type: "text/plain" } as File,
  name,
  size: 12,
  isDir: false,
});

const conflict = (
  index: number,
  checked: Array<"origin" | "dest">
): ConflictingResource => ({
  index,
  name: `/target/file-${index}.txt`,
  origin: { size: 12 },
  dest: { size: 10 },
  checked,
  isSmallerOnServer: true,
});

function setup(component: any) {
  return component.setup({}, { expose: vi.fn() });
}

function confirmPrompt(result: ConflictingResource[]) {
  const prompt = harness.showHover.mock.calls[0][0];
  prompt.confirm({ preventDefault: vi.fn() }, result);
}

async function runActualHandleFiles() {
  const [files, path, overwrite] = harness.handleFiles.mock.calls[0];
  const actual =
    await vi.importActual<typeof import("@/utils/upload")>("@/utils/upload");
  actual.handleFiles(files, path, overwrite);
}

function uploadFlags() {
  return harness.queueUpload.mock.calls.map((call) => call[3]);
}

describe("upload conflict resolution", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    harness.fileStore.preselect = "";
    vi.stubGlobal("window", {
      innerWidth: 1024,
      innerHeight: 768,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    });
    vi.stubGlobal("document", {
      getElementsByClassName: () => [],
    });
  });

  it("authorizes an explicit drag/drop Replace choice", async () => {
    const upload = file("file-0.txt");
    harness.scanFiles.mockResolvedValue([upload]);
    harness.checkConflict.mockResolvedValue([conflict(0, ["origin"])]);
    const bindings = setup(FileListing);

    await bindings.drop({
      preventDefault: vi.fn(),
      dataTransfer: { files: [upload.file], items: [] },
      target: {
        classList: { contains: () => false },
        parentElement: null,
      },
    });
    confirmPrompt([conflict(0, ["origin"])]);
    await runActualHandleFiles();

    expect(uploadFlags()).toEqual([true]);
  });

  it("keeps the forbidden both-selected upload overwrite-safe", async () => {
    const upload = file("file-0.txt");
    harness.scanFiles.mockResolvedValue([upload]);
    harness.checkConflict.mockResolvedValue([conflict(0, ["origin"])]);
    const bindings = setup(FileListing);

    await bindings.drop({
      preventDefault: vi.fn(),
      dataTransfer: { files: [upload.file], items: [] },
      target: {
        classList: { contains: () => false },
        parentElement: null,
      },
    });
    confirmPrompt([conflict(0, ["origin", "dest"])]);
    await runActualHandleFiles();

    expect(uploadFlags()).toEqual([false]);
  });

  it("applies Replace only to the selected file in a mixed batch", async () => {
    const conflicting = file("file-0.txt");
    const nonconflicting = file("new-file.txt");
    harness.checkConflict.mockResolvedValue([conflict(0, ["origin"])]);
    const bindings = setup(FileListing);

    await bindings.uploadInput({
      currentTarget: { files: [conflicting.file, nonconflicting.file] },
    });
    confirmPrompt([conflict(0, ["origin"])]);
    await runActualHandleFiles();

    expect(uploadFlags()).toEqual([true, false]);
  });

  it("keeps a nonconflicting sibling protected after Skip", async () => {
    const conflicting = file("file-0.txt");
    const nonconflicting = file("late-file.txt");
    harness.checkConflict.mockResolvedValue([conflict(0, ["origin"])]);
    const bindings = setup(FileListing);

    await bindings.uploadInput({
      currentTarget: { files: [conflicting.file, nonconflicting.file] },
    });
    confirmPrompt([conflict(0, ["dest"])]);
    await runActualHandleFiles();

    expect(harness.queueUpload).toHaveBeenCalledTimes(1);
    expect(harness.queueUpload).toHaveBeenCalledWith(
      "/files/target/late-file.txt",
      "late-file.txt",
      nonconflicting.file,
      false,
      "text"
    );
  });

  it("matches the Upload prompt per-file overwrite semantics", async () => {
    const conflicting = file("file-0.txt");
    const nonconflicting = file("new-file.txt");
    harness.checkConflict.mockResolvedValue([conflict(0, ["origin"])]);
    const bindings = setup(UploadPrompt);

    await bindings.uploadInput({
      currentTarget: { files: [conflicting.file, nonconflicting.file] },
    });
    confirmPrompt([conflict(0, ["origin"])]);
    await runActualHandleFiles();

    expect(uploadFlags()).toEqual([true, false]);
  });
});
