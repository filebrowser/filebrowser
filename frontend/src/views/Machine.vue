<template>
  <div class="m-page" :class="{ 'm-page--kiosk': kioskMode }">
    <header-bar v-if="!kioskMode" showMenu showLogo>
      <title>{{ machineLabel }}</title>
    </header-bar>

    <div class="m-frame">
      <TopBar
        @open-connection="openConnection"
        @toggle-queue="toggleQueue"
      />

      <QueuePanel
        ref="queuePanelRef"
        :force-open="queueForceOpen"
        @stop-machine="promptStopMachine"
      />

      <HeroStateBar />

      <div class="m-main">
        <div class="m-main__left">
          <TabStrip
            :active="activeTab"
            :tool-mismatch-count="toolMismatchCount"
            :file-tabs="fileTabs"
            @select="onSelectTab"
          />

          <div class="m-main__body">
            <!-- G-code + 3D (default) -->
            <div v-show="activeTab === 'gcode'" class="m-main__gcode3d">
              <div class="m-pane m-pane--code">
                <div
                  v-if="cnc.attachedFile && !cnc.running"
                  class="m-pane__attached"
                  :class="{ 'm-pane__attached--auto': cnc.attachedSource === 'auto' }"
                >
                  <span class="m-pane__attached-label">
                    {{ cnc.attachedSource === 'auto'
                      ? t('machine.attachedBadgeAuto')
                      : t('machine.attachedBadge') }}
                  </span>
                  <span class="m-pane__attached-path">{{ cnc.attachedFile }}</span>
                  <button class="m-pane__attached-detach" @click="onDetach">
                    {{ t('machine.attachDetach') }}
                  </button>
                </div>
                <GcodeFollow
                  v-if="ncContent !== null"
                  :gcode="ncContent"
                  :machine-line="Number(cnc.lineCurrent) || 0"
                />
                <div v-else class="m-pane__hint">
                  {{ ncLoading ? t("machine.ncLoading") : t("machine.ncIdle") }}
                </div>
              </div>
              <div class="m-pane m-pane--viewer">
                <GCode3DViewer
                  v-if="ncContent !== null"
                  :gcode="ncContent"
                  :machine-line="Number(cnc.lineCurrent) || 0"
                />
                <div v-else class="m-pane__hint m-pane__hint--dark">
                  {{ t("machine.viewerIdle") }}
                </div>
              </div>
            </div>

            <!-- Tools tab -->
            <div v-show="activeTab === 'tools'" class="m-main__filltab">
              <ToolsTab :preflight="preflight" />
            </div>

            <!-- File tab -->
            <div v-if="isFileTab" class="m-main__filltab">
              <FilePreview :file-path="activeTab" />
            </div>
          </div>
        </div>

        <RightRail
          :axes="effectiveAxes"
          :position-tolerance="positionTolerance"
          :camera-u-r-l="cameraURL"
          :camera-type="cameraType"
          :line-current="Number(cnc.lineCurrent) || 0"
          :line-total="Number(cnc.lineTotal) || 0"
          :eta-ms="etaMs"
          @expand-camera="cameraExpanded = true"
        />
      </div>
    </div>

    <ConnectionModal
      v-if="connectionOpen"
      :initial-tab="connectionTab"
      @close="connectionOpen = false"
    />

    <div v-if="cameraExpanded" class="m-camera-overlay" @click="cameraExpanded = false">
      <iframe
        v-if="cameraType === 'iframe'"
        :src="cameraURL"
        class="m-camera-overlay__frame m-camera-overlay__frame--iframe"
        allow="autoplay; fullscreen; encrypted-media"
        referrerpolicy="no-referrer"
      />
      <video
        v-else-if="cameraURL.endsWith('.m3u8') || cameraType === 'hls'"
        :src="cameraURL"
        controls
        autoplay
        muted
        playsinline
        class="m-camera-overlay__frame"
      />
      <img
        v-else-if="cameraURL"
        :src="cameraURL"
        class="m-camera-overlay__frame"
        alt=""
      />
      <button class="m-camera-overlay__close" @click.stop="cameraExpanded = false">✕</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";
