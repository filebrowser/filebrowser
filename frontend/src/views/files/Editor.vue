<template>
  <div id="editor-container">
    <header-bar>
      <action icon="close" :label="t('buttons.close')" @action="close()" />
      <title>{{ fileStore.req?.name ?? "" }}</title>

      <action
        icon="add"
        @action="increaseFontSize"
        :label="t('buttons.increaseFontSize')"
      />
      <span class="editor-font-size">{{ fontSize }}px</span>
      <action
        icon="remove"
        @action="decreaseFontSize"
        :label="t('buttons.decreaseFontSize')"
      />

      <action
        v-if="authStore.user?.perm.modify"
        id="save-button"
        icon="save"
        :label="t('buttons.save')"
        @action="save()"
      />

      <action
        icon="preview"
        :label="t('buttons.preview')"
        @action="preview()"
        v-show="isMarkdownFile"
      />

      <action
        v-if="isGcodeFile && authStore.user?.perm.modify && !cncRunning"
        icon="send"
        :label="t('buttons.sendToMachine')"
        @action="promptSendToMachine()"
      />
      <action
        v-if="isGcodeFile && authStore.user?.perm.modify && cncRunning"
        icon="stop_circle"
        :label="t('buttons.stopMachine')"
        @action="promptStopMachine()"
      />
    </header-bar>

    <!-- loading spinner -->
    <div class="loading delayed" v-if="layoutStore.loading">
      <div class="spinner">
        <div class="bounce1"></div>
        <div class="bounce2"></div>
        <div class="bounce3"></div>
      </div>
    </div>

    <template v-else>
      <div class="editor-header">
        <Breadcrumbs base="/files" noLink />

        <div>
          <button
            :disabled="isSelectionEmpty"
            @click="executeEditorCommand('copy')"
          >
            <span><i class="material-icons">content_copy</i></span>
          </button>
          <button
            :disabled="isSelectionEmpty"
            @click="executeEditorCommand('cut')"
          >
            <span><i class="material-icons">content_cut</i></span>
          </button>
          <button @click="executeEditorCommand('paste')">
            <span><i class="material-icons">content_paste</i></span>
          </button>
          <button @click="executeEditorCommand('openCommandPalette')">
            <span><i class="material-icons">more_vert</i></span>
          </button>
        </div>
      </div>

      <!-- markdown preview -->
      <div
        v-show="isPreview && isMarkdownFile"
        id="preview-container"
        class="md_preview"
        v-html="previewContent"
      ></div>

      <!-- editor + (optional) 3D viewer -->
      <div
        v-show="!isPreview || !isMarkdownFile"
        class="editor-layout"
        :style="
          isGcodeFile
            ? { '--editor-pct': editorPct + '%' }
            : { '--editor-pct': '100%' }
        "
      >
        <div class="editor-pane" id="editor"></div>

        <template v-if="isGcodeFile">
          <div
            class="splitter"
            role="separator"
            aria-orientation="vertical"
            @pointerdown="startResize"
            @dblclick="resetSplit"
            :title="t('buttons.dragToResize') || 'Drag to resize · double-click to reset'"
          ></div>
          <div class="viewer-pane">
            <GCode3DViewer
              :gcode="debouncedGcode"
              :cursor-line="cursorLine"
              :machine-line="machineLine"
              @select-line="handleViewerLineSelect"
            />
            <button
              v-if="cncRunning"
              class="follow-machine-btn"
              :class="{ active: followMachine }"
              :title="
                followMachine
                  ? t('buttons.followMachineOn')
                  : t('buttons.resumeFollow')
              "
              @click="toggleFollowMachine"
            >
              <i class="material-icons">precision_manufacturing</i>
              <span>{{
                followMachine
                  ? t("buttons.followMachineOn")
                  : t("buttons.resumeFollow")
              }}</span>
            </button>
          </div>
        </template>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import "@/ace-gcode.js";

import { files as api, cnc as cncApi } from "@/api";
import buttons from "@/utils/buttons";
import url from "@/utils/url";

