import { describe, it, expect, vi, beforeEach } from "vitest";
import CopyPrompt from "@/components/prompts/Copy.vue";
import MovePrompt from "@/components/prompts/Move.vue";
import { files as api } from "@/api";
import { checkConflict } from "@/utils/upload";

vi.mock("@/api", () => ({
  files: {
    copy: vi.fn().mockResolvedValue(undefined),
    move: vi.fn().mockResolvedValue(undefined),
  },
}));

vi.mock("@/api/utils", () => ({
  removePrefix: (value: string) => value.replace(/^\/files/, ""),
}));

vi.mock("@/utils/buttons", () => ({
  default: {
    loading: vi.fn(),
    success: vi.fn(),
    done: vi.fn(),
  },
}));

vi.mock("@/utils/upload", () => ({
  checkConflict: vi.fn(),
}));

const conflict = [
  {
    index: 0,
    name: "/target/file.txt",
    origin: { size: 12 },
    dest: { size: 10 },
    checked: ["origin"],
    isSmallerOnServer: true,
  },
];

function makeContext() {
  return {
    selected: [0],
    req: {
      items: [
        {
          url: "/files/source/file.txt",
          name: "file.txt",
          size: 12,
          isDir: false,
          modified: "2026-06-04T00:00:00Z",
        },
      ],
    },
    dest: "/files/target/",
    user: { redirectAfterCopyMove: false },
    $route: { path: "/files/source/" },
    $router: { push: vi.fn() },
    reload: false,
    preselect: "",
    showHover: vi.fn(),
    closeHovers: vi.fn(),
    $showError: vi.fn(),
  };
}

const event = {
  preventDefault: vi.fn(),
};

describe("copy and move conflict prompts", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(checkConflict).mockResolvedValue(conflict);
  });

  it("waits for copy conflict detection before calling the copy API", async () => {
    const context = makeContext();

    await (CopyPrompt as any).methods.copy.call(context, event);

    expect(checkConflict).toHaveBeenCalledWith(
      [
        expect.objectContaining({
          to: "/files/target/file.txt",
          isDir: false,
        }),
      ],
      "/files/target/"
    );
    expect(context.showHover).toHaveBeenCalledWith(
      expect.objectContaining({
        prompt: "resolve-conflict",
        props: { conflict },
      })
    );
    expect(api.copy).not.toHaveBeenCalled();
  });

  it("waits for move conflict detection before calling the move API", async () => {
    const context = makeContext();

    await (MovePrompt as any).methods.move.call(context, event);

    expect(checkConflict).toHaveBeenCalledWith(
      [
        expect.objectContaining({
          to: "/files/target/file.txt",
          isDir: false,
        }),
      ],
      "/files/target/"
    );
    expect(context.showHover).toHaveBeenCalledWith(
      expect.objectContaining({
        prompt: "resolve-conflict",
        props: expect.objectContaining({
          conflict,
          files: expect.any(Array),
        }),
      })
    );
    expect(api.move).not.toHaveBeenCalled();
  });
});
