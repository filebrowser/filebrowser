<template>
  <div class="machine-page">
    <header-bar showMenu showLogo>
      <title>{{ t("sidebar.machine") }}</title>
    </header-bar>

    <div class="machine-grid">
      <!-- Camera tile (Z-13). Picks renderer by URL: HLS (.m3u8) →
           native <video>; snapshot/MJPEG (.jpg, /snapshot) → polled
           <img>; rtsp:// → render a hint instead of trying. -->
      <section class="machine-card camera-card" v-if="cameraKind !== 'none'">
        <div class="card-header">
          <i class="material-icons">videocam</i>
          {{ t("machine.cameraTitle") }}
        </div>
        <div class="card-body">
          <div v-if="cameraKind === 'rtsp'" class="hint">
            {{ t("machine.rtspNotSupported") }}
          </div>
          <video
            v-else-if="cameraKind === 'hls'"
            :src="cameraURL"
            controls
            autoplay
            muted
            playsinline
            class="camera-frame"
          />
          <img
            v-else-if="cameraKind === 'snapshot'"
            :src="snapshotSrc"
            class="camera-frame"
            alt=""
          />
        </div>
      </section>

      <!-- Dashboard iframe (Z-11). Stamps the machine token into the
           query string so haas-dashboard can authenticate the embed
           without forcing a separate login. The dashboard's D-4 work
           is what reads ?token= and trusts it. -->
      <section class="machine-card dashboard-card">
        <div class="card-header">
          <i class="material-icons">dashboard</i>
          {{ t("machine.dashboardTitle") }}
        </div>
        <div class="card-body">
          <iframe
            v-if="dashboardURL"
            :src="dashboardURL"
            class="dashboard-frame"
            allow="autoplay; fullscreen"
            referrerpolicy="no-referrer"
          />
          <div v-else class="hint">
            {{ t("machine.dashboardNotConfigured") }}
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { cnc as cncApi } from "@/api";
import HeaderBar from "@/components/header/HeaderBar.vue";

const { t } = useI18n();

const haasDashboardUrl = ref("");
const cameraURL = ref("");
const machineToken = ref("");

// MJPEG-style snapshots get a cache-busting timestamp every 200 ms so
// browsers actually re-request the frame. Real MJPEG streams (Content-
// Type: multipart/x-mixed-replace) handle themselves; the polling here
// is a fallback for cameras that only expose still-image endpoints.
const snapshotTick = ref(0);
let snapshotTimer: ReturnType<typeof setInterval> | null = null;

const cameraKind = computed<"none" | "hls" | "snapshot" | "rtsp">(() => {
  const u = cameraURL.value;
  if (!u) return "none";
  if (u.startsWith("rtsp://")) return "rtsp";
  if (u.endsWith(".m3u8")) return "hls";
  if (
    u.endsWith(".jpg") ||
    u.endsWith(".jpeg") ||
    u.endsWith("/snapshot") ||
    u.includes("/snapshot?") ||
    u.includes("snapshot=")
  ) {
    return "snapshot";
  }
  // Unknown — assume snapshot; a real MJPEG stream will keep the
  // <img> open and re-render frames as they arrive.
  return "snapshot";
});

const snapshotSrc = computed(() => {
  if (!cameraURL.value) return "";
  const sep = cameraURL.value.includes("?") ? "&" : "?";
  return `${cameraURL.value}${sep}_t=${snapshotTick.value}`;
});

const dashboardURL = computed(() => {
  if (!haasDashboardUrl.value) return "";
  const sep = haasDashboardUrl.value.includes("?") ? "&" : "?";
  return machineToken.value
    ? `${haasDashboardUrl.value}${sep}token=${encodeURIComponent(
        machineToken.value
      )}`
    : haasDashboardUrl.value;
});

watch(cameraKind, (kind) => {
  if (snapshotTimer) {
    clearInterval(snapshotTimer);
    snapshotTimer = null;
  }
  if (kind === "snapshot") {
    snapshotTimer = setInterval(() => {
      snapshotTick.value++;
    }, 200);
  }
});

onMounted(async () => {
  try {
    const s = await cncApi.getSettings();
    haasDashboardUrl.value = s.haasDashboardUrl || "";
    cameraURL.value = s.cameraUrl || "";
    machineToken.value = s.machineToken || "";
  } catch {
    /* leave fields empty; the view renders the configure-me hints */
  }
  // Kick the snapshot poll on first mount if the URL was already
  // snapshot-shaped (watcher only fires on changes after this point).
  if (cameraKind.value === "snapshot" && !snapshotTimer) {
    snapshotTimer = setInterval(() => {
      snapshotTick.value++;
    }, 200);
  }
});
</script>

<style scoped>
.machine-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: var(--background, #f5f5f5);
}

.machine-grid {
  flex: 1;
  display: grid;
  grid-template-columns: minmax(320px, 1fr) 2fr;
  gap: 1rem;
  padding: 1rem;
  min-height: 0;
}

@media (max-width: 900px) {
  .machine-grid {
    grid-template-columns: 1fr;
    grid-template-rows: auto 1fr;
  }
}

.machine-card {
  display: flex;
  flex-direction: column;
  background: var(--surface, #fff);
  border: 1px solid var(--border-color, #ddd);
  border-radius: 6px;
  overflow: hidden;
  min-height: 0;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.6rem 0.9rem;
  background: var(--alt-background, #fafafa);
  border-bottom: 1px solid var(--border-color, #eee);
  font-weight: 500;
  font-size: 0.95rem;
}

.card-header .material-icons {
  font-size: 1.2rem;
}

.card-body {
  flex: 1;
  display: flex;
  align-items: stretch;
  justify-content: stretch;
  min-height: 0;
  overflow: hidden;
}

.camera-frame {
  width: 100%;
  height: 100%;
  object-fit: contain;
  background: #000;
}

.dashboard-frame {
  width: 100%;
  height: 100%;
  border: none;
  background: #fff;
}

.hint {
  margin: auto;
  padding: 1rem 1.5rem;
  color: var(--fg-muted, #888);
  font-size: 0.9rem;
  text-align: center;
}
</style>