import ace, { Ace } from "ace-builds";
import "ace-builds/src-noconflict/ext-language_tools";
import modelist from "ace-builds/src-noconflict/ext-modelist";
import "@/utils/aceModules";
import DOMPurify from "dompurify";

import Breadcrumbs from "@/components/Breadcrumbs.vue";
import Action from "@/components/header/Action.vue";
import HeaderBar from "@/components/header/HeaderBar.vue";
import { useAuthStore } from "@/stores/auth";
import { useCncStore } from "@/stores/cnc";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { getEditorTheme } from "@/utils/theme";
import { marked } from "marked";
import markedKatex from "marked-katex-extension";
import {
  computed,
  inject,
  onBeforeUnmount,
  onMounted,
  ref,
  watch,
  watchEffect,
} from "vue";
import { useI18n } from "vue-i18n";
import { onBeforeRouteUpdate, useRoute, useRouter } from "vue-router";
import { read, copy } from "@/utils/clipboard";

import GCode3DViewer from "@/components/GCode3DViewer.vue";

// debounce helper — avoids pulling in lodash for one call site
function debounce<T extends (...args: any[]) => void>(fn: T, delay: number): T {
  let timer: ReturnType<typeof setTimeout> | null = null;
  return ((...args: any[]) => {
    if (timer) clearTimeout(timer);
    timer = setTimeout(() => fn(...args), delay);
  }) as T;
}

const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;

const fileStore = useFileStore();
const authStore = useAuthStore();
const cncStore = useCncStore();
const layoutStore = useLayoutStore();

const { t } = useI18n();

const route = useRoute();
const router = useRouter();

const editor = ref<Ace.Editor | null>(null);
const cursorLine = ref<number | null>(null);
const fontSize = ref(parseInt(localStorage.getItem("editorFontSize") || "14"));

const isPreview = ref(false);
const previewContent = ref("");

const isMarkdownFile = computed(() => {
  const name = fileStore.req?.name ?? "";
  return name.endsWith(".md") || name.endsWith(".markdown");
});

const isGcodeFile = computed(() => {
  const name = (fileStore.req?.name ?? "").toLowerCase();
  return (
    name.endsWith(".nc") ||
    name.endsWith(".tap") ||
    name.endsWith(".gcode") ||
    name.endsWith(".cnc")
  );
});

const katexOptions = {
  output: "mathml" as const,
  throwOnError: false,
};
marked.use(markedKatex(katexOptions));

const isSelectionEmpty = ref(true);

// ── Send-to-Machine (Z-9) + Machine tracker (Z-10) ─────────────────────────
// cncRunning + machineLine are driven by the global useCncStore — same
// source the header pill reads, so all surfaces stay in sync via the
// /api/cnc/stream WebSocket (with the 2 s poll backstop).
const cncHaasHostLabel = ref<string>("");
const cncRunning = computed(() => cncStore.running);

// machineLine is null unless the streamer is on THIS file. Without the
// path check the marker would jump around when a different file is
// being streamed but the operator is reviewing this one in the editor.
const isStreamingThisFile = computed(() => {
  if (!cncStore.running) return false;
  const here = "/" + (fileStore.req?.path?.replace(/^\/+/, "") ?? "");
  const there = "/" + (cncStore.filePath?.replace(/^\/+/, "") ?? "");
  return here === there;
});
const machineLine = computed(() =>
  isStreamingThisFile.value ? cncStore.lineCurrent : null
);

const refreshCncStatus = () => cncStore.pollOnce();

// Follow-machine: when the user is streaming THIS file, snap the
// editor cursor + 3D viewer cursor to wherever the machine is. User
// click/keystroke in either pane breaks the lock; clicking the toggle
// resumes. The doc explicitly said "approximation is fine" — we don't
// try to be perfectly synchronous with the wire, just close.
const followMachine = ref(true);
let programmaticCursorMove = false;

