<template>
  <div class="m-modal-backdrop" @click.self="$emit('close')">
    <div class="m-modal">
      <div class="m-modal__header">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="m-modal__tab"
          :class="{ 'm-modal__tab--active': current === tab.key }"
          @click="current = tab.key"
        >
          {{ tab.label }}
        </button>
        <span class="m-modal__spacer" />
        <button class="m-modal__close" @click="$emit('close')">✕</button>
      </div>
      <div class="m-modal__body">
        <template v-if="current === 'bridge'">
          <h3>{{ t("machine.connModalBridge") }}</h3>
          <div class="m-modal__stat">
            <span>{{ t("machine.connModalAddress") }}</span>
            <code v-if="checkResult?.bridge.address">{{ checkResult.bridge.address }}</code>
            <span v-else>—</span>
          </div>
          <div class="m-modal__stat">
            <span>{{ t("machine.connModalLastLatency") }}</span>
            <span v-if="checkResult?.bridge.latency_ms !== undefined">
              {{ Math.round(checkResult.bridge.latency_ms) }} ms
            </span>
            <span v-else>—</span>
          </div>
          <div v-if="checkResult?.bridge.error" class="m-modal__err">
            {{ checkResult.bridge.error }}
          </div>
          <button class="m-modal__btn" :disabled="checking" @click="runCheck">
            {{ checking ? t("machine.checking") : t("machine.checkConnection") }}
          </button>
        </template>

        <template v-else-if="current === 'controller'">
          <h3>{{ t("machine.connModalController") }}</h3>
          <div class="m-modal__stat">
            <span>{{ t("machine.connModalLastQ104") }}</span>
            <span v-if="checkResult?.controller.mode">{{ checkResult.controller.mode }}</span>
            <span v-else>—</span>
          </div>
          <div class="m-modal__stat">
            <span>{{ t("machine.connModalLastLatency") }}</span>
            <span v-if="checkResult?.controller.latency_ms !== undefined">
              {{ Math.round(checkResult.controller.latency_ms) }} ms
            </span>
            <span v-else>—</span>
          </div>
          <div v-if="checkResult?.controller.error" class="m-modal__err">
            {{ checkResult.controller.error }}
          </div>
          <button class="m-modal__btn" :disabled="checking" @click="runCheck">
            {{ checking ? t("machine.checking") : t("machine.checkConnection") }}
          </button>
        </template>

        <template v-else>
          <h3>{{ t("machine.connModalActivity") }}</h3>
          <div class="m-modal__activity-filters">
            <label v-for="lvl in ['info', 'warn', 'error']" :key="lvl">
              <input type="checkbox" v-model="filterLevels" :value="lvl" />
              {{ lvl }}
            </label>
            <span class="m-modal__spacer" />
            <button class="m-modal__btn" @click="cnc.clearLog()">{{ t("machine.activityClear") }}</button>
          </div>
          <ol class="m-modal__activity">
            <li v-for="(entry, i) in filteredLog" :key="i" :class="`m-activity-row m-activity-row--${entry.level}`">
              <span class="m-activity-row__ts">{{ fmtTs(entry.ts) }}</span>
              <span class="m-activity-row__level">{{ entry.level }}</span>
              <span class="m-activity-row__msg">{{ entry.msg }}</span>
            </li>
            <li v-if="filteredLog.length === 0" class="m-activity-empty">
              {{ t("machine.activityEmpty") }}
            </li>
          </ol>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useCncStore } from "@/stores/cnc";
import { cnc as cncApi } from "@/api";
import type { CncCheckResult } from "@/api/cnc";

const { t } = useI18n();
const cnc = useCncStore();

const props = defineProps<{ initialTab?: "bridge" | "controller" | "activity" }>();
defineEmits<{ (e: "close"): void }>();

type TabKey = "bridge" | "controller" | "activity";
const tabs = computed<{ key: TabKey; label: string }[]>(() => [
  { key: "bridge", label: t("machine.connModalBridge") },
  { key: "controller", label: t("machine.connModalController") },
  { key: "activity", label: t("machine.connModalActivity") },
]);

