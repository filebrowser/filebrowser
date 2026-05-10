<template>
  <div class="m-topbar">
    <button
      class="m-pill"
      :class="{ 'm-pill--ok': bridgeOK, 'm-pill--err': bridgeKnown && !bridgeOK }"
      @click="$emit('open-connection', 'bridge')"
      :title="t('machine.topBridgeTitle')"
    >
      <span class="m-pill__dot" />
      <span class="m-pill__label">{{ t("machine.topBridge") }}</span>
      <span v-if="bridgeLatency !== null" class="m-pill__muted">{{ bridgeLatency }}ms</span>
    </button>

    <button
      class="m-pill"
      :class="{ 'm-pill--ok': controllerOK, 'm-pill--err': controllerKnown && !controllerOK }"
      @click="$emit('open-connection', 'controller')"
      :title="t('machine.topControllerTitle')"
    >
      <span class="m-pill__dot" />
      <span class="m-pill__label">{{ t("machine.topController") }}</span>
    </button>

    <button
      class="m-pill"
      @click="$emit('open-connection', 'activity')"
    >
      <span class="m-pill__label">{{ t("machine.topActivity") }}</span>
      <span class="m-pill__count" :class="{ 'm-pill__count--active': activityCount > 0 }">
        {{ activityCount }}
      </span>
    </button>

    <div class="m-activity-stream" :class="{ 'm-activity-stream--info': latestIsInfo }">
      <span v-if="latest" class="m-activity-stream__ts">{{ latest.tsLabel }}</span>
      <span v-if="latest" class="m-activity-stream__msg">{{ latest.msg }}</span>
    </div>

    <button
      class="m-pill m-pill--queue"
      :class="{ 'm-pill--queue-active': queueCount > 0 }"
      @click="$emit('toggle-queue')"
    >
      <span class="m-pill__label">{{ t("machine.topQueue") }}</span>
      <span class="m-pill__count" :class="{ 'm-pill__count--active': queueCount > 0 }">{{ queueCount }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useCncStore } from "@/stores/cnc";

const { t } = useI18n();
const cnc = useCncStore();

defineEmits<{
  (e: "open-connection", tab: "bridge" | "controller" | "activity"): void;
  (e: "toggle-queue"): void;
}>();

// "Ok-ish" heuristic from the live store. The expensive truth (latency
// numbers, recent timeout history) lives in the Connection modal —
// this is just enough to color the pills.
const bridgeOK = computed(() => cnc.haasOk && cnc.metricsSeeded);
const bridgeKnown = computed(() => cnc.metricsSeeded || cnc.initialized);
// Bridge latency: comes from the last connection check. We don't have
// a real number on every poll, so when missing we just hide the suffix.
const bridgeLatency = computed<number | null>(() => null);

const controllerOK = computed(() => cnc.haasOk && Object.values(cnc.metrics).some((m) => m && !m.stale));
const controllerKnown = computed(() => cnc.initialized);

const activityCount = computed(() => cnc.log.length);

const latest = computed(() => {
  const e = cnc.log[0];
  if (!e) return null;
  const d = new Date(e.ts);
  return {
    tsLabel: d.toLocaleTimeString(),
    msg: e.msg,
    level: e.level,
  };
});
const latestIsInfo = computed(() => (latest.value?.level || "") === "info" && cnc.running);

const queueCount = computed(() => cnc.queue.length);
</script>

<style scoped>
.m-topbar {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border: 1px solid var(--border-color, #ddd);
  border-radius: 6px;
  background: var(--surface, #fff);
  font-size: 10px;
  flex-shrink: 0;
}

.m-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 8px;
  border-radius: 4px;
  background: var(--alt-background, #f4f4f4);
  border: 0;
  cursor: pointer;
  font-size: 10px;
  flex-shrink: 0;
  color: inherit;
}

.m-pill:hover { filter: brightness(0.97); }

.m-pill__dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #999;
  display: inline-block;
}

.m-pill--ok .m-pill__dot { background: #639922; }
.m-pill--err .m-pill__dot { background: #c0392b; }

.m-pill__label { font-weight: 500; color: var(--textPrimary, #222); }
.m-pill__muted { color: var(--fg-muted, #888); }

.m-pill__count {
  font-size: 9px;
  padding: 1px 6px;
  border-radius: 3px;
  background: var(--border-color, #e2e2e2);
  color: var(--fg-muted, #888);
  font-weight: 500;
}
.m-pill__count--active {
  background: rgba(33, 150, 243, 0.18);
  color: #1976d2;
}

.m-pill--queue {
  margin-left: auto;
}
.m-pill--queue-active {
  background: rgba(33, 150, 243, 0.12);
}
.m-pill--queue-active .m-pill__label {
  color: #1976d2;
}

.m-activity-stream {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 9px;
  color: var(--fg-muted, #888);
  padding: 0 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  min-width: 0;
}
.m-activity-stream--info { color: #1976d2; }
.m-activity-stream__ts { color: var(--fg-muted, #888); margin-right: 6px; }
.m-activity-stream__msg { color: inherit; }
</style>