const snapToLine = (line: number) => {
  if (!editor.value) return;
  // Ace lines are 0-indexed; the streamer emits 1-based.
  const row = Math.max(0, line - 1);
  programmaticCursorMove = true;
  editor.value.gotoLine(row + 1, 0, false);
  cursorLine.value = row;
  // Clear the flag after Ace's internal events have flushed.
  setTimeout(() => {
    programmaticCursorMove = false;
  }, 0);
};

const promptStopMachine = () => {
  layoutStore.showHover({
    prompt: "stopMachine",
    props: {
      filePath: cncStore.filePath,
      lineCurrent: cncStore.lineCurrent,
    },
    confirm: async (event: Event) => {
      event.preventDefault();
      layoutStore.closeHovers();
      try {
        await cncApi.stop();
        cncStore.pollOnce();
      } catch (e: any) {
        $showError(e);
      }
    },
  });
};

const toggleFollowMachine = () => {
  followMachine.value = !followMachine.value;
  if (followMachine.value && machineLine.value != null) {
    snapToLine(machineLine.value);
  }
};

const promptSendToMachine = async () => {
  await refreshCncStatus();
  if (cncStore.recoveryPending) {
    $showError(new Error(t("errors.cncRecoveryPending") as string));
    return;
  }
  if (cncRunning.value) {
    $showError(new Error(t("errors.cncJobAlreadyRunning") as string));
    return;
  }
  // Resolve host label from the default machine (machines[0]) for the
  // confirm-prompt destination line. Multi-machine destination
  // selection lands in a follow-on PR.
  let settings: { haasHost: string; haasPort: number } | null = null;
  try {
    const s = await cncApi.getSettings();
    const m = s.machines?.[0];
    if (m) {
      settings = { haasHost: m.host, haasPort: m.port };
    }
    cncHaasHostLabel.value = settings?.haasHost
      ? `${settings.haasHost}:${settings.haasPort}`
      : "";
  } catch {
    cncHaasHostLabel.value = "";
  }
  if (!settings?.haasHost) {
    $showError(
      new Error(t("errors.cncSettingsMissing") as string)
    );
    return;
  }
  const filePath = "/" + (fileStore.req?.path?.replace(/^\/+/, "") ?? "");
  const lineCount = (editor.value?.session.getLength() ?? 0) || 0;

  layoutStore.showHover({
    prompt: "sendToMachine",
    props: {
      filePath,
      lineCount,
      haasHostLabel: cncHaasHostLabel.value,
    },
    confirm: async (event: Event) => {
      event.preventDefault();
      layoutStore.closeHovers();
      try {
        await cncApi.start(filePath);
        // Trigger an immediate poll so cncRunning (a computed off the
        // store) flips before the WS broadcasts the status frame.
        cncStore.pollOnce();
        followMachine.value = true;
        $showSuccess(t("buttons.sendingToMachine"));
        // Operator's eyes belong on the machine page now — live state,
        // progress bar, and the NC mirror are all there. Use the
        // router so the editor cleanly tears down (Ace + 3D scene).
        router.push({ name: "Machine" });
      } catch (e: any) {
        $showError(e);
      }
    },
  });
};

// ── Splitter between code + 3D viewer ──────────────────────────────────────
// Width is stored as a percent of the .editor-layout flexbox row. Default
// is narrow (~22% on the code side) — the toolpath visualization is the
// star here, the NC text is just the source. The localStorage key is
// versioned (V2) so users carrying a wide value from before the redesign
// land on the new default; bump again next time the default changes.
const SPLIT_KEY = "gcodeEditorSplitPctV2";
const SPLIT_DEFAULT = 22;
const SPLIT_MIN = 12;
const SPLIT_MAX = 85;
const editorPct = ref<number>(
  (() => {
    const v = parseFloat(localStorage.getItem(SPLIT_KEY) || "");
    return Number.isFinite(v) && v >= SPLIT_MIN && v <= SPLIT_MAX
      ? v
      : SPLIT_DEFAULT;
  })()
);

