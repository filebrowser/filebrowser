<template>
  <div class="m-file-preview">
    <iframe
      v-if="kind === 'pdf'"
      :src="rawURL"
      class="m-file-preview__iframe"
    />
    <Part3DViewer
      v-else-if="kind === 'model'"
      :url="rawURL"
    />
    <div v-else class="m-file-preview__hint">
      {{ t("machine.filePreviewUnsupported") }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { baseURL } from "@/utils/constants";
import { useAuthStore } from "@/stores/auth";
import Part3DViewer from "@/components/Part3DViewer.vue";

const { t } = useI18n();
const authStore = useAuthStore();

const props = defineProps<{ filePath: string }>();

const ext = computed(() => {
  const i = props.filePath.lastIndexOf(".");
  return i >= 0 ? props.filePath.slice(i + 1).toLowerCase() : "";
});

const kind = computed<"pdf" | "model" | "other">(() => {
  if (ext.value === "pdf") return "pdf";
  if (["step", "stp", "3mf", "obj", "stl", "x_t", "x_b", "iges", "igs"].includes(ext.value))
    return "model";
  return "other";
});

const rawURL = computed(() => {
  // /api/raw renders the file inline; token-bearer auth.
  return `${baseURL}/api/raw${props.filePath}?auth=${encodeURIComponent(authStore.jwt)}&inline=true`;
});
</script>

<style scoped>
.m-file-preview {
  height: 100%;
  min-height: 0;
  background: var(--surface, #fff);
  border: 1px solid var(--border-color, #eee);
  border-radius: 6px;
  overflow: hidden;
}
.m-file-preview__iframe {
  width: 100%;
  height: 100%;
  border: 0;
}
.m-file-preview__hint {
  padding: 20px;
  text-align: center;
  color: var(--fg-muted, #888);
  font-size: 12px;
}
</style>
