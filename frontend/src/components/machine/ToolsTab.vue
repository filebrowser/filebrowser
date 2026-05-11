<template>
  <div class="m-tools-tab">
    <ToolTablePanel
      :machine-id="cnc.currentMachineId || undefined"
      :cnc-running="cnc.running"
      :tool-comments="toolComments"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useCncStore } from "@/stores/cnc";
import ToolTablePanel from "@/components/ToolTablePanel.vue";
import type { Preflight } from "@/api/cnc";

const cnc = useCncStore();

// Preflight (when available) carries the CAM tool-list comments keyed
// by tool number. The Haas convention is tool N → slot N, so a 1:1
// pass-through to the tool table is correct. Empty map when no
// preflight is on file.
const props = defineProps<{
  preflight?: Preflight | null;
}>();

const toolComments = computed<Record<number, string>>(() => {
  const out: Record<number, string> = {};
  for (const t of props.preflight?.tools ?? []) {
    if (t.comment) out[t.tool] = t.comment;
  }
  return out;
});
</script>

<style scoped>
.m-tools-tab {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  overflow: auto;
  overscroll-behavior: contain;
}
</style>
