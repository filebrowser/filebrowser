<template>
  <aside class="m-rail">
    <div class="m-card">
      <div class="m-card__label">{{ t("machine.progress") }}</div>
      <div class="m-progress-row">
        <span class="m-progress-num">{{ lineCurrent.toLocaleString() }} / {{ lineTotal.toLocaleString() }}</span>
        <span v-if="etaLabel" class="m-progress-eta">{{ etaLabel }}</span>
      </div>
      <div class="m-progress-bar">
        <div class="m-progress-fill" :style="{ width: pctWidth }" />
      </div>
    </div>

    <div class="m-card">
      <div class="m-card__label">{{ t("machine.positionLabel") }}</div>
      <div class="m-pos-table" :style="{ 'grid-template-rows': `auto repeat(${axes.length}, auto)` }">
        <div class="m-pos-th-axis"></div>
        <div class="m-pos-th">{{ t("machine.posMach") }}</div>
        <div class="m-pos-th">{{ t("machine.posWork") }}</div>
        <div class="m-pos-th">{{ t("machine.posDeltaCmd") }}</div>
        <template v-for="ax in axes" :key="ax">
          <div class="m-pos-axis">{{ ax }}</div>
          <div class="m-pos-val">{{ fmtAxis(parsed(`pos_${ax.toLowerCase()}`)) }}</div>
          <div class="m-pos-val">{{ fmtAxis(parsed(`work_${ax.toLowerCase()}`)) }}</div>
          <div class="m-pos-val" :class="deltaClass(ax)">{{ fmtAxis(deltaCmd(ax)) }}</div>
        </template>
      </div>
    </div>

    <div class="m-camera">
      <span v-if="cameraConfigured" class="m-camera__live">● {{ t("machine.cameraLive") }}</span>
      <video
        v-if="cameraKind === 'hls'"
        :src="cameraURL"
        controls
        autoplay
        muted
        playsinline
        class="m-camera__frame"
      />
      <img
        v-else-if="cameraKind === 'snapshot'"
        :src="snapshotSrc"
        class="m-camera__frame"
        alt=""
      />
      <iframe
        v-else-if="cameraKind === 'iframe'"
        :src="cameraURL"
        class="m-camera__frame m-camera__frame--iframe"
        allow="autoplay; fullscreen; encrypted-media"
        referrerpolicy="no-referrer"
      />
      <div v-else-if="cameraKind === 'rtsp'" class="m-camera__hint">
        {{ t("machine.rtspNotSupported") }}
      </div>
      <div v-else class="m-camera__hint">{{ t("machine.cameraNone") }}</div>
      <button v-if="cameraConfigured" class="m-camera__expand" @click="$emit('expand-camera')" :title="t('machine.cameraExpand')">⛶</button>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useCncStore } from "@/stores/cnc";

const { t } = useI18n();
const cnc = useCncStore();

const props = defineProps<{
  axes: string[];
  positionTolerance: number; // inches
  cameraURL: string;
  cameraType: string;
  lineCurrent: number;
  lineTotal: number;
  etaMs: number | null;
}>();

defineEmits<{
  (e: "expand-camera"): void;
}>();

const metric = (key: string) => cnc.metrics[key];
const parsed = (key: string): unknown => metric(key)?.parsed ?? null;

const fmtAxis = (v: unknown): string => {
  if (typeof v === "number" && Number.isFinite(v)) return v.toFixed(4);
  if (typeof v === "string" && v !== "") return v;
  return "—";
};

const deltaCmd = (ax: string): number | null => {
  const mp = parsed(`pos_${ax.toLowerCase()}`);
  // No commanded-position metric on the wire today; the field is
  // reserved for when Q-code spec lands. For now display 0.0000 when
  // the machine position itself reads cleanly, "—" otherwise.
  if (typeof mp !== "number" || !Number.isFinite(mp)) return null;
  return 0;
};

const deltaClass = (ax: string) => {
  const v = deltaCmd(ax);
  if (v === null) return "m-pos-val--unknown";
  if (Math.abs(v) > props.positionTolerance) return "m-pos-val--warn";
  return "m-pos-val--ok";
};

const pctWidth = computed(() => {
  if (props.lineTotal <= 0) return "0%";
  const pct = Math.min(100, (props.lineCurrent / props.lineTotal) * 100);
  return `${pct}%`;
});