const startResize = (e: PointerEvent) => {
  e.preventDefault();
  const target = e.currentTarget as HTMLElement;
  const layout = target.parentElement;
  if (!layout) return;
  const rect = layout.getBoundingClientRect();
  const pointerId = e.pointerId;
  // Pointer-capture keeps drags responsive on touch (tracking the
  // finger after it slides past the 6 px hit area). The document-level
  // listeners below are the actual transport — they work even if
  // capture fails or Vue re-renders the splitter element mid-drag.
  try {
    target.setPointerCapture(pointerId);
  } catch {
    /* ignore — document listeners cover us */
  }
  const onMove = (ev: PointerEvent) => {
    const pct = ((ev.clientX - rect.left) / rect.width) * 100;
    editorPct.value = Math.min(SPLIT_MAX, Math.max(SPLIT_MIN, pct));
    editor.value?.resize();
  };
  const onUp = () => {
    document.removeEventListener("pointermove", onMove);
    document.removeEventListener("pointerup", onUp);
    document.removeEventListener("pointercancel", onUp);
    document.body.style.cursor = "";
    document.body.style.userSelect = "";
    try {
      target.releasePointerCapture(pointerId);
    } catch {
      /* ignore */
    }
    localStorage.setItem(SPLIT_KEY, String(editorPct.value));
    editor.value?.resize();
  };
  document.body.style.cursor = "col-resize";
  document.body.style.userSelect = "none";
  document.addEventListener("pointermove", onMove);
  document.addEventListener("pointerup", onUp);
  document.addEventListener("pointercancel", onUp);
};

const resetSplit = () => {
  editorPct.value = SPLIT_DEFAULT;
  localStorage.setItem(SPLIT_KEY, String(SPLIT_DEFAULT));
  editor.value?.resize();
};

// Debounced gcode — only updates after the user stops typing for 600ms,
// preventing the parser + Three.js geometry rebuild from running on every keystroke.
const debouncedGcode = ref<string>(fileStore.req?.content || "");
const updateDebouncedGcode = debounce((val: string) => {
  debouncedGcode.value = val;
}, 600);

const executeEditorCommand = (name: string) => {
  if (name == "paste") {
    read()
      .then((data) => {
        editor.value?.execCommand("paste", {
          text: data,
        });
      })
      .catch((e) => {
        if (
          document.queryCommandSupported &&
          document.queryCommandSupported("paste")
        ) {
          document.execCommand("paste");
        } else {
          console.warn("the clipboard api is not supported", e);
        }
      });
    return;
  }
  if (name == "copy" || name == "cut") {
    const selectedText = editor.value?.getCopyText();
    copy({ text: selectedText });
  }
  editor.value?.execCommand(name);
};

watchEffect(async () => {
  if (isMarkdownFile.value && isPreview.value) {
    const new_value = editor.value?.getValue() || "";
    try {
      previewContent.value = DOMPurify.sanitize(await marked(new_value));
    } catch (error) {
      console.error("Failed to convert content to HTML:", error);
      previewContent.value = "";
    }
  }
});

onMounted(() => {
  window.addEventListener("keydown", keyEvent);
  window.addEventListener("beforeunload", handlePageChange);

  if (isGcodeFile.value) {
    refreshCncStatus();
  }

  const fileContent = fileStore.req?.content || "";

  // seed the debounced value immediately so the viewer has content on first render
  debouncedGcode.value = fileContent;

  if (!layoutStore.loading) {
    initEditor(fileContent);
  } else {
    const unwatch = watchEffect(() => {
      if (!layoutStore.loading) {
        setTimeout(() => {
          initEditor(fileContent);
          unwatch();
        }, 50);
      }
    });
  }
});

onBeforeUnmount(() => {
  window.removeEventListener("keydown", keyEvent);
  window.removeEventListener("beforeunload", handlePageChange);
  editor.value?.destroy();
});

onBeforeRouteUpdate((to, from, next) => {
  if (editor.value?.session.getUndoManager().isClean()) {
    next();
    return;
  }

  layoutStore.showHover({
    prompt: "discardEditorChanges",
    confirm: (event: Event) => {
      event.preventDefault();
      next();
    },
    saveAction: async () => {
      await save();
      next();
    },
  });
});