import { cnc as cncApi, files as filesApi } from "@/api";
import GCode3DViewer from "@/components/GCode3DViewer.vue";
import HeaderBar from "@/components/header/HeaderBar.vue";
import TopBar from "@/components/machine/TopBar.vue";
import HeroStateBar from "@/components/machine/HeroStateBar.vue";
import TabStrip from "@/components/machine/TabStrip.vue";
import type { FileTabSpec } from "@/components/machine/TabStrip.vue";
import GcodeFollow from "@/components/machine/GcodeFollow.vue";
import RightRail from "@/components/machine/RightRail.vue";
import QueuePanel from "@/components/machine/QueuePanel.vue";
import ConnectionModal from "@/components/machine/ConnectionModal.vue";
import ToolsTab from "@/components/machine/ToolsTab.vue";
import FilePreview from "@/components/machine/FilePreview.vue";
import { useCncStore } from "@/stores/cnc";
import { useLayoutStore } from "@/stores/layout";

const { t } = useI18n();
const cnc = useCncStore();
const layoutStore = useLayoutStore();
const route = useRoute();
const $showError = inject<IToastError>("$showError")!;

const kioskMode = computed(() => route.query.kiosk === "1");

const machineLabel = computed(
  () => cnc.currentMachine?.name || t("sidebar.machine")
);

// ── Connection modal ──────────────────────────────────────────────
const connectionOpen = ref(false);
const connectionTab = ref<"bridge" | "controller" | "activity">("bridge");
const openConnection = (tab: "bridge" | "controller" | "activity") => {
  connectionTab.value = tab;
  connectionOpen.value = true;
};

// ── Queue panel toggle (controlled here so other surfaces can expand) ──
const queuePanelRef = ref<InstanceType<typeof QueuePanel> | null>(null);
const queueForceOpen = ref(false);
const toggleQueue = () => {
  if (queuePanelRef.value) queuePanelRef.value.toggle();
};

// ?queue=open from the file browser's "Send to machine" flow.
watch(
  () => route.query.queue,
  (v) => {
    queueForceOpen.value = v === "open";
  },
  { immediate: true }
);

// ── Active tab. "gcode" | "tools" | "<file-path>" ──
const activeTab = ref<string>("gcode");
const isFileTab = computed(
  () => activeTab.value !== "gcode" && activeTab.value !== "tools"
);

const onSelectTab = (tab: string) => {
  activeTab.value = tab;
};

// ── Tool mismatch counter (drives "Tools ⚠ N") ──
// Re-uses the preflight result against the current NC file.
import type { Preflight } from "@/api/cnc";
const preflight = ref<Preflight | null>(null);
const toolMismatchCount = computed(() => {
  const p = preflight.value;
  if (!p) return 0;
  return p.summary.warn + p.summary.empty + p.summary.offline + p.summary.missing;
});

const refreshPreflight = async (path: string) => {
  if (!path) {
    preflight.value = null;
    return;
  }
  try {
    preflight.value = await cncApi.getPreflight(
      path,
      cnc.currentMachineId || undefined
    );
  } catch {
    preflight.value = null;
  }
};

// ── NC content + file tabs derived from the running file's folder ──
const ncContent = ref<string | null>(null);
const ncLoading = ref(false);
const fileTabs = ref<FileTabSpec[]>([]);

const fileIcon = (name: string): string => {
  const lower = name.toLowerCase();
  if (lower.endsWith(".pdf")) return "description";
  if (/\.(step|stp|3mf|obj|stl|x_t|x_b|iges|igs)$/i.test(lower))
    return "view_in_ar";
  if (lower.endsWith(".nc")) return "code";
  return "insert_drive_file";
};