const fmtDuration = (ms: number) => {
  const s = Math.floor(ms / 1000);
  const m = Math.floor(s / 60);
  const sec = s % 60;
  if (m >= 60) {
    const h = Math.floor(m / 60);
    return `${h}h ${String(m % 60).padStart(2, "0")}m`;
  }
  return `${m}:${String(sec).padStart(2, "0")} ${t("machine.etaLeft")}`;
};
const etaLabel = computed(() => (props.etaMs && props.etaMs > 0 ? fmtDuration(props.etaMs) : ""));

// ── Camera ──
const cameraConfigured = computed(() => !!props.cameraURL && props.cameraType !== "none");
const snapshotTick = ref(0);
let snapshotTimer: ReturnType<typeof setInterval> | null = null;

const cameraKind = computed<"none" | "hls" | "snapshot" | "iframe" | "rtsp">(() => {
  const u = props.cameraURL;
  if (!u) return "none";
  switch (props.cameraType) {
    case "none": return "none";
    case "hls": return "hls";
    case "mjpeg": return "snapshot";
    case "iframe": return "iframe";
  }
  if (u.startsWith("rtsp://") || u.startsWith("rtsps://")) return "rtsp";
  if (u.endsWith(".m3u8")) return "hls";
  return "snapshot";
});

const snapshotSrc = computed(() => {
  if (!props.cameraURL) return "";
  const sep = props.cameraURL.includes("?") ? "&" : "?";
  return `${props.cameraURL}${sep}_t=${snapshotTick.value}`;
});

watch(cameraKind, (kind) => {
  if (snapshotTimer) { clearInterval(snapshotTimer); snapshotTimer = null; }
  if (kind === "snapshot") snapshotTimer = setInterval(() => snapshotTick.value++, 200);
}, { immediate: true });

onBeforeUnmount(() => { if (snapshotTimer) clearInterval(snapshotTimer); });
</script>

<style scoped>
.m-rail {
  display: flex;
  flex-direction: column;
  gap: 5px;
  min-height: 0;
  min-width: 0;
}
.m-card {
  background: var(--alt-background, #fafafa);
  border-radius: 6px;
  padding: 10px 12px;
  flex-shrink: 0;
  border: 1px solid var(--border-color, #eee);
}
.m-card__label {
  font-size: 10px;
  color: var(--fg-muted, #888);
  letter-spacing: 0.5px;
  font-weight: 500;
  text-transform: uppercase;
}
.m-progress-row {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-top: 4px;
}
.m-progress-num {
  font-size: 16px;
  color: var(--textPrimary, #222);
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}
.m-progress-eta { font-size: 11px; color: var(--fg-muted, #888); }
.m-progress-bar {
  height: 4px;
  background: var(--border-color, #e2e2e2);
  border-radius: 2px;
  overflow: hidden;
  margin-top: 6px;
}
.m-progress-fill {
  height: 100%;
  background: #185FA5;
  transition: width 0.25s ease;
}

.m-pos-table {
  display: grid;
  /* Position is the operator's primary read while standing at the
     machine — make it large enough to glance from a few feet away. */
  grid-template-columns: 22px 1fr 1fr 1fr;
  gap: 4px 8px;
  font-variant-numeric: tabular-nums;
  margin-top: 6px;
}
.m-pos-th {
  font-size: 10px;
  color: var(--fg-muted, #888);
  font-weight: 500;
  text-align: right;
  letter-spacing: 0.3px;
}
.m-pos-th-axis { font-size: 10px; }
.m-pos-axis { font-size: 18px; color: #185FA5; font-weight: 600; }
.m-pos-val { font-size: 18px; color: var(--textPrimary, #222); text-align: right; font-weight: 500; }
.m-pos-val--ok { color: #639922; }
.m-pos-val--warn { color: #BA7517; }
.m-pos-val--unknown { color: var(--fg-muted, #888); }

.m-camera {
  background: #2C2C2A;
  border-radius: 6px;
  flex: 1 1 0;
  min-height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #888780;
  font-size: 10px;
  position: relative;
  overflow: hidden;
}
.m-camera__frame {
  width: 100%;
  height: 100%;
  object-fit: contain;
  background: #000;
}
.m-camera__frame--iframe { border: 0; display: block; }
.m-camera__live {
  position: absolute;
  top: 4px;
  right: 4px;
  font-size: 9px;
  color: #C0DD97;
  z-index: 2;
}
.m-camera__expand {
  position: absolute;
  bottom: 4px;
  right: 4px;
  font-size: 12px;
  color: #888780;
  background: transparent;
  border: 0;
  cursor: pointer;
  z-index: 2;
}
.m-camera__hint {
  padding: 8px;
  text-align: center;
  font-size: 10px;
  color: #888780;
}
</style>
