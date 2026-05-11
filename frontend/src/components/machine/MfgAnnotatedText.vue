<template>
  <span class="m-mfgann">
    <template v-for="(seg, i) in segments" :key="i">
      <a
        v-if="seg.type === 'ref'"
        :href="seg.ref.url"
        class="m-mfgann__link"
        target="_blank"
        rel="noopener noreferrer"
        :title="`Search for ${seg.ref.vendor} ${seg.ref.partNumber}`"
        @click.stop
      >{{ seg.ref.matchedText }}</a>
      <span v-else>{{ seg.text }}</span>
    </template>
  </span>
</template>

<script setup lang="ts">
import { computed } from "vue";
import {
  extractMfgRefs,
  splitTextAroundRefs,
  type Segment,
} from "@/utils/mfgRefs";

const props = defineProps<{ text: string }>();

const segments = computed<Segment[]>(() => {
  const refs = extractMfgRefs(props.text);
  return splitTextAroundRefs(props.text, refs);
});
</script>

<style scoped>
.m-mfgann {
  display: inline;
  white-space: pre-wrap;
}
.m-mfgann__link {
  color: var(--primaryColor, #2196f3);
  text-decoration: underline;
  text-decoration-style: dotted;
  text-underline-offset: 2px;
}
.m-mfgann__link:hover {
  text-decoration-style: solid;
}
</style>
