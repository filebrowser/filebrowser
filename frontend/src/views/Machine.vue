<template>
  <div class="machine-page" :class="{ 'machine-page--kiosk': kioskMode }">
    <header-bar v-if="!kioskMode" showMenu showLogo>
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
          <iframe
            v-else-if="cameraKind === 'iframe'"
            :src="cameraURL"
            class="camera-frame camera-frame--iframe"
            allow="autoplay; fullscreen; encrypted-media"
            referrerpolicy="no-referrer"
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
          <span
            class="card-header__hint"
            :class="{ 'card-header__hint--err': connectTimedOut }"
            v-else-if="!anyFresh"
          >
            {{ connectTimedOut ? t("machine.connectTimeout") : t("machine.waitingFirstPoll") }}
          </span>
          <button
            v-if="!kioskMode"
            class="check-btn"
            :disabled="checking || cncStore.running"
            @click="runConnectionCheck"
            :title="t('machine.checkConnection')"
          >
            <i class="material-icons">network_check</i>
            {{ checking ? t("machine.checking") : t("machine.checkConnection") }}
          </button>
          <button
            v-if="cncStore.running && canModify && !kioskMode"
            class="stop-btn"
            @click="promptStopMachine"
            :title="t('buttons.stopMachine')"
          >
            <i class="material-icons">stop_circle</i>
            {{ t("buttons.stopMachine") }}
          </button>
        </div>
        <div class="card-body dashboard-body">
          <!-- Connection-check result: rendered briefly after the button -->
          <div v-if="checkResult" class="check-result" :class="checkResultClass">
            <div class="check-result__row">
              <i class="material-icons">{{ checkResult.bridge.ok ? "check_circle" : "error" }}</i>
              <span class="check-result__label">{{ t("machine.checkBridge") }}</span>
              <span class="check-result__detail">
                <template v-if="checkResult.bridge.ok">
                  {{ checkResult.bridge.address }} · {{ formatNum(checkResult.bridge.latency_ms, 0) }} ms
                </template>
                <template v-else>{{ checkResult.bridge.error || "?" }}</template>
              </span>
            </div>
            <div class="check-result__row">
              <i class="material-icons">{{ checkResult.controller.ok ? "check_circle" : "error" }}</i>
              <span class="check-result__label">{{ t("machine.checkController") }}</span>
              <span class="check-result__detail">
                <template v-if="checkResult.controller.ok">
                  Q104 → {{ checkResult.controller.mode }} · {{ formatNum(checkResult.controller.latency_ms, 0) }} ms
                </template>
                <template v-else>{{ checkResult.controller.error || "?" }}</template>
              </span>
            </div>
          </div>
          <!-- Live send progress (only while a job is streaming) -->
          <div v-if="cncStore.running" class="send-progress">
            <div class="send-progress__head">
              <span class="send-progress__file">{{ cncStore.filePath || "—" }}</span>
              <span class="send-progress__counter">
                {{ cncStore.lineCurrent }} / {{ cncStore.lineTotal }}
                <span v-if="cncStore.lineTotal > 0" class="send-progress__pct">
                  ({{ ((cncStore.lineCurrent / cncStore.lineTotal) * 100).toFixed(1) }}%)
                </span>
              </span>
            </div>
            <div class="send-progress__bar">
              <div
                class="send-progress__bar-fill"
                :style="{ width: cncStore.lineTotal > 0 ? `${(cncStore.lineCurrent / cncStore.lineTotal) * 100}%` : '0%' }"
              />
            </div>
            <div class="send-progress__time">
              <span>
                <i class="material-icons">schedule</i>
                {{ t("machine.elapsed") }}: {{ fmtDuration(elapsedMs) }}
              </span>
              <span v-if="etaMs !== null">
                <i class="material-icons">hourglass_bottom</i>
                {{ t("machine.eta") }}: {{ fmtDuration(etaMs) }}
              </span>
              <span v-if="linesPerSec !== null">
                <i class="material-icons">speed</i>
                {{ formatNum(linesPerSec, 1) }} {{ t("machine.linesPerSec") }}
              </span>
            </div>
          </div>

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

          <!-- Activity feed: backend log events + status transitions -->
          <div v-if="!kioskMode" class="activity">
            <div class="activity__title">
              {{ t("machine.activity") }}
              <button
                v-if="cncStore.log.length > 0"
                class="activity__clear"
                @click="cncStore.clearLog()"
                :title="t('machine.activityClear')"
              >
                {{ t("machine.activityClear") }}
              </button>
            </div>
            <div v-if="cncStore.log.length === 0" class="activity__empty">
              {{ t("machine.activityEmpty") }}
            </div>
            <ol v-else class="activity__list">
              <li
                v-for="(entry, i) in cncStore.log"
                :key="i"
                class="activity__row"
                :class="`activity__row--${entry.level}`"
              >
                <span class="activity__ts">{{ fmtTs(entry.ts) }}</span>
                <span class="activity__level">{{ entry.level }}</span>
                <span class="activity__msg">{{ entry.msg }}</span>
              </li>
            </ol>
          </div>
        </div>
      </section>

      <!-- Tool-table readout — manual refresh, persisted as JSON
           dumps under /cnc-tool-tables/<machine-id>/ in the user share
           so history is browsable in the file UI. -->
      <ToolTablePanel
        v-if="!kioskMode"
        :machine-id="undefined"
        :cnc-running="cncStore.running"
      />

      <!-- NC code + toolpath, side-by-side. Only when a job has a
           file path (running, or just ended and we still have it). -->
      <section v-if="cncStore.filePath" class="machine-card nc-card">
        <div class="card-header">
          <i class="material-icons">code</i>
          {{ cncStore.filePath }}
          <span class="card-header__spacer" />
          <a
            v-if="siblings.drawing_url"
            class="sibling-link"
            :href="siblings.drawing_url"
            target="_blank"
            rel="noopener"
            :title="siblings.drawing_name"
          >
            <i class="material-icons">description</i>
            {{ t("machine.drawing") }}
          </a>
          <router-link
            v-if="siblings.model_path"
            class="sibling-link"
            :to="`/files${siblings.model_path}`"
            :title="siblings.model_name"
          >
            <i class="material-icons">view_in_ar</i>
            {{ t("machine.model") }}
          </router-link>
          <span v-if="ncLoading" class="card-header__hint">{{ t("machine.ncLoading") }}</span>
          <span v-else-if="ncError" class="card-header__hint card-header__hint--err">{{ ncError }}</span>
        </div>
        <div class="nc-split" :style="{ '--nc-pct': ncSplitPct + '%' }">
          <div class="nc-split__pane nc-split__pane--code">
            <MachineNcMirror
              v-if="ncContent !== null"
              :gcode="ncContent"
              :machine-line="cncStore.lineCurrent"
            />
          </div>
          <div
            class="nc-split__bar"
            role="separator"
            aria-orientation="vertical"
            @pointerdown="startNcResize"
            @dblclick="resetNcSplit"
            :title="t('buttons.dragToResize') || 'Drag to resize · double-click to reset'"
          ></div>
          <div class="nc-split__pane nc-split__pane--viewer">
            <GCode3DViewer
              v-if="ncContent !== null"
              :gcode="ncContent"
              :machine-line="cncStore.lineCurrent"
            />
          </div>
        </div>
      </section>

      <!-- 3D part viewer below the NC card when a sibling model exists -->
      <section v-if="cncStore.filePath && siblings.model_url" class="machine-card part-card">
        <div class="card-header">
          <i class="material-icons">view_in_ar</i>
          {{ siblings.model_name }}
        </div>
        <div class="part-card__body">
          <Part3DViewer :url="siblings.model_url" />
        </div>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, h, inject, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { cnc as cncApi, files as filesApi } from "@/api";
