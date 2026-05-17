<template>
  <div class="m-chapters" v-if="chapters.length > 0" :style="topStyle">
    <button
      class="m-chapters__btn"
      :title="t('machine.chaptersTitle', { n: chapters.length })"
      @click.stop="open = !open"
    >
      ☰ {{ t("machine.chapters", { n: chapters.length }) }}
    </button>
    <div v-if="currentLabel" class="m-chapters__current" :title="t('machine.chaptersCurrentTitle')">
      {{ currentLabel }}
    </div>
    <div v-if="open" class="m-chapters__menu" @click.stop>
      <ol class="m-chapters__list">
        <li
          v-for="c in chapters"
          :key="c.line"
          class="m-chapters__item"
          :class="{ 'm-chapters__item--current': current?.line === c.line }"
          @click="onJump(c.line)"
        >
          <span class="m-chapters__line">L{{ c.line }}</span>
          <span class="m-chapters__text">{{ c.comment }}</span>
        </li>
      </ol>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useI18n } from "vue-i18n";
import type { Chapter } from "@/api/cnc";

const { t } = useI18n();

const props = defineProps<{
  chapters: Chapter[];
  currentLine: number;
  // Pixel offset from the top of the parent pane. Machine.vue bumps
  // this up when the "Attached as current" banner is visible so the
  // toolpath toggle doesn't collide with it.
  topOffset?: number;
}>();

const topStyle = computed(() => ({
  top: `${props.topOffset ?? 6}px`,
}));

const emit = defineEmits<{ (e: "jump", line: number): void }>();

const open = ref(false);

// Find the chapter whose line ≤ currentLine, with the highest line.
// Equivalent to the Go-side ChapterAt; replicated here so a stale
// network call doesn't gate the live indicator.
const current = computed<Chapter | null>(() => {
  if (!props.currentLine || props.chapters.length === 0) return null;
  let match: Chapter | null = null;
  for (const c of props.chapters) {
    if (c.line <= props.currentLine) {
      match = c;
    } else {
      break;
    }
  }
  return match;
});

const currentLabel = computed(() => {
  const c = current.value;
  if (!c) return "";
  return t("machine.chaptersCurrent", { name: c.comment });
});

const onJump = (line: number) => {
  open.value = false;
  emit("jump", line);
};
</script>

<style scoped>
.m-chapters {
  position: absolute;
  left: 6px;
  z-index: 2;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
  pointer-events: none;
}
.m-chapters > * { pointer-events: auto; }

.m-chapters__btn {
  background: rgba(212, 209, 199, 0.1);
  color: #B4B2A9;
  border: 1px solid rgba(180, 178, 169, 0.25);
  font: inherit;
  font-size: 10px;
  padding: 2px 8px;
  border-radius: 3px;
  cursor: pointer;
}
.m-chapters__btn:hover { background: rgba(212, 209, 199, 0.18); }

.m-chapters__current {
  background: rgba(24, 95, 165, 0.22);
  color: #cfe6ff;
  font-size: 10px;
  padding: 2px 8px;
  border-radius: 3px;
  max-width: 360px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.m-chapters__menu {
  background: #2C2C2A;
  color: #B4B2A9;
  border: 1px solid rgba(180, 178, 169, 0.25);
  border-radius: 4px;
  box-shadow: 0 6px 18px rgba(0, 0, 0, 0.4);
  min-width: 280px;
  max-width: 480px;
  max-height: 60vh;
  overflow-y: auto;
  overscroll-behavior: contain;
}
.m-chapters__list {
  list-style: none;
  margin: 0;
  padding: 4px 0;
  font-size: 11px;
}
.m-chapters__item {
  display: grid;
  grid-template-columns: 3.5rem 1fr;
  gap: 6px;
  padding: 3px 8px;
  cursor: pointer;
}
.m-chapters__item:hover { background: rgba(212, 209, 199, 0.08); }
.m-chapters__item--current {
  background: rgba(24, 95, 165, 0.22);
  color: #cfe6ff;
}
.m-chapters__line {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  color: var(--fg-muted, #888780);
}
.m-chapters__text {
  white-space: normal;
  word-break: break-word;
}
</style>