const initEditor = (fileContent: string) => {
  editor.value = ace.edit("editor", {
    value: fileContent,
    showPrintMargin: false,
    readOnly: fileStore.req?.type === "textImmutable",
    theme: getEditorTheme(authStore.user?.aceEditorTheme ?? ""),
    mode: isGcodeFile.value
      ? "ace/mode/gcode"
      : modelist.getModeForPath(fileStore.req!.name).mode,
    wrap: true,
    enableBasicAutocompletion: true,
    enableLiveAutocompletion: true,
    enableSnippets: true,
  });

  editor.value.setFontSize(fontSize.value);
  editor.value.focus();

  const selection = editor.value.getSelection();
  selection.on("changeSelection", function () {
    isSelectionEmpty.value = selection.isEmpty();
  });

  // sphere highlight in the 3D viewer follows the editor cursor.
  // Any cursor move that wasn't programmatic (i.e. typed in by the
  // user) breaks the follow-machine lock — Z-10 spec.
  editor.value.getSession().selection.on("changeCursor", () => {
    const pos = editor.value!.getCursorPosition();
    cursorLine.value = pos.row;
    if (!programmaticCursorMove && followMachine.value) {
      followMachine.value = false;
    }
  });

  // re-parse for the viewer only after the user pauses typing
  editor.value.session.on("change", () => {
    if (isGcodeFile.value) {
      updateDebouncedGcode(editor.value!.getValue());
    }
  });
};

const keyEvent = (event: KeyboardEvent) => {
  if (event.code === "Escape") {
    close();
  }
  if (!event.ctrlKey && !event.metaKey) return;
  if (event.key !== "s") return;
  event.preventDefault();
  save();
};

const handlePageChange = (event: BeforeUnloadEvent) => {
  if (!editor.value?.session.getUndoManager().isClean()) {
    event.preventDefault();
    event.returnValue = true;
  }
};

const save = async (throwError?: boolean) => {
  const button = "save";
  buttons.loading("save");
  try {
    await api.put(route.path, editor.value?.getValue());
    editor.value?.session.getUndoManager().markClean();
    buttons.success(button);
  } catch (e: any) {
    buttons.done(button);
    $showError(e);
    if (throwError) throw e;
  }
};

const increaseFontSize = () => {
  fontSize.value += 1;
  editor.value?.setFontSize(fontSize.value);
  localStorage.setItem("editorFontSize", fontSize.value.toString());
};

const decreaseFontSize = () => {
  if (fontSize.value > 1) {
    fontSize.value -= 1;
    editor.value?.setFontSize(fontSize.value);
    localStorage.setItem("editorFontSize", fontSize.value.toString());
  }
};

const close = () => {
  if (!editor.value?.session.getUndoManager().isClean()) {
    layoutStore.showHover({
      prompt: "discardEditorChanges",
      confirm: (event: Event) => {
        event.preventDefault();
        editor.value?.session.getUndoManager().reset();
        finishClose();
      },
      saveAction: async () => {
        try {
          await save(true);
          finishClose();
        } catch {}
      },
    });
    return;
  }
  finishClose();
};

const finishClose = () => {
  const uri = url.removeLastDir(route.path) + "/";
  router.push({ path: uri });
};

const preview = () => {
  isPreview.value = !isPreview.value;
};

// click in 3D viewer → jump Ace editor to that line, and break the
// follow-machine lock since this is a manual user action.
const handleViewerLineSelect = (lineIndex: number) => {
  if (!editor.value) return;
  const session = editor.value.getSession();
  const maxRow = session.getLength() - 1;
  const row = Math.max(0, Math.min(lineIndex, maxRow));
  editor.value.gotoLine(row + 1, 0, true);
  editor.value.centerSelection();
  if (followMachine.value) followMachine.value = false;
};

