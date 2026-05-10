<template>
  <div class="m-tabs">
    <div class="m-tabs__view">
      <button
        class="m-tab"
        :class="{ 'm-tab--active': active === 'gcode' }"
        @click="$emit('select', 'gcode')"
      >
        {{ t("machine.tabGcode3d") }}
      </button>
      <button
        class="m-tab"
        :class="{ 'm-tab--active': active === 'tools' }"
        @click="$emit('select', 'tools')"
      >
        {{ t("machine.tabTools") }}
        <span
          v-if="toolMismatchCount > 0"
          class="m-tab__counter m-tab__counter--warn"
        >
          ⚠ {{ toolMismatchCount }}
        </span>
      </button>
    </div>

    <span v-if="fileTabs.length > 0" class="m-tabs__divider" />
    <div class="m-tabs__files" v-if="fileTabs.length > 0">
      <button
        v-for="ft in visibleFileTabs"
        :key="ft.path"
        class="m-tab m-file-tab"
        :class="{ 'm-tab--active': active === ft.path }"
        :title="ft.title"
        @click="$emit('select', ft.path)"
      >
        <i class="material-icons m-file-tab__icon">{{ ft.icon }}</i>
        <span class="m-file-tab__name">{{ ft.label }}</span>
      </button>
      <button
        v-if="overflowFiles.length > 0"
        class="m-tab m-file-tab"
        @click="overflowOpen = !overflowOpen"
      >
        <i class="material-icons m-file-tab__icon">more_horiz</i>
        <span class="m-file-tab__name">{{ overflowFiles.length }} {{ t("machine.tabMore") }}</span>
      </button>
      <div v-if="overflowOpen" class="m-overflow">
        <button
          v-for="ft in overflowFiles"
          :key="ft.path"
          class="m-overflow__item"
          @click="onPickOverflow(ft.path)"
        >
          <i class="material-icons">{{ ft.icon }}</i>
          {{ ft.label }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

export interface FileTabSpec {
  path: string;
  label: string;
  title: string;
  icon: string;
}

const props = defineProps<{
  active: string;
  toolMismatchCount: number;
  fileTabs: FileTabSpec[];
}>();

const emit = defineEmits<{
  (e: "select", tab: string): void;
}>();

const MAX_VISIBLE = 3;
const overflowOpen = ref(false);

const visibleFileTabs = computed(() => props.fileTabs.slice(0, MAX_VISIBLE));
const overflowFiles = computed(() => props.fileTabs.slice(MAX_VISIBLE));

const onPickOverflow = (path: string) => {
  overflowOpen.value = false;
  emit("select", path);
};
</script>

<style scoped>
.m-tabs {
  display: flex;
  align-items: center;
  padding: 0 2px;
  gap: 4px;
  flex-shrink: 0;
  position: relative;
  min-width: 0;
}
.m-tabs__view { display: flex; gap: 4px; flex-shrink: 0; }
.m-tabs__files { display: flex; gap: 4px; flex: 1; min-width: 0; justify-content: flex-end; overflow: hidden; }
.m-tabs__divider { width: 1px; height: 12px; background: var(--border-color, #ddd); flex-shrink: 0; margin: 0 4px; }

.m-tab {
  font-size: 10px;
  padding: 3px 10px;
  border-radius: 4px;
  background: transparent;
  color: var(--textSecondary, #555);
  cursor: pointer;
  border: 1px solid transparent;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  white-space: nowrap;
  flex-shrink: 0;
}
.m-tab--active {
  background: var(--alt-background, #f4f4f4);
  color: var(--textPrimary, #222);
  border-color: var(--border-color, #ddd);
}
.m-tab__counter {
  font-size: 9px;
  padding: 1px 5px;
  border-radius: 3px;
  font-weight: 500;
}
.m-tab__counter--warn { background: #FAEEDA; color: #854F0B; }

.m-file-tab {
  border-color: var(--border-color, #ddd);
  max-width: 130px;
}
.m-file-tab__icon { font-size: 12px; }
.m-file-tab__name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 10px;
}

.m-overflow {
  position: absolute;
  right: 4px;
  top: 100%;
  margin-top: 4px;
  background: var(--surface, #fff);
  border: 1px solid var(--border-color, #ddd);
  border-radius: 4px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.08);
  z-index: 5;
  min-width: 200px;
  padding: 4px;
}
.m-overflow__item {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
  padding: 6px 8px;
  font-size: 11px;
  background: transparent;
  border: 0;
  cursor: pointer;
  border-radius: 3px;
  color: var(--textPrimary, #222);
}
.m-overflow__item:hover { background: var(--alt-background, #f4f4f4); }
.m-overflow__item .material-icons { font-size: 14px; }
</style>