const loadJobFolder = async (filePath: string) => {
  if (!filePath) {
    fileTabs.value = [];
    return;
  }
  const dir = filePath.split("/").slice(0, -1).join("/") || "/";
  try {
    const res = await filesApi.fetch(dir);
    const items = ((res as any).items || []) as { name: string; path: string }[];
    const cwd = (res as any).path || dir;
    fileTabs.value = items
      .filter((it) => {
        const n = it.name.toLowerCase();
        // Skip the NC itself (it's the G-code tab) and anything inside
        // hidden / cnc-tool-tables subfolders.
        if (n.endsWith(".nc") || n.endsWith(".tap") || n.endsWith(".ngc")) return false;
        return /\.(pdf|step|stp|3mf|obj|stl|x_t|x_b|iges|igs|png|jpg|jpeg|webp)$/i.test(n);
      })
      .map((it) => {
        const path = (cwd.endsWith("/") ? cwd : cwd + "/") + it.name;
        return {
          path,
          label: it.name,
          title: path,
          icon: fileIcon(it.name),
        };
      });
  } catch {
    fileTabs.value = [];
  }
};

const fetchNc = async (path: string) => {
  ncLoading.value = true;
  try {
    const res = await filesApi.fetch(path);
    ncContent.value = (res as any).content ?? "";
  } catch {
    ncContent.value = null;
  } finally {
    ncLoading.value = false;
  }
};

// Drive everything off effectiveFilePath (real job's file OR
// operator-attached file). Switches to/from attachment refresh the
// preview without a manual reload.
watch(
  () => cnc.effectiveFilePath,
  (p) => {
    if (p) {
      fetchNc(p);
      loadJobFolder(p);
      refreshPreflight(p);
    } else {
      ncContent.value = null;
      fileTabs.value = [];
      preflight.value = null;
    }
  },
  { immediate: false }
);

// ── ETA derived locally so the right rail can display "02:14 left" ──
const now = ref(Date.now());
let nowTimer: ReturnType<typeof setInterval> | null = null;
watch(
  () => cnc.running,
  (running) => {
    if (running && !nowTimer) {
      nowTimer = setInterval(() => (now.value = Date.now()), 1000);
    } else if (!running && nowTimer) {
      clearInterval(nowTimer);
      nowTimer = null;
    }
  },
  { immediate: true }
);
const etaMs = computed<number | null>(() => {
  if (!cnc.running) return null;
  const startedAt = cnc.raw?.started_at;
  if (!startedAt) return null;
  const elapsed = now.value - new Date(startedAt).getTime();
  if (elapsed < 1000) return null;
  const lps = (Number(cnc.lineCurrent) || 0) / (elapsed / 1000);
  if (!Number.isFinite(lps) || lps <= 0) return null;
  const remaining = Math.max(0, (Number(cnc.lineTotal) || 0) - (Number(cnc.lineCurrent) || 0));
  return (remaining / lps) * 1000;
});

// ── Camera config + axes resolution from current machine settings ──
const cameraURL = ref("");
const cameraType = ref<string>("auto");
const effectiveAxes = ref<string[]>(["X", "Y", "Z"]);
const positionTolerance = ref(0.001);
const cameraExpanded = ref(false);

const loadMachineCfg = async () => {
  try {
    const s = await cncApi.getSettings();
    const id = cnc.currentMachineId;
    const m =
      (id ? s.machines?.find((x) => x.id === id) : null) || s.machines?.[0];
    cameraURL.value = m?.cameraUrl ?? s.cameraUrl ?? "";
    cameraType.value = m?.cameraType ?? "auto";
    if (Array.isArray(m?.axesEnabled) && m?.axesEnabled.length > 0) {
      effectiveAxes.value = m.axesEnabled.map((a: string) => a.toUpperCase());
    } else {
      effectiveAxes.value = ["X", "Y", "Z"];
    }
    if (typeof m?.positionToleranceIn === "number" && m.positionToleranceIn > 0) {
      positionTolerance.value = m.positionToleranceIn;
    } else {
      positionTolerance.value = 0.001;
    }
  } catch {
    /* leave defaults; configure-me hint surfaces elsewhere */
  }
};

