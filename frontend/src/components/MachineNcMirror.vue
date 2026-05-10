<template>
  <div class="nc-mirror">
    <div ref="containerEl" class="nc-mirror__editor"></div>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from "vue";
import ace, { Ace } from "ace-builds";
import "ace-builds/src-noconflict/ext-language_tools";
import "@/utils/aceModules";
import "@/ace-gcode.js";
import { useAuthStore } from "@/stores/auth";
import { getEditorTheme } from "@/utils/theme";

const props = defineProps<{
  gcode: string;
  // 1-based line currently being streamed. Triggers auto-scroll +
  // highlight when it changes.
  machineLine?: number | null;
}>();

const containerEl = ref<HTMLElement | null>(null);
const editor = ref<Ace.Editor | null>(null);
const authStore = useAuthStore();

let highlightMarkerId: number | null = null;

const applyMachineLine = (n: number | null | undefined) => {
  const ed = editor.value;
  if (!ed) return;
  if (highlightMarkerId !== null) {
    ed.session.removeMarker(highlightMarkerId);
    highlightMarkerId = null;
  }
  if (!n || n < 1) return;
  const row = n - 1;
  const Range = (ace as any).require("ace/range").Range;
  highlightMarkerId = ed.session.addMarker(
    new Range(row, 0, row, 1),
    "nc-mirror__line-current",
    "fullLine"
  );
  // Scroll to it in the middle of the viewport.
  const rowsVisible = ed.renderer.getScrollBottomRow() - ed.renderer.getScrollTopRow();
  ed.scrollToLine(Math.max(0, row - Math.floor(rowsVisible / 2)), false, true, () => {});
};

onMounted(() => {
  if (!containerEl.value) return;
  editor.value = ace.edit(containerEl.value, {
    value: props.gcode,
    showPrintMargin: false,
    readOnly: true,
    theme: getEditorTheme(authStore.user?.aceEditorTheme ?? ""),
    mode: "ace/mode/gcode",
    wrap: false,
    highlightActiveLine: false,
  });
  editor.value.setHighlightSelectedWord(false);
  editor.value.setOptions({ fontSize: "13px" });
  applyMachineLine(props.machineLine);
});

onBeforeUnmount(() => {
  editor.value?.destroy();
  editor.value = null;
});

watch(
  () => props.gcode,
  (g) => {
    if (!editor.value) return;
    editor.value.session.setValue(g);
    applyMachineLine(props.machineLine);
  }
);

watch(
  () => props.machineLine,
  (n) => applyMachineLine(n)
);
</script>

<style>
.nc-mirror {
  width: 100%;
  height: 100%;
  position: relative;
}

.nc-mirror__editor {
  position: absolute;
  inset: 0;
}

/* Bright bar across the line currently being streamed. !important so
   we override the active-line theme styling that ships with most Ace
   themes. */
.nc-mirror__line-current {
  position: absolute;
  background: rgba(46, 204, 113, 0.22) !important;
  border-left: 3px solid #2ecc71;
  z-index: 10;
}
</style>