// Z-10: while followMachine is on and a job is streaming this file,
// snap the cursor to the machine line as it advances. Throttled to
// prevent thrash on a fast Haas — we only resync if the line moved.
let lastFollowedLine = -1;
watch(
  () => [machineLine.value, followMachine.value] as [number | null, boolean],
  ([line, follow]: [number | null, boolean]) => {
    if (!follow || line == null) return;
    if (line === lastFollowedLine) return;
    lastFollowedLine = line;
    snapToLine(line);
  }
);
</script>

<style scoped>
.editor-font-size {
  margin: 0 0.5em;
  color: var(--fg);
}

.editor-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.editor-header > div > button {
  background: transparent;
  color: var(--action);
  border: none;
  outline: none;
  opacity: 0.8;
  cursor: pointer;
}

.editor-header > div > button:hover:not(:disabled) {
  opacity: 1;
}

.editor-header > div > button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.editor-header > div > button > span > i {
  font-size: 1.2rem;
}
</style>

<style>
/*
 * G-code accents for ace-gcode.js.
 *
 * Colors come from the active Ace theme (Ambiance, Monokai, Solarized…)
 * via the standard token categories the mode emits — keyword, variable,
 * support.function, comment, etc. The rules below add only structural
 * accents (italic, weight, dim) that work regardless of theme.
 */

.ace_comment.ace_gcode {
  font-style: italic;
}

.ace_constant.ace_block.ace_gcode {
  /* N-codes recede so the actual codes stand out. */
  opacity: 0.55;
}

.ace_keyword.ace_marker.ace_gcode,
.ace_keyword.ace_gword.ace_gcode,
.ace_keyword.ace_mcode.ace_gcode {
  font-weight: bold;
}

.ace_constant.ace_numeric {
  opacity: 0.6;
}

#editor-container {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 52px); /* 52px = header-bar height */
}

.editor-layout {
  display: flex;
  flex: 1;
  min-height: 0;
  /* Default split is owned by the inline style on the row; this fallback
     keeps the editor visible if the script is somehow late binding.
     Kept in sync with SPLIT_DEFAULT in the script block. */
  --editor-pct: 22%;
}

.editor-pane {
  flex: 0 0 var(--editor-pct);
  min-width: 0;
}

.viewer-pane {
  flex: 1 1 0;
  min-width: 0;
  border-left: 1px solid var(--border-color, #333);
  position: relative;
}

/* Follow-machine toggle floats over the 3D viewer, bottom-right. Only
   rendered while a job is running, so the placement doesn't compete
   with the toolbar in the idle state. */
.follow-machine-btn {
  position: absolute;
  right: 12px;
  bottom: 12px;
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.4rem 0.7rem;
  border: 1px solid #333;
  border-radius: 999px;
  background: rgba(20, 20, 20, 0.85);
  color: #ccc;
  font-size: 0.8rem;
  font-weight: 500;
  cursor: pointer;
  z-index: 2;
  transition: background 0.15s, color 0.15s;
}

.follow-machine-btn .material-icons {
  font-size: 1rem;
}

.follow-machine-btn:hover {
  background: rgba(40, 40, 40, 0.9);
  color: #fff;
}

.follow-machine-btn.active {
  background: rgba(51, 255, 102, 0.15);
  border-color: #33ff66;
  color: #66ff99;
}

/* 10 px hit area, 1 px visible line. Bumped from 6 px because the
   thinner strip was too narrow to grab reliably on touch. Hover/drag
   highlights it so the user can find the grip. */
.splitter {
  flex: 0 0 10px;
  cursor: col-resize;
  background: transparent;
  position: relative;
  user-select: none;
  z-index: 2;
  /* Stops the browser from interpreting drag as a horizontal scroll
     gesture on touch devices, which was eating the move events. */
  touch-action: none;
}

.splitter::before {
  content: "";
  position: absolute;
  top: 0;
  bottom: 0;
  left: 50%;
  width: 1px;
  background: var(--border-color, #333);
  transform: translateX(-50%);
}

.splitter:hover::before,
.splitter:active::before {
  background: var(--primaryColor, #2196f3);
  width: 2px;
}
</style>