import type { CncCheckResult, CncMetric } from "@/api/cnc";
import GCode3DViewer from "@/components/GCode3DViewer.vue";
import MachineNcMirror from "@/components/MachineNcMirror.vue";
import Part3DViewer from "@/components/Part3DViewer.vue";
import ToolTablePanel from "@/components/ToolTablePanel.vue";
import { useAuthStore } from "@/stores/auth";
import { useCncStore } from "@/stores/cnc";
import { useLayoutStore } from "@/stores/layout";
import { useRoute } from "vue-router";
import HeaderBar from "@/components/header/HeaderBar.vue";

const { t } = useI18n();
const $showError = inject<IToastError>("$showError")!;
const authStore = useAuthStore();
const layoutStore = useLayoutStore();
const route = useRoute();
const canModify = computed(() => !!authStore.user?.perm.modify);

// /machine?kiosk=1 strips the global header + activity log + check
// button so the page can be iframed cleanly into Home Assistant
// Lovelace WebPage cards or hung on a shop tablet. Same data, no
// chrome.
const kioskMode = computed(() => route.query.kiosk === "1");

const promptStopMachine = () => {
  layoutStore.showHover({
    prompt: "stopMachine",
    props: {
      filePath: cncStore.filePath,
      lineCurrent: cncStore.lineCurrent,
    },
    confirm: async (event: Event) => {
      event.preventDefault();
      layoutStore.closeHovers();
      try {
        await cncApi.stop();
        cncStore.pollOnce();
      } catch (e: any) {
        $showError(e);
      }
    },
  });
};

