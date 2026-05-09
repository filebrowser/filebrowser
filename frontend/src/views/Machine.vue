<template>
  <div class="machine-page">
    <header-bar showMenu showLogo>
      <title>{{ t("sidebar.machine") }}</title>
    </header-bar>

    <div class="machine-grid">
      <!-- Camera tile (Z-13). HLS / snapshot / RTSP-hint. -->
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

      <!-- Native dashboard. Polls /api/cnc/state every 1s. -->
      <section class="machine-card dashboard-card">
        <div class="card-header">
          <i class="material-icons">precision_manufacturing</i>
          {{ t("machine.dashboardTitle") }}
          <span class="card-header__spacer" />
          <span class="card-header__hint" v-if="!hostConfigured">
            {{ t("machine.notConfigured") }}
          </span>
          <span class="card-header__hint" v-else-if="!anyFresh">
            {{ t("machine.waitingFirstPoll") }}
          </span>
        </div>
        <div class="card-body dashboard-body">
          <!-- Hero: program + status + mode -->
          <div class="hero">
            <div class="hero__program">
              <div class="hero__label">{{ t("machine.program") }}</div>
              <div class="hero__value">{{ programDisplay || "—" }}</div>
            </div>
            <div class="hero__status" :class="statusClass">
              {{ statusText || "—" }}
            </div>
            <div class="hero__mode">
              <div class="hero__label">{{ t("machine.mode") }}</div>
              <div class="hero__value">{{ rawValue("mode") || "—" }}</div>
            </div>
          </div>

          <!-- Tiles row: spindle / parts / tool / cycle -->
          <div class="tiles">
            <Tile
              :label="t('machine.spindleRpm')"
              :value="formatNum(parsed('spindle_actual'))"
              :sub="t('machine.spindleCmd', { n: formatNum(parsed('spindle_cmd'), 0) })"
              icon="rotate_right"
            />
            <Tile
              :label="t('machine.parts')"
              :value="formatNum(parsed('parts'), 0)"
              icon="inventory_2"
            />
            <Tile
              :label="t('machine.tool')"
              :value="formatNum(parsed('tool'), 0)"
              icon="construction"
            />
            <Tile
              :label="t('machine.lastCycle')"
              :value="rawValue('last_cycle') || '—'"
              icon="timer"
            />
          </div>

          <!-- Position grid: machine vs work -->
          <div class="positions">
            <div class="positions__col">
              <div class="positions__title">{{ t("machine.machinePos") }}</div>
              <Axis label="X" :value="parsed('pos_x')" />
              <Axis label="Y" :value="parsed('pos_y')" />
              <Axis label="Z" :value="parsed('pos_z')" />
            </div>
            <div class="positions__col">
              <div class="positions__title">{{ t("machine.workPos") }}</div>
              <Axis label="X" :value="parsed('work_x')" />
              <Axis label="Y" :value="parsed('work_y')" />
              <Axis label="Z" :value="parsed('work_z')" />
            </div>
            <div class="positions__col">
              <div class="positions__title">{{ t("machine.g54Offset") }}</div>
              <Axis label="X" :value="parsed('g54_x')" />
              <Axis label="Y" :value="parsed('g54_y')" />
              <Axis label="Z" :value="parsed('g54_z')" />
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, h, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { cnc as cncApi } from "@/api";
import type { CncMetric, CncStateSnapshot } from "@/api/cnc";
import HeaderBar from "@/components/header/HeaderBar.vue";

const { t } = useI18n();

// ── Config from /api/cnc/settings ──────────────────────────────────────────
const cameraURL = ref("");
const hostConfigured = ref(false);

// ── State snapshot from /api/cnc/state, polled every 1 s ───────────────────
const state = ref<CncStateSnapshot>({});
const STATE_POLL_MS = 1000;
let stateTimer: ReturnType<typeof setInterval> | null = null;

const refreshState = async () => {
  if (!hostConfigured.value) return;
  try {
    state.value = await cncApi.getState();
  } catch {
    /* leave previous values; the tile will go "stale" via the
       last_update timestamp the server sent. */
  }
};

const metric = (key: string): CncMetric | undefined => state.value[key];
const parsed = (key: string): unknown => metric(key)?.parsed ?? null;
const rawValue = (key: string): string => metric(key)?.value ?? "";

const formatNum = (v: unknown, digits = 1): string => {
  if (typeof v === "number" && Number.isFinite(v)) {
    return digits === 0 ? Math.round(v).toString() : v.toFixed(digits);
  }
  return "—";
};

// Q500 returns either a "PROGRAM,O123,RUNNING,PARTS,n" dict from the
// streamer's parser, or a plain string. Surface what we have.
const programDisplay = computed(() => {
  const p = parsed("status_combined");
  if (p && typeof p === "object" && "program" in (p as Record<string, unknown>)) {
    return (p as Record<string, string>).program;
  }
  return rawValue("status_combined");
});

const statusText = computed(() => {
  const p = parsed("status_combined");
  if (p && typeof p === "object" && "status" in (p as Record<string, unknown>)) {
    return (p as Record<string, string>).status;
  }
  return "";
});

const statusClass = computed(() => {
  const s = (statusText.value || "").toLowerCase();
  if (s.includes("run")) return "is-running";
  if (s.includes("hold") || s.includes("stop")) return "is-warn";
  if (s.includes("alarm") || s.includes("fault") || s.includes("error"))
    return "is-error";
  return "";
});