const current = ref<"bridge" | "controller" | "activity">(props.initialTab || "bridge");
watch(() => props.initialTab, (v) => { if (v) current.value = v; });

const filterLevels = ref<string[]>(["info", "warn", "error"]);
const filteredLog = computed(() =>
  cnc.log.filter((e) => filterLevels.value.includes(e.level || "info"))
);

const checkResult = ref<CncCheckResult | null>(null);
const checking = ref(false);
const runCheck = async () => {
  checking.value = true;
  try {
    checkResult.value = await cncApi.checkConnection(
      cnc.currentMachineId || undefined
    );
  } catch (e: any) {
    checkResult.value = {
      bridge: { ok: false, error: e?.message || String(e) },
      controller: { ok: false, error: "skipped — bridge unreachable" },
    };
  } finally {
    checking.value = false;
  }
};

const fmtTs = (ts: number) => new Date(ts).toLocaleTimeString();
</script>

<style scoped>
.m-modal-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}
.m-modal {
  width: min(640px, 92vw);
  max-height: 80vh;
  background: var(--surface, #fff);
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 12px 32px rgba(0,0,0,0.25);
}
.m-modal__header {
  display: flex;
  align-items: center;
  padding: 6px 6px 0;
  gap: 4px;
  border-bottom: 1px solid var(--border-color, #eee);
}
.m-modal__tab {
  font-size: 11px;
  padding: 6px 12px;
  background: transparent;
  border: 0;
  border-bottom: 2px solid transparent;
  cursor: pointer;
  color: var(--textSecondary, #555);
}
.m-modal__tab--active {
  color: var(--textPrimary, #222);
  border-bottom-color: #185FA5;
}
.m-modal__spacer { flex: 1 1 0; }
.m-modal__close {
  background: transparent;
  border: 0;
  font-size: 16px;
  cursor: pointer;
  color: var(--fg-muted, #888);
  padding: 4px 8px;
}
.m-modal__body {
  padding: 12px 16px;
  overflow: auto;
}
.m-modal__body h3 {
  margin: 0 0 8px;
  font-size: 14px;
  font-weight: 500;
}
.m-modal__stat {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  padding: 4px 0;
  border-bottom: 1px solid var(--border-color, #eee);
}
.m-modal__err {
  margin: 8px 0;
  padding: 6px 8px;
  background: rgba(198, 40, 40, 0.08);
  color: #c62828;
  font-size: 11px;
  border-radius: 4px;
}
.m-modal__btn {
  margin-top: 8px;
  padding: 6px 12px;
  background: var(--alt-background, #f4f4f4);
  border: 1px solid var(--border-color, #ddd);
  border-radius: 4px;
  font-size: 11px;
  cursor: pointer;
}
.m-modal__btn:hover:not(:disabled) { filter: brightness(0.97); }

.m-modal__activity-filters {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 10px;
  margin-bottom: 8px;
}
.m-modal__activity {
  list-style: none;
  margin: 0;
  padding: 0;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 10px;
  max-height: 40vh;
  overflow-y: auto;
  overscroll-behavior: contain;
}
.m-activity-row {
  display: grid;
  grid-template-columns: auto 4.5rem 1fr;
  gap: 8px;
  padding: 2px 0;
  border-bottom: 1px solid var(--border-color, #f4f4f4);
}
.m-activity-row__ts { color: var(--fg-muted, #888); }
.m-activity-row__level { font-weight: 600; text-transform: uppercase; font-size: 9px; }
.m-activity-row--error .m-activity-row__level { color: #c62828; }
.m-activity-row--info .m-activity-row__level { color: #1976d2; }
.m-activity-row--warn .m-activity-row__level { color: #b85100; }
.m-activity-empty { color: var(--fg-muted, #888); padding: 6px 0; }
</style>