// ── Config from /api/cnc/settings ──────────────────────────────────────────
const cameraURL = ref("");
const cameraType = ref<string>("auto");
const hostConfigured = ref(false);

// ── File ↔ NC sibling discovery (model + drawing) ─────────────────────────
// Looks up a 3D model + PDF drawing in the same folder as the active
// NC file. Surfaced as a Drawing link tile and (once Online3DViewer
// lands) as a 3D part viewer below the toolpath.
import type { CncSiblings } from "@/api/cnc";
const siblings = ref<CncSiblings>({});

const fetchSiblings = async (path: string) => {
  try {
    siblings.value = await cncApi.getSiblings(path);
  } catch {
    siblings.value = {};
  }
};

watch(
  () => cncStore.filePath,
  (p) => {
    if (p) {
      fetchSiblings(p);
    } else {
      siblings.value = {};
    }
  },
  { immediate: false }
);

// ── NC content fetch (drives the mirror + 3D toolpath) ────────────────────
// When the streamer reports a filePath (job is running, or just ended
// and the streamer still has the path), pull the NC content via the
// resources API so we can render code + toolpath. Refetch only when
// filePath changes — re-fetching on lineCurrent ticks would be silly.
const ncContent = ref<string | null>(null);
const ncLoading = ref(false);
const ncError = ref<string>("");

const fetchNc = async (path: string) => {
  ncLoading.value = true;
  ncError.value = "";
  try {
    const res = await filesApi.fetch(path);
    ncContent.value = (res as any).content ?? "";
  } catch (e: any) {
    ncError.value = e?.message || "fetch failed";
    ncContent.value = null;
  } finally {
    ncLoading.value = false;
  }
};

watch(
  () => cncStore.filePath,
  (p) => {
    if (p) {
      fetchNc(p);
    } else {
      ncContent.value = null;
      ncError.value = "";
    }
  },
  { immediate: false }
);

// ── Send progress timing (elapsed + ETA + lines/sec) ─────────────────────
// `now` ticks every second while running so the elapsed and ETA
// figures update without waiting for a status frame from the server.
const now = ref<number>(Date.now());
let nowTimer: ReturnType<typeof setInterval> | null = null;

watch(
  () => cncStore.running,
  (running) => {
    if (running && !nowTimer) {
      now.value = Date.now();
      nowTimer = setInterval(() => {
        now.value = Date.now();
      }, 1000);
    } else if (!running && nowTimer) {
      clearInterval(nowTimer);
      nowTimer = null;
    }
  },
  { immediate: true }
);

const elapsedMs = computed(() => {
  const startedAt = cncStore.raw?.started_at;
  if (!startedAt) return 0;
  const t0 = new Date(startedAt).getTime();
  if (!Number.isFinite(t0)) return 0;
  return Math.max(0, now.value - t0);
});

const linesPerSec = computed<number | null>(() => {
  const e = elapsedMs.value;
  if (e < 1000) return null; // not enough samples yet
  const lps = cncStore.lineCurrent / (e / 1000);
  return Number.isFinite(lps) && lps > 0 ? lps : null;
});

