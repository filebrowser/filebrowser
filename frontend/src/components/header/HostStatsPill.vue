<template>
  <div v-if="stats" class="host-pill" :title="tooltip">
    <span class="host-pill__seg" :class="tempClass">
      <i class="material-icons">device_thermostat</i>
      {{ tempLabel }}
    </span>
    <span class="host-pill__seg" :class="loadClass">
      <i class="material-icons">speed</i>
      {{ loadLabel }}
    </span>
    <span class="host-pill__seg" :class="memClass">
      <i class="material-icons">memory</i>
      {{ memLabel }}
    </span>
    <span class="host-pill__seg" :class="diskClass">
      <i class="material-icons">storage</i>
      {{ diskLabel }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { useAuthStore } from "@/stores/auth";
import { useI18n } from "vue-i18n";
import { fetchJSON } from "@/api/utils";

interface HostStats {
  temp_c?: number;
  load_1m?: number;
  mem_used_pct?: number;
  disk_used_pct?: number;
  uptime_s?: number;
  cores?: number;
}

const { t } = useI18n();
const auth = useAuthStore();
const stats = ref<HostStats | null>(null);
let timer: ReturnType<typeof setInterval> | null = null;

// Poll cadence: 10 s. Each call is a handful of /proc reads + one
// statfs — well under a millisecond on a Pi 4 — but it adds up on a
// public deploy with many sessions, so we keep it conservative.
const POLL_MS = 10_000;

const poll = async () => {
  if (!auth.user) return;
  try {
    stats.value = await fetchJSON<HostStats>("/api/cnc/host-stats", {});
  } catch {
    // Silent — pill just hides when the endpoint is unreachable.
    stats.value = null;
  }
};

onMounted(() => {
  poll();
  timer = setInterval(poll, POLL_MS);
});
onUnmounted(() => {
  if (timer) clearInterval(timer);
});

const tempLabel = computed(() => {
  const t = stats.value?.temp_c;
  if (!t) return "—";
  return `${t.toFixed(0)}°C`;
});
const tempClass = computed(() => {
  const t = stats.value?.temp_c;
  if (!t) return "";
  if (t >= 80) return "host-pill__seg--alarm";
  if (t >= 70) return "host-pill__seg--warn";
  return "";
});

const loadLabel = computed(() => {
  const l = stats.value?.load_1m;
  if (l == null) return "—";
  return l.toFixed(2);
});
const loadClass = computed(() => {
  const l = stats.value?.load_1m;
  const cores = stats.value?.cores || 1;
  if (l == null) return "";
  if (l >= cores * 1.5) return "host-pill__seg--alarm";
  if (l >= cores) return "host-pill__seg--warn";
  return "";
});

const memLabel = computed(() => {
  const m = stats.value?.mem_used_pct;
  if (m == null) return "—";
  return `${m.toFixed(0)}%`;
});
const memClass = computed(() => {
  const m = stats.value?.mem_used_pct;
  if (m == null) return "";
  if (m >= 90) return "host-pill__seg--alarm";
  if (m >= 75) return "host-pill__seg--warn";
  return "";
});

const diskLabel = computed(() => {
  const d = stats.value?.disk_used_pct;
  if (d == null) return "—";
  return `${d.toFixed(0)}%`;
});
const diskClass = computed(() => {
  const d = stats.value?.disk_used_pct;
  if (d == null) return "";
  if (d >= 95) return "host-pill__seg--alarm";
  if (d >= 85) return "host-pill__seg--warn";
  return "";
});

const tooltip = computed(() => {
  const s = stats.value;
  if (!s) return "";
  const lines: string[] = [];
  if (s.temp_c) lines.push(`${t("hostStats.temp")}: ${s.temp_c.toFixed(1)}°C`);
  if (s.load_1m != null) {
    const cores = s.cores ? ` (${s.cores} cores)` : "";
    lines.push(`${t("hostStats.load1m")}: ${s.load_1m.toFixed(2)}${cores}`);
  }
  if (s.mem_used_pct != null) lines.push(`${t("hostStats.memUsed")}: ${s.mem_used_pct.toFixed(1)}%`);
  if (s.disk_used_pct != null) lines.push(`${t("hostStats.diskUsed")}: ${s.disk_used_pct.toFixed(1)}%`);
  if (s.uptime_s) lines.push(`${t("hostStats.uptime")}: ${formatUptime(s.uptime_s)}`);
  return lines.join("\n");
});

function formatUptime(sec: number): string {
  const d = Math.floor(sec / 86400);
  const h = Math.floor((sec % 86400) / 3600);
  const m = Math.floor((sec % 3600) / 60);
  if (d > 0) return `${d}d ${h}h ${m}m`;
  if (h > 0) return `${h}h ${m}m`;
  return `${m}m`;
}
</script>

<style>
.host-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.2rem 0.6rem;
  margin: 0 0.5rem;
  border-radius: 999px;
  background: var(--alt-background, rgba(128, 128, 128, 0.1));
  color: var(--textPrimary, #555);
  font-size: 0.78rem;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}
.host-pill__seg {
  display: inline-flex;
  align-items: center;
  gap: 0.2rem;
  opacity: 0.85;
}
.host-pill__seg i {
  font-size: 0.95rem;
  opacity: 0.7;
}
.host-pill__seg--warn {
  color: #d28a00;
  opacity: 1;
}
.host-pill__seg--warn i {
  opacity: 0.95;
}
.host-pill__seg--alarm {
  color: #c0392b;
  opacity: 1;
  font-weight: 600;
}
.host-pill__seg--alarm i {
  opacity: 1;
}

@media (max-width: 900px) {
  /* Drop disk on smaller screens; temp + load + mem are the most
     useful when the operator only has 900px to play with. */
  .host-pill .host-pill__seg:nth-child(4) {
    display: none;
  }
}
@media (max-width: 640px) {
  .host-pill .host-pill__seg:nth-child(3) {
    display: none;
  }
}
</style>