// ── Stop machine ──
const promptStopMachine = () => {
  layoutStore.showHover({
    prompt: "stopMachine",
    props: {
      filePath: cnc.filePath,
      lineCurrent: cnc.lineCurrent,
    },
    confirm: async (event: Event) => {
      event.preventDefault();
      layoutStore.closeHovers();
      try {
        await cncApi.stop(cnc.currentMachineId || undefined);
        cnc.pollOnce();
      } catch (e: any) {
        $showError(e);
      }
    },
  });
};

// ── Lifecycle ──
onMounted(async () => {
  // ?machine_id= overrides persisted selection on this tab.
  const requested = route.query.machine_id;
  if (typeof requested === "string" && requested) {
    await cnc.setCurrentMachine(requested);
  }
  await cnc.start();
  await loadMachineCfg();
  await cnc.seedMetrics();
  await cnc.loadQueue();
  const initial = cnc.effectiveFilePath;
  if (initial) {
    fetchNc(initial);
    loadJobFolder(initial);
    refreshPreflight(initial);
  }
});

watch(() => cnc.currentMachineId, async (id) => {
  if (id) {
    await loadMachineCfg();
    const eff = cnc.effectiveFilePath;
    if (eff) refreshPreflight(eff);
  }
});

// Detach handler — clears the attachment + emits a status broadcast
// server-side so any other open dashboard tabs reset too.
const onDetach = async () => {
  try {
    await cnc.detachFile();
  } catch {
    /* surface via store log if needed */
  }
};

onBeforeUnmount(() => {
  if (nowTimer) clearInterval(nowTimer);
});
</script>

<style scoped>
.m-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: var(--background, #f5f5f5);
  overflow: hidden;
}
.m-page--kiosk { background: transparent; }

.m-frame {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 5px;
  padding: 8px;
  min-height: 0;
  overflow: hidden;
}

.m-main {
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 220px;
  gap: 8px;
  min-height: 0;
}

.m-main__left {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-height: 0;
  min-width: 0;
}

.m-main__body {
  flex: 1;
  min-height: 0;
  display: flex;
}

.m-main__gcode3d {
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 6px;
  min-height: 0;
}

.m-pane {
  border-radius: 6px;
  overflow: hidden;
  min-height: 0;
  min-width: 0;
  display: flex;
  flex-direction: column;
}
.m-pane--code { background: #2C2C2A; }
.m-pane--viewer { background: #1a1a1a; }

.m-pane__hint {
  margin: auto;
  padding: 16px;
  color: var(--fg-muted, #888);
  font-size: 12px;
}
.m-pane__hint--dark { color: #888780; }

.m-pane__attached {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px;
  background: rgba(24, 95, 165, 0.16);
  color: #cfe6ff;
  font-size: 11px;
  border-bottom: 1px solid rgba(24, 95, 165, 0.4);
}
.m-pane__attached--auto {
  background: rgba(184, 81, 0, 0.18);
  color: #ffd5a8;
  border-bottom-color: rgba(184, 81, 0, 0.45);
}
.m-pane__attached-label {
  font-weight: 600;
  text-transform: uppercase;
  font-size: 10px;
  letter-spacing: 0.04em;
}
.m-pane__attached-path {
  flex: 1 1 0;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.m-pane__attached-detach {
  background: none;
  border: 1px solid currentColor;
  color: inherit;
  padding: 1px 6px;
  border-radius: 3px;
  font: inherit;
  cursor: pointer;
}
.m-pane__attached-detach:hover { background: rgba(255, 255, 255, 0.08); }

.m-main__filltab {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.m-camera-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.92);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
}
.m-camera-overlay__frame { max-width: 96vw; max-height: 92vh; }
.m-camera-overlay__frame--iframe { width: 96vw; height: 92vh; border: 0; }
.m-camera-overlay__close {
  position: absolute;
  top: 16px;
  right: 24px;
  font-size: 24px;
  background: transparent;
  border: 0;
  color: #fff;
  cursor: pointer;
}

/* ── Mobile / tablet ── */
@media (max-width: 1024px) {
  .m-main { grid-template-columns: 1fr; }
  .m-main__gcode3d { grid-template-columns: 1fr; }
}
</style>
