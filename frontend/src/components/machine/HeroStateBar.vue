<template>
  <div class="m-hero" :class="`m-hero--${variant}`">
    <div class="m-hero__left">
      <div class="m-hero__label">{{ t("machine.program") }} · {{ t("machine.statusLabel") }}</div>
      <div class="m-hero__headline">
        {{ program || "—" }} <span class="m-hero__sep">·</span> {{ status || "—" }}
      </div>
    </div>
    <div class="m-hero__stats">
      <div class="m-hero__stat">
        <div class="m-hero__label">{{ t("machine.spindleRpm") }}</div>
        <div class="m-hero__stat-val">{{ spindle }}</div>
      </div>
      <div class="m-hero__stat">
        <div class="m-hero__label">{{ t("machine.tool") }}</div>
        <div class="m-hero__stat-val">{{ tool }}</div>
      </div>
      <div class="m-hero__stat">
        <div class="m-hero__label">{{ t("machine.parts") }}</div>
        <div class="m-hero__stat-val">{{ parts }}</div>
      </div>
      <div class="m-hero__stat">
        <div class="m-hero__label">{{ t("machine.lastCycle") }}</div>
        <div class="m-hero__stat-val">{{ cycle }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useCncStore } from "@/stores/cnc";

const { t } = useI18n();
const cnc = useCncStore();

const metric = (key: string) => cnc.metrics[key];
const parsed = (key: string): unknown => metric(key)?.parsed ?? null;
const rawValue = (key: string): string => metric(key)?.value ?? "";

const program = computed(() => {
  const p = parsed("status_combined");
  if (p && typeof p === "object" && "program" in (p as Record<string, unknown>)) {
    return (p as Record<string, string>).program;
  }
  return rawValue("status_combined");
});

const status = computed(() => {
  const p = parsed("status_combined");
  if (p && typeof p === "object" && "status" in (p as Record<string, unknown>)) {
    return ((p as Record<string, string>).status || "").toUpperCase();
  }
  return "";
});

const variant = computed<"running" | "idle" | "warn" | "error">(() => {
  const s = (status.value || "").toLowerCase();
  if (s.includes("run")) return "running";
  if (s.includes("alarm") || s.includes("fault") || s.includes("error")) return "error";
  if (s.includes("hold") || s.includes("feed")) return "warn";
  return "idle";
});

const fmtNum = (v: unknown, digits = 0) => {
  if (typeof v === "number" && Number.isFinite(v)) {
    return digits === 0 ? Math.round(v).toLocaleString() : v.toFixed(digits);
  }
  return "—";
};

const spindle = computed(() => fmtNum(parsed("spindle_actual")));
const tool = computed(() => {
  const v = parsed("tool");
  return typeof v === "number" ? `T${Math.round(v)}` : "—";
});
const parts = computed(() => fmtNum(parsed("parts")));
const cycle = computed(() => rawValue("last_cycle") || "—");
</script>

<style scoped>
.m-hero {
  border-radius: 6px;
  padding: 6px 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
  gap: 8px;
  color: #E6F1FB;
}

.m-hero--idle { background: var(--alt-background, #5a5a5a); color: var(--textPrimary, #fff); }
.m-hero--running { background: #185FA5; color: #E6F1FB; }
.m-hero--warn { background: #b85100; color: #FFF3E6; }
.m-hero--error { background: #9b1c1c; color: #FFE8E8; }

.m-hero__left { min-width: 0; }
.m-hero__label {
  font-size: 9px;
  opacity: 0.7;
  letter-spacing: 0.5px;
  text-transform: uppercase;
  font-weight: 500;
}
.m-hero__headline {
  font-size: 16px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.m-hero__sep { opacity: 0.6; padding: 0 0.2em; }

.m-hero__stats { display: flex; gap: 22px; flex-shrink: 0; }
.m-hero__stat-val { font-size: 13px; font-weight: 500; font-variant-numeric: tabular-nums; }

@media (max-width: 640px) {
  .m-hero { flex-direction: column; align-items: flex-start; gap: 6px; }
  .m-hero__stats { display: grid; grid-template-columns: repeat(2, 1fr); gap: 6px 18px; width: 100%; }
}
</style>
