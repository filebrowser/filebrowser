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
      <div v-show="!isPreview || !isMarkdownFile" class="editor-layout">
        <div class="editor-pane" id="editor"></div>

        <div v-if="isGcodeFile" class="viewer-pane">
          <GCode3DViewer
            :gcode="debouncedGcode"
            :cursor-line="cursorLine"
            @select-line="handleViewerLineSelect"
          />
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import "@/ace-gcode.js";

import { files as api } from "@/api";
import buttons from "@/utils/buttons";
import url from "@/utils/url";

import ace, { Ace, version as ace_version } from "ace-builds";
import "ace-builds/src-noconflict/ext-language_tools";
import modelist from "ace-builds/src-noconflict/ext-modelist";
import DOMPurify from "dompurify";

import Breadcrumbs from "@/components/Breadcrumbs.vue";
import Action from "@/components/header/Action.vue";
import HeaderBar from "@/components/header/HeaderBar.vue";
import { useAuthStore } from "@/stores/auth";
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

const fileStore = useFileStore();
const authStore = useAuthStore();
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

  const fileContent = fileStore.req?.content || "";

  // seed the debounced value immediately so the viewer has content on first render
  debouncedGcode.value = fileContent;

  ace.config.set(
    "basePath",
    `https://cdn.jsdelivr.net/npm/ace-builds@${ace_version}/src-min-noconflict/`
  );

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

  // sphere highlight in the 3D viewer follows the editor cursor
  editor.value.getSession().selection.on("changeCursor", () => {
    const pos = editor.value!.getCursorPosition();
    cursorLine.value = pos.row;
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

// click in 3D viewer → jump Ace editor to that line
const handleViewerLineSelect = (lineIndex: number) => {
  if (!editor.value) return;
  const session = editor.value.getSession();
  const maxRow = session.getLength() - 1;
  const row = Math.max(0, Math.min(lineIndex, maxRow));
  editor.value.gotoLine(row + 1, 0, true);
  editor.value.centerSelection();
};
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
 * G-code syntax highlighting for ace-gcode.js.
 *
 * Colors picked in OKLCH at ~60% lightness so they read on both light and
 * dark Ace themes without per-theme overrides. N-codes inherit the theme's
 * own foreground color, just dimmed, so they're always legible.
 */

.ace_gcode.ace_comment {
  color: oklch(55% 0.1 140) !important;
  font-style: italic;
}

.ace_gcode.ace_block {
  /* N-codes track the theme's foreground color, slightly dimmed.
     Works in any theme — dark or light, monokai or solarized — without
     enumerating each one. */
  color: currentColor;
  opacity: 0.65;
}

.ace_gcode.ace_marker {
  color: oklch(60% 0.21 25) !important;
  font-weight: bold;
}

.ace_gcode.ace_gword {
  color: oklch(60% 0.18 305) !important;
  font-weight: bold;
}

.ace_gcode.ace_mcode {
  color: oklch(65% 0.13 85) !important;
  font-weight: bold;
}

.ace_gcode.ace_xparam {
  color: oklch(62% 0.15 50) !important;
}

.ace_gcode.ace_yparam {
  color: oklch(58% 0.13 175) !important;
}

.ace_gcode.ace_zparam {
  color: oklch(58% 0.16 245) !important;
}

.ace_gcode.ace_feedspeed {
  color: oklch(62% 0.13 230) !important;
}

.ace_gcode.ace_subprog {
  color: oklch(62% 0.13 230) !important;
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
}

.editor-pane {
  flex: 1 1 50%;
  min-width: 0;
}

.viewer-pane {
  flex: 1 1 50%;
  min-width: 0;
  border-left: 1px solid var(--border-color, #333);
}
</style>