const etaMs = computed<number | null>(() => {
  const lps = linesPerSec.value;
  if (lps === null) return null;
  if (cncStore.lineTotal <= 0) return null;
  const remaining = Math.max(0, cncStore.lineTotal - cncStore.lineCurrent);
  return (remaining / lps) * 1000;
});

const fmtDuration = (ms: number): string => {
  if (!Number.isFinite(ms) || ms < 0) return "—";
  const totalSec = Math.floor(ms / 1000);
  const h = Math.floor(totalSec / 3600);
  const m = Math.floor((totalSec % 3600) / 60);
  const s = totalSec % 60;
  if (h > 0) return `${h}h ${String(m).padStart(2, "0")}m ${String(s).padStart(2, "0")}s`;
  if (m > 0) return `${m}m ${String(s).padStart(2, "0")}s`;
  return `${s}s`;
};

// ── NC card splitter (code | toolpath) ───────────────────────────────────
// Same pointer-events pattern as the editor splitter (PR #37). Stored
// per-browser in localStorage so an operator's preferred ratio
// persists across reloads. Defaults to 50/50; clamps to 15..85% so a
// pane never collapses into invisibility.
const NC_SPLIT_KEY = "machineNcSplitPct";
const NC_SPLIT_DEFAULT = 50;
const NC_SPLIT_MIN = 15;
const NC_SPLIT_MAX = 85;
const ncSplitPct = ref<number>(
  (() => {
    const stored = Number(localStorage.getItem(NC_SPLIT_KEY));
    return Number.isFinite(stored) && stored >= NC_SPLIT_MIN && stored <= NC_SPLIT_MAX
      ? stored
      : NC_SPLIT_DEFAULT;
  })()
);

const startNcResize = (e: PointerEvent) => {
  e.preventDefault();
  const target = e.currentTarget as HTMLElement;
  const layout = target.parentElement;
  if (!layout) return;
  const rect = layout.getBoundingClientRect();
  const pointerId = e.pointerId;
  try {
    target.setPointerCapture(pointerId);
  } catch {
    /* ignore */
  }
  const onMove = (ev: PointerEvent) => {
    const pct = ((ev.clientX - rect.left) / rect.width) * 100;
    ncSplitPct.value = Math.min(NC_SPLIT_MAX, Math.max(NC_SPLIT_MIN, pct));
  };
  const onUp = () => {
    target.removeEventListener("pointermove", onMove);
    target.removeEventListener("pointerup", onUp);
    target.removeEventListener("pointercancel", onUp);
    document.body.style.cursor = "";
    document.body.style.userSelect = "";
    try {
      target.releasePointerCapture(pointerId);
    } catch {
      /* ignore */
    }
    localStorage.setItem(NC_SPLIT_KEY, String(ncSplitPct.value));
  };
  document.body.style.cursor = "col-resize";
  document.body.style.userSelect = "none";
  target.addEventListener("pointermove", onMove);
  target.addEventListener("pointerup", onUp);
  target.addEventListener("pointercancel", onUp);
};

const resetNcSplit = () => {
  ncSplitPct.value = NC_SPLIT_DEFAULT;
  localStorage.setItem(NC_SPLIT_KEY, String(NC_SPLIT_DEFAULT));
};

// ── Connection check (button in card-header) ──────────────────────────────
const checkResult = ref<CncCheckResult | null>(null);
const checking = ref(false);
const checkResultClass = computed(() => {
  if (!checkResult.value) return "";
  if (checkResult.value.bridge.ok && checkResult.value.controller.ok) {
    return "check-result--ok";
  }
  return "check-result--err";
});

const runConnectionCheck = async () => {
  checking.value = true;
  try {
    checkResult.value = await cncApi.checkConnection();
  } catch (e: any) {
    $showError(e);
    checkResult.value = null;
  } finally {
    checking.value = false;
  }
};

