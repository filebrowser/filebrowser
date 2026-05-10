<template>
  <div class="m-tools-tab">
    <ToolTablePanel
      :machine-id="cnc.currentMachineId || undefined"
      :cnc-running="cnc.running"
    />
  </div>
</template>

<script setup lang="ts">
import { useCncStore } from "@/stores/cnc";
import ToolTablePanel from "@/components/ToolTablePanel.vue";

const cnc = useCncStore();
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
