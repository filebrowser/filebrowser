<template>
  <div class="part-viewer">
    <div ref="containerEl" class="part-viewer__canvas"></div>
    <div v-if="loading" class="part-viewer__overlay">
      <div class="part-viewer__hint">{{ t("machine.partLoading") }}</div>
    </div>
    <div v-else-if="error" class="part-viewer__overlay part-viewer__overlay--err">
      <div class="part-viewer__hint">{{ error }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
// Thin wrapper around Online3DViewer's EmbeddedViewer. Renders a 3D
// model from a URL into a canvas filling the parent element. Handles
// .stl / .step / .3mf / .obj / .x_t / .iges and friends out of the
// box — see modelExtensions in http/cnc.go for the matching server
// rule.
//
// We pull the engine module dynamically so the bundle cost (parsers
// + WASM for STEP) only lands when the user actually opens a 3D file.

import { onBeforeUnmount, ref, watch } from "vue";
import { useI18n } from "vue-i18n";

const props = defineProps<{
  url: string;
}>();

const { t } = useI18n();
const containerEl = ref<HTMLElement | null>(null);
const loading = ref(false);
const error = ref<string>("");
let viewer: { Destroy?: () => void } | null = null;

const teardown = () => {
  if (viewer && typeof viewer.Destroy === "function") {
    try {
      viewer.Destroy();
    } catch {
      /* ignore — best-effort */
    }
  }
  viewer = null;
  if (containerEl.value) {
    containerEl.value.innerHTML = "";
  }
};

const render = async (url: string) => {
  if (!containerEl.value || !url) return;
  teardown();
  loading.value = true;
  error.value = "";
  try {
    const ov = await import("online-3d-viewer");
    viewer = new ov.EmbeddedViewer(containerEl.value, {
      backgroundColor: new ov.RGBAColor(245, 245, 245, 255),
      defaultColor: new ov.RGBColor(160, 160, 160),
      onModelLoaded: () => {
        loading.value = false;
      },
    });
    (viewer as any).LoadModelFromUrlList([url]);
  } catch (e: any) {
    error.value = e?.message || "viewer load failed";
    loading.value = false;
  }
};

watch(
  () => props.url,
  (u) => {
    render(u);
  },
  { immediate: true }
);

onBeforeUnmount(() => {
  teardown();
});
</script>

<style>
.part-viewer {
  position: relative;
  width: 100%;
  height: 100%;
  background: #f5f5f5;
}

.part-viewer__canvas {
  position: absolute;
  inset: 0;
}

.part-viewer__canvas canvas {
  display: block;
  width: 100% !important;
  height: 100% !important;
}

.part-viewer__overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(245, 245, 245, 0.85);
  pointer-events: none;
}

.part-viewer__overlay--err {
  background: rgba(255, 235, 235, 0.95);
}

.part-viewer__hint {
  font-size: 0.85rem;
  color: var(--fg-muted, #888);
}

.part-viewer__overlay--err .part-viewer__hint {
  color: #c62828;
}
</style>