// ── Live telemetry from useCncStore ────────────────────────────────────────
// Initial seed via /api/cnc/state on mount; from then on, WS "metric"
// events keep the store fresh. A background 30 s reseed runs as a
// belt-and-braces safety net for cases where the WS dropped silently.
const cncStore = useCncStore();
const RESEED_MS = 30_000;
let reseedTimer: ReturnType<typeof setInterval> | null = null;

const metric = (key: string): CncMetric | undefined => cncStore.metrics[key];
const parsed = (key: string): unknown => metric(key)?.parsed ?? null;
const rawValue = (key: string): string => metric(key)?.value ?? "";

const formatNum = (v: unknown, digits = 1): string => {
  if (typeof v === "number" && Number.isFinite(v)) {
    return digits === 0 ? Math.round(v).toString() : v.toFixed(digits);
  }
  return "—";
};

const fmtTs = (ts: number): string => {
  const d = new Date(ts);
  return d.toLocaleTimeString();
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
  Object.values(cncStore.metrics).some((m) => m && !m.stale)
);

// Watchdog: after the first wake (mount), give the aggregator some
// time to land at least one fresh metric. If nothing fresh after 8 s
// the bridge probably isn't responding — surface "couldn't connect"
// instead of the indefinite "Waiting for first poll…" hint.
const connectTimedOut = ref(false);
let connectWatchdog: ReturnType<typeof setTimeout> | null = null;
watch(
  anyFresh,
  (fresh) => {
    if (fresh && connectWatchdog) {
      clearTimeout(connectWatchdog);
      connectWatchdog = null;
      connectTimedOut.value = false;
    }
  },
  { immediate: false }
);

// ── Camera dispatch ────────────────────────────────────────────────────────
// Operator picks an explicit type in Settings → Machine. "auto" keeps
// the legacy URL-suffix dispatch. Browsers can't play raw RTSP/RTSPS,
// so any rtsp(s)://… URL renders the "not browser-renderable" hint —
// the proper path for UniFi Protect / Reolink is to switch the type
// to "iframe" and paste the controller's web Live View URL.
const snapshotTick = ref(0);
let snapshotTimer: ReturnType<typeof setInterval> | null = null;

const cameraKind = computed<"none" | "hls" | "snapshot" | "iframe" | "rtsp">(() => {
  const u = cameraURL.value;
  if (!u) return "none";
  // Honor an explicit type unconditionally (except "auto"). Operators
  // who selected "iframe" likely pasted a URL the suffix heuristic
  // would have misclassified as snapshot.
  switch (cameraType.value) {
    case "none":
      return "none";
    case "hls":
      return "hls";
    case "mjpeg":
      return "snapshot";
    case "iframe":
      return "iframe";
    case "auto":
    case undefined:
    case "":
      break;
    default:
      // Unknown values fall through to auto rather than blanking the tile.
      break;
  }
  if (u.startsWith("rtsp://") || u.startsWith("rtsps://")) return "rtsp";
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
    // Phase C still pending — read the default machine (machines[0])
    // for camera + host config, falling back to the legacy mirrored
    // top-level fields the server keeps populated for compat.
    const m = s.machines?.[0];
    cameraURL.value = m?.cameraUrl ?? s.cameraUrl ?? "";
    cameraType.value = m?.cameraType ?? "auto";
    hostConfigured.value = !!(m?.host ?? s.haasHost);
  } catch {
    /* ignore — view renders the configure-me hints */
  }
  // Seed the store once; WS "metric" events from /api/cnc/stream will
  // keep it fresh after that. The 30 s reseed catches the case where
  // the WS was dropped silently and reconnect failed quietly.
  await cncStore.seedMetrics();
  reseedTimer = setInterval(() => cncStore.seedMetrics(), RESEED_MS);
  // Page-reload during a job: pollOnce already populated the store
  // from /status above (via seedMetrics chain → status push); if a
  // filePath landed, fetch its NC content now.
  if (cncStore.filePath) {
    fetchNc(cncStore.filePath);
    fetchSiblings(cncStore.filePath);
  }
  // Start the connect watchdog if we still have nothing fresh after
  // the seed. The watch above clears it as soon as a metric lands.
  if (!anyFresh.value && hostConfigured.value) {
    connectWatchdog = setTimeout(() => {
      connectTimedOut.value = true;
      connectWatchdog = null;
    }, 8000);
  }
  if (cameraKind.value === "snapshot" && !snapshotTimer) {
    snapshotTimer = setInterval(() => snapshotTick.value++, 200);
  }
});