const anyFresh = computed(() =>
  Object.values(state.value).some((m) => m && !m.stale)
);

// ── Camera dispatch (same as before) ───────────────────────────────────────
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
  return "snapshot";
});

const snapshotSrc = computed(() => {
  if (!cameraURL.value) return "";
  const sep = cameraURL.value.includes("?") ? "&" : "?";
  return `${cameraURL.value}${sep}_t=${snapshotTick.value}`;
});

watch(cameraKind, (kind) => {
  if (snapshotTimer) {
    clearInterval(snapshotTimer);
    snapshotTimer = null;
  }
  if (kind === "snapshot") {
    snapshotTimer = setInterval(() => snapshotTick.value++, 200);
  }
});

// ── Lifecycle ──────────────────────────────────────────────────────────────
onMounted(async () => {
  try {
    const s = await cncApi.getSettings();
    cameraURL.value = s.cameraUrl || "";
    hostConfigured.value = !!s.haasHost;
  } catch {
    /* ignore — view renders the configure-me hints */
  }
  refreshState();
  stateTimer = setInterval(refreshState, STATE_POLL_MS);
  if (cameraKind.value === "snapshot" && !snapshotTimer) {
    snapshotTimer = setInterval(() => snapshotTick.value++, 200);
  }
});

onBeforeUnmount(() => {
  if (stateTimer) clearInterval(stateTimer);
  if (snapshotTimer) clearInterval(snapshotTimer);
});

// ── Inline mini-components ─────────────────────────────────────────────────
const Tile = (props: { label: string; value: string; sub?: string; icon: string }) =>
  h("div", { class: "tile" }, [
    h("i", { class: "material-icons tile__icon" }, props.icon),
    h("div", { class: "tile__body" }, [
      h("div", { class: "tile__label" }, props.label),
      h("div", { class: "tile__value" }, props.value),
      props.sub ? h("div", { class: "tile__sub" }, props.sub) : null,
    ]),
  ]);

const Axis = (props: { label: string; value: unknown }) => {
  let v = "—";
  if (typeof props.value === "number" && Number.isFinite(props.value)) {
    v = props.value.toFixed(4);
  } else if (typeof props.value === "string" && props.value !== "") {
    v = props.value;
  }
  return h("div", { class: "axis" }, [
    h("span", { class: "axis__label" }, props.label),
    h("span", { class: "axis__value" }, v),
  ]);
};
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

.card-header__spacer {
  flex: 1;
}

.card-header__hint {
  font-size: 0.78rem;
  font-weight: 400;
  color: var(--fg-muted, #888);
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

.dashboard-body {
  flex-direction: column;
  padding: 1rem;
  gap: 1rem;
  overflow: auto;
}

/* Hero row */
.hero {
  display: grid;
  grid-template-columns: 1fr auto auto;
  gap: 1rem;
  align-items: center;
  padding: 0.8rem 1rem;
  background: var(--alt-background, #fafafa);
  border: 1px solid var(--border-color, #eee);
  border-radius: 6px;
}

.hero__label {
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--fg-muted, #888);
}

.hero__value {
  font-size: 1.1rem;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

.hero__status {
  padding: 0.3rem 0.8rem;
  border-radius: 999px;
  font-size: 0.85rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  background: rgba(0, 0, 0, 0.06);
}

.hero__status.is-running {
  background: rgba(46, 160, 67, 0.18);
  color: #2ea043;
}

.hero__status.is-warn {
  background: rgba(201, 122, 0, 0.2);
  color: #c97a00;
}

.hero__status.is-error {
  background: rgba(220, 53, 69, 0.18);
  color: #dc3545;
}

/* Tiles row */
.tiles {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 0.6rem;
}

.tile {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.6rem 0.8rem;
  background: var(--alt-background, #fafafa);
  border: 1px solid var(--border-color, #eee);
  border-radius: 6px;
}

.tile__icon {
  font-size: 1.6rem;
  color: var(--primaryColor, #2196f3);
  opacity: 0.9;
}

.tile__label {
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--fg-muted, #888);
}

.tile__value {
  font-size: 1.3rem;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  line-height: 1.1;
}

.tile__sub {
  font-size: 0.75rem;
  color: var(--fg-muted, #888);
  font-variant-numeric: tabular-nums;
  margin-top: 0.15rem;
}

/* Position grid */
.positions {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 0.6rem;
}

.positions__col {
  background: var(--alt-background, #fafafa);
  border: 1px solid var(--border-color, #eee);
  border-radius: 6px;
  padding: 0.6rem 0.8rem;
}

.positions__title {
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--fg-muted, #888);
  margin-bottom: 0.4rem;
}

.axis {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  font-variant-numeric: tabular-nums;
  padding: 0.15rem 0;
}

.axis__label {
  font-weight: 600;
  color: var(--primaryColor, #2196f3);
  margin-right: 0.5rem;
}

.axis__value {
  font-size: 1.05rem;
  font-weight: 500;
}

.camera-frame {
  width: 100%;
  height: 100%;
  object-fit: contain;
  background: #000;
}

.hint {
  margin: auto;
  padding: 1rem 1.5rem;
  color: var(--fg-muted, #888);
  font-size: 0.9rem;
  text-align: center;
}
</style>
