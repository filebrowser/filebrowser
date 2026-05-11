<template>
  <span class="m-codeann">
    <span class="m-codeann__text">{{ msg }}</span>
    <span
      v-for="(a, i) in annotations"
      :key="i"
      class="m-codeann__pill"
      :title="a.summary + (a.hint ? '\n\n' + a.hint : '')"
    >
      <span class="m-codeann__pill-kind">{{ a.kind }}</span>
      <span class="m-codeann__pill-num">{{ a.number }}</span>
      <span class="m-codeann__pill-title">{{ a.title }}</span>
    </span>
  </span>
</template>

<script setup lang="ts">
import { ref, watch } from "vue";
import {
  extractCodeRefs,
  resolveAllCodeRefs,
} from "@/utils/codeRefs";

interface Annotation {
  kind: string;
  number: number;
  title: string;
  summary: string;
  hint?: string;
}

const props = defineProps<{ msg: string }>();
const annotations = ref<Annotation[]>([]);

const refresh = async (msg: string) => {
  annotations.value = [];
  const refs = extractCodeRefs(msg);
  if (refs.length === 0) return;
  const entries = await resolveAllCodeRefs(refs);
  // Bail if the prop changed under us during the async lookup.
  if (msg !== props.msg) return;
  const out: Annotation[] = [];
  refs.forEach((r, i) => {
    const e = entries[i];
    if (!e) return;
    out.push({
      kind: r.kind,
      number: r.number,
      title: e.title,
      summary: e.summary,
      hint: e.hint,
    });
  });
  annotations.value = out;
};

watch(
  () => props.msg,
  (v) => {
    refresh(v);
  },
  { immediate: true }
);
</script>

<style scoped>
.m-codeann {
  display: inline;
}
.m-codeann__text {
  white-space: pre-wrap;
}
.m-codeann__pill {
  display: inline-flex;
  align-items: baseline;
  margin-left: 6px;
  padding: 0 6px;
  border-radius: 3px;
  background: rgba(24, 95, 165, 0.08);
  border: 1px solid rgba(24, 95, 165, 0.18);
  font-size: 11px;
  line-height: 1.5;
  cursor: help;
  white-space: nowrap;
}
.m-codeann__pill-kind {
  text-transform: lowercase;
  color: #185fa5;
  margin-right: 4px;
}
.m-codeann__pill-num {
  font-weight: 600;
  margin-right: 6px;
}
.m-codeann__pill-title {
  color: var(--fg-muted, #555);
  max-width: 320px;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