onBeforeUnmount(() => {
  if (reseedTimer) clearInterval(reseedTimer);
  if (snapshotTimer) clearInterval(snapshotTimer);
  if (connectWatchdog) clearTimeout(connectWatchdog);
  if (nowTimer) {
    clearInterval(nowTimer);
    nowTimer = null;
  }
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

/* Kiosk view (HA Lovelace embed / shop tablet): no global header,
   tighter padding, no background contrast against the embedder. */
.machine-page--kiosk {
  background: transparent;
}

.machine-page--kiosk .machine-grid {
  padding: 0.4rem;
  gap: 0.4rem;
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

.card-header__hint--err {
  color: #c62828;
}

.stop-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.3rem 0.7rem;
  border: 1px solid #c0392b;
  border-radius: 4px;
  background: rgba(192, 57, 43, 0.12);
  color: #c0392b;
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
}

.stop-btn:hover {
  background: rgba(192, 57, 43, 0.22);
}

.stop-btn .material-icons {
  font-size: 1rem;
}

.check-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.3rem 0.7rem;
  border: 1px solid var(--border-color, #ccc);
  border-radius: 4px;
  background: transparent;
  color: var(--textSecondary, inherit);
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  margin-right: 0.4rem;
}

.check-btn:hover:not(:disabled) {
  background: var(--alt-background, #f0f0f0);
}

.check-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.check-btn .material-icons {
  font-size: 1rem;
}

.check-result {
  border: 1px solid var(--border-color, #eee);
  border-radius: 6px;
  padding: 0.5rem 0.7rem;
  margin-bottom: 0.6rem;
  font-size: 0.85rem;
}

.check-result--ok {
  border-color: #2e7d32;
  background: rgba(46, 125, 50, 0.08);
}

.check-result--err {
  border-color: #c62828;
  background: rgba(198, 40, 40, 0.08);
}

.check-result__row {
  display: grid;
  grid-template-columns: 1.4rem 6rem 1fr;
  gap: 0.5rem;
  align-items: center;
  padding: 0.15rem 0;
}

.check-result__row .material-icons {
  font-size: 1.1rem;
}

.check-result--ok .material-icons {
  color: #2e7d32;
}

.check-result--err .material-icons {
  color: #c62828;
}

.check-result__label {
  font-weight: 600;
  text-transform: uppercase;
  font-size: 0.7rem;
  letter-spacing: 0.05em;
}

.check-result__detail {
  color: var(--fg-muted, #666);
  font-variant-numeric: tabular-nums;
  word-break: break-word;
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

.camera-frame--iframe {
  border: 0;
  display: block;
}

.sibling-link {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.2rem 0.55rem;
  margin-right: 0.4rem;
  border: 1px solid var(--border-color, #ccc);
  border-radius: 4px;
  font-size: 0.78rem;
  text-decoration: none;
  color: var(--textSecondary, inherit);
}

.sibling-link:hover {
  background: var(--alt-background, #f0f0f0);
}

.sibling-link .material-icons {
  font-size: 0.95rem;
}

/* NC card: side-by-side code mirror + 3D toolpath, with a draggable
   splitter between them. Width is driven by --nc-pct (an inline
   style set in the template, persisted to localStorage). */
.nc-card .nc-split {
  display: flex;
  height: 50vh;
  min-height: 400px;
  --nc-pct: 50%;
}

.nc-split__pane {
  position: relative;
  min-width: 0;
  border: 1px solid var(--border-color, #eee);
  border-radius: 6px;
  overflow: hidden;
  background: var(--surfacePrimary, #fff);
}

.nc-split__pane--code {
  flex: 0 0 var(--nc-pct);
  margin-right: 0; /* gap is the splitter bar */
}

.nc-split__pane--viewer {
  flex: 1 1 0;
  background: #111;
}

.nc-split__bar {
  flex: 0 0 6px;
  cursor: col-resize;
  background: transparent;
  position: relative;
  user-select: none;
  z-index: 1;
  touch-action: none;
}

.nc-split__bar::before {
  content: "";
  position: absolute;
  top: 0;
  bottom: 0;
  left: 50%;
  width: 1px;
  background: var(--border-color, #ddd);
  transform: translateX(-50%);
}

.nc-split__bar:hover::before,
.nc-split__bar:active::before {
  background: var(--primaryColor, #2196f3);
  width: 2px;
}

@media (max-width: 900px) {
  .nc-card .nc-split {
    flex-direction: column;
    height: auto;
  }
  .nc-split__pane--code,
  .nc-split__pane--viewer {
    flex: 0 0 50vh;
  }
  .nc-split__bar {
    display: none;
  }
}

/* Part 3D viewer card sits below the NC card */
.part-card .part-card__body {
  height: 50vh;
  min-height: 380px;
  position: relative;
}

/* Live send progress strip — visible only while streaming */
.send-progress {
  background: var(--alt-background, #fafafa);
  border: 1px solid var(--primaryColor, #2196f3);
  border-radius: 6px;
  padding: 0.5rem 0.7rem;
  margin-bottom: 0.6rem;
}

.send-progress__head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  gap: 0.6rem;
  font-size: 0.85rem;
  margin-bottom: 0.35rem;
}

.send-progress__file {
  font-weight: 600;
  word-break: break-all;
}

.send-progress__counter {
  font-variant-numeric: tabular-nums;
  color: var(--fg-muted, #888);
  flex-shrink: 0;
}

.send-progress__pct {
  margin-left: 0.25rem;
}

.send-progress__bar {
  height: 6px;
  background: var(--border-color, #eee);
  border-radius: 3px;
  overflow: hidden;
}

.send-progress__bar-fill {
  height: 100%;
  background: var(--primaryColor, #2196f3);
  transition: width 0.2s ease;
}

.send-progress__time {
  display: flex;
  flex-wrap: wrap;
  gap: 0.9rem;
  margin-top: 0.4rem;
  font-size: 0.78rem;
  color: var(--fg-muted, #888);
  font-variant-numeric: tabular-nums;
}

.send-progress__time span {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
}

.send-progress__time .material-icons {
  font-size: 0.95rem;
}

/* Activity log — backend log events + status transitions */
.activity {
  background: var(--alt-background, #fafafa);
  border: 1px solid var(--border-color, #eee);
  border-radius: 6px;
  padding: 0.6rem 0.8rem;
  margin-top: 0.6rem;
}

.activity__title {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--fg-muted, #888);
  margin-bottom: 0.4rem;
}

.activity__clear {
  background: none;
  border: 1px solid var(--border-color, #ccc);
  border-radius: 3px;
  padding: 0.1rem 0.5rem;
  color: var(--fg-muted, #888);
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  cursor: pointer;
}

.activity__clear:hover {
  background: var(--surfacePrimary, #fff);
  color: var(--textSecondary, inherit);
}

.activity__empty {
  color: var(--fg-muted, #888);
  font-size: 0.85rem;
  font-style: italic;
}

.activity__list {
  list-style: none;
  margin: 0;
  padding: 0;
  max-height: 14rem;
  overflow-y: auto;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 0.8rem;
}

.activity__row {
  display: grid;
  grid-template-columns: auto 4.5rem 1fr;
  gap: 0.6rem;
  padding: 0.15rem 0;
  border-bottom: 1px solid var(--border-color, #f0f0f0);
}

.activity__row:last-child {
  border-bottom: none;
}

.activity__ts {
  color: var(--fg-muted, #888);
  font-variant-numeric: tabular-nums;
}

.activity__level {
  font-weight: 600;
  text-transform: uppercase;
  font-size: 0.7rem;
  align-self: center;
}

.activity__row--error .activity__level {
  color: #c62828;
}

.activity__row--info .activity__level {
  color: var(--primaryColor, #2196f3);
}

.activity__msg {
  word-break: break-word;
}

.hint {
  margin: auto;
  padding: 1rem 1.5rem;
  color: var(--fg-muted, #888);
  font-size: 0.9rem;
  text-align: center;
}
</style>
