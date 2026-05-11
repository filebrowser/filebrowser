<template>
  <div class="m-queue" :class="{ 'm-queue--empty': cnc.queue.length === 0 }">
    <div class="m-queue__header" @click="toggle">
      <span class="m-queue__chevron">{{ expanded ? "▾" : "▸" }}</span>
      <span class="m-queue__title">{{ t("machine.queueTitle") }}</span>
      <span v-if="cnc.queue.length > 0" class="m-queue__count">· {{ cnc.queue.length }}</span>
      <span v-if="runningRow" class="m-vert-divider" />
      <span
        v-if="runningRow"
        class="m-running-crumb"
        @click.stop="onCrumbClick"
      >
        <span class="m-running-crumb__dot" />
        <span class="m-running-crumb__label">{{ runningRow.state === 'sending' ? t('machine.queueSending') : t('machine.queueRunning') }}</span>
        <span class="m-running-crumb__sep">/</span>
        <span class="m-running-crumb__text">{{ runningRow.job_name ? `${runningRow.job_name} / ` : "" }}{{ basename(runningRow.file_path) }}</span>
        <span class="m-running-crumb__progress">
          <template v-if="runningRow.onumber_hint">{{ runningRow.onumber_hint }} ·</template>
          {{ runningRow.line_current || 0 }}/{{ runningRow.line_total || 0 }} · {{ runningPct }}%
        </span>
      </span>

      <span v-if="expanded && cnc.queue.length > pageSize" class="m-pager">
        {{ pageRangeStart }}–{{ pageRangeEnd }} of {{ cnc.queue.length }}
        <button class="m-pager__btn" :disabled="page === 0" @click.stop="page = Math.max(0, page - 1)">‹</button>
        <span>{{ page + 1 }}/{{ totalPages }}</span>
        <button class="m-pager__btn" :disabled="page >= totalPages - 1" @click.stop="page = Math.min(totalPages - 1, page + 1)">›</button>
      </span>
    </div>

    <div v-if="expanded" class="m-queue__body">
      <div v-if="cnc.queue.length === 0" class="m-queue__empty">
        {{ t("machine.queueEmpty") }}
      </div>
      <ol v-else class="m-queue__rows">
        <li
          v-for="(item, idx) in pageItems"
          :key="item.id"
          class="m-queue-row"
          :class="rowClass(item)"
          draggable="true"
          @dragstart="onDragStart(item.id, $event)"
          @dragover.prevent
          @drop="onDrop(item.id)"
        >
          <span class="m-queue-row__handle" aria-hidden="true">⋮⋮</span>
          <span class="m-queue-row__num">{{ pageStartIndex + idx + 1 }}</span>
          <div class="m-queue-row__name">
            <span class="m-queue-row__title">{{ basename(item.file_path) }}</span>
            <span class="m-queue-row__meta">
              <template v-if="item.job_name">{{ item.job_name }} · </template>
              <template v-if="item.size_bytes">{{ fmtSize(item.size_bytes) }}</template>
              <template v-if="item.state === 'sending' || item.state === 'running'">
                · {{ item.method?.toUpperCase() }} · {{ item.line_current || 0 }}/{{ item.line_total || 0 }} · {{ pctOf(item) }}%
              </template>
            </span>
          </div>

          <!-- Send progressive disclosure ────────────── -->
          <template v-if="item.state === 'sending' || item.state === 'running'">
            <span class="m-queue-row__spinner" />
            <button
              v-if="item.state === 'sending'"
              class="m-queue-row__btn m-queue-row__btn--cancel"
              :title="t('machine.queueCancel')"
              @click.stop="onCancel()"
            >✕</button>
          </template>
          <template v-else>
            <button
              v-if="autoSendEnabled"
              class="m-queue-row__btn m-queue-row__btn--auto"
              :disabled="anyInFlight"
              :title="anyInFlight ? t('machine.queueBusy') : t('machine.queueAutoSendTitle')"
              @click.stop="onAutoSend(item)"
            >⚡</button>
            <button
              v-if="openSendId !== item.id"
              class="m-queue-row__btn m-queue-row__btn--send"
              :disabled="anyInFlight"
              :title="anyInFlight ? t('machine.queueBusy') : t('machine.queueSendTitle')"
              @click.stop="openSendId = item.id"
            >➤</button>
            <template v-else>
              <button class="m-queue-row__send-opt m-queue-row__send-opt--mem" @click.stop="onSend(item, 'mem')">
                {{ t("sendWizard.methodMem").split(" ")[0] }}
              </button>
              <button class="m-queue-row__send-opt" @click.stop="onSend(item, 'dnc')">
                DNC
              </button>
              <button class="m-queue-row__btn" @click.stop="openSendId = ''">✕</button>
            </template>
          </template>

          <span class="m-queue-row__spacer" />

          <button
            class="m-queue-row__btn"
            :title="t('machine.queueOverflow')"
            @click.stop="openOverflowId = openOverflowId === item.id ? '' : item.id"
          >⋯</button>
          <div v-if="openOverflowId === item.id" class="m-overflow-menu" @click.stop>
            <button class="m-overflow__item" @click="onOpenFolder(item)"><i class="material-icons">folder</i> {{ t("machine.queueOpenFolder") }}</button>
            <button class="m-overflow__item" @click="onOpenInEditor(item)"><i class="material-icons">edit_note</i> {{ t("machine.queueOpenEditor") }}</button>
            <div class="m-overflow__sep" />
            <button class="m-overflow__item" @click="onMoveTop(item)"><i class="material-icons">vertical_align_top</i> {{ t("machine.queueMoveTop") }}</button>
            <button class="m-overflow__item" @click="onMoveBottom(item)"><i class="material-icons">vertical_align_bottom</i> {{ t("machine.queueMoveBottom") }}</button>
            <div class="m-overflow__sep" />
            <button class="m-overflow__item m-overflow__item--danger" @click="onRemove(item)"><i class="material-icons">delete</i> {{ t("machine.queueRemove") }}</button>
          </div>
        </li>
      </ol>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useRouter } from "vue-router";
import { useCncStore } from "@/stores/cnc";
import type { QueueItem, SendMethod } from "@/api/cnc";

const { t } = useI18n();
const cnc = useCncStore();
const router = useRouter();

const props = defineProps<{
  forceOpen?: boolean;
}>();

const expanded = ref(false);
const openSendId = ref<string>("");
const openOverflowId = ref<string>("");
const page = ref(0);
const pageSize = 3;

watch(
  () => props.forceOpen,
  (v) => {
    if (v) expanded.value = true;
  },
  { immediate: true }
);

const toggle = () => {
  expanded.value = !expanded.value;
};

const runningRow = computed<QueueItem | undefined>(() =>
  cnc.queue.find((q) => q.state === "running" || q.state === "sending")
);

const anyInFlight = computed(() => !!runningRow.value);

const totalPages = computed(() => Math.max(1, Math.ceil(cnc.queue.length / pageSize)));
const pageStartIndex = computed(() => page.value * pageSize);
const pageRangeStart = computed(() => Math.min(cnc.queue.length, pageStartIndex.value + 1));
const pageRangeEnd = computed(() => Math.min(cnc.queue.length, pageStartIndex.value + pageSize));
const pageItems = computed(() => cnc.queue.slice(pageStartIndex.value, pageStartIndex.value + pageSize));

const basename = (p: string) => {
  if (!p) return "";
  const parts = p.split("/").filter(Boolean);
  return parts[parts.length - 1] || p;
};

const fmtSize = (b: number) => {
  if (b < 1024) return `${b} B`;
  if (b < 1024 * 1024) return `${(b / 1024).toFixed(0)} KB`;
  return `${(b / 1024 / 1024).toFixed(1)} MB`;
};

const pctOf = (item: QueueItem) => {
  if (!item.line_total) return 0;
  return Math.round((Number(item.line_current) / Number(item.line_total)) * 100);
};

const runningPct = computed(() => (runningRow.value ? pctOf(runningRow.value) : 0));

const rowClass = (item: QueueItem) => ({
  "m-queue-row--sending": item.state === "sending",
  "m-queue-row--running": item.state === "running",
  "m-queue-row--locked": anyInFlight.value && item.state === "queued",
});

// ── Drag reorder ──
const dragId = ref<string>("");
const onDragStart = (id: string, ev: DragEvent) => {
  dragId.value = id;
  if (ev.dataTransfer) ev.dataTransfer.effectAllowed = "move";
};
const onDrop = async (overId: string) => {
  if (!dragId.value || dragId.value === overId) return;
  const ids = cnc.queue.map((q) => q.id);
  const from = ids.indexOf(dragId.value);
  const to = ids.indexOf(overId);
  if (from < 0 || to < 0) return;
  ids.splice(from, 1);
  ids.splice(to, 0, dragId.value);
  dragId.value = "";
  await cnc.reorderQueue(ids);
};

// ── Send disclosure ──
const onSend = async (item: QueueItem, method: SendMethod) => {
  openSendId.value = "";
  try {
    await cnc.sendFromQueue(item, method);
  } catch (e) {
    /* error surfaces via store + log; the row stays "queued" */
    console.error(e);
  }
};

// autoSendEnabled mirrors the active machine's flag. The button only
// shows when the machine has opted in — otherwise the regular send
// disclosure is the only path.
const autoSendEnabled = computed(
  () => !!cnc.currentMachine?.autoSendEnabled
);

// onAutoSend hits /api/cnc/auto-send. On block (preflight failed or
// spindle swap pending) we fall back to opening the send disclosure
// so the operator can confirm in the manual path. The block reason is
// pushed to the activity log so it's visible without a modal.
const onAutoSend = async (item: QueueItem) => {
  if (anyInFlight.value) return;
  try {
    const r = await cnc.autoSendFromQueue(item, "mem");
    if (!r.started) {
      cnc.pushLog(
        "warn",
        `auto-send blocked: ${r.blocked_reason || "preflight not green"}`
      );
      openSendId.value = item.id;
    }
  } catch (e) {
    console.error(e);
  }
};

// ── Cancel (active send) ──
const emit = defineEmits<{ (e: "stop-machine"): void }>();
const onCancel = () => emit("stop-machine");

// ── Overflow actions ──
const onRemove = async (item: QueueItem) => {
  openOverflowId.value = "";
  await cnc.removeFromQueue(item.id);
};

const onMoveTop = async (item: QueueItem) => {
  openOverflowId.value = "";
  const ids = cnc.queue.map((q) => q.id).filter((x) => x !== item.id);
  ids.unshift(item.id);
  await cnc.reorderQueue(ids);
};

const onMoveBottom = async (item: QueueItem) => {
  openOverflowId.value = "";
  const ids = cnc.queue.map((q) => q.id).filter((x) => x !== item.id);
  ids.push(item.id);
  await cnc.reorderQueue(ids);
};

const onOpenInEditor = (item: QueueItem) => {
  openOverflowId.value = "";
  router.push(`/files${item.file_path}`);
};

const onOpenFolder = (item: QueueItem) => {
  openOverflowId.value = "";
  const dir = item.file_path.split("/").slice(0, -1).join("/") || "/";
  window.open(`/files${dir}`, "_blank");
};

const onCrumbClick = () => {
  expanded.value = true;
  // Jump to the page that contains the running row.
  const idx = cnc.queue.findIndex((q) => q.id === runningRow.value?.id);
  if (idx >= 0) page.value = Math.floor(idx / pageSize);
};

defineExpose({ toggle });
</script>

<style scoped>
.m-queue {
  background: var(--alt-background, #fafafa);
  border: 1px solid var(--border-color, #ddd);
  border-radius: 6px;
  overflow: visible;
  flex-shrink: 0;
}
.m-queue__header {
  display: flex;
  align-items: center;
  padding: 4px 10px;
  gap: 8px;
  cursor: pointer;
  min-width: 0;
}
.m-queue__chevron { color: var(--fg-muted, #888); font-size: 10px; flex-shrink: 0; }
.m-queue__title { font-size: 11px; font-weight: 500; color: var(--textPrimary, #222); flex-shrink: 0; }
.m-queue__count { font-size: 10px; color: var(--fg-muted, #888); flex-shrink: 0; }
.m-vert-divider { width: 1px; height: 14px; background: var(--border-color, #ddd); flex-shrink: 0; margin: 0 4px; }

.m-running-crumb {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 2px 8px;
  border-radius: 4px;
  background: rgba(33, 150, 243, 0.12);
  color: #1976d2;
  font-size: 10px;
  cursor: pointer;
  min-width: 0;
  flex: 1;
}
.m-running-crumb__dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #185FA5;
  flex-shrink: 0;
  animation: pulse 1.6s ease-in-out infinite;
}
@keyframes pulse { 0%,100% { opacity: 1; } 50% { opacity: 0.4; } }
.m-running-crumb__label { font-weight: 500; flex-shrink: 0; font-size: 10px; }
.m-running-crumb__sep { opacity: 0.4; flex-shrink: 0; }
.m-running-crumb__text {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  min-width: 0;
  font-size: 10px;
}
.m-running-crumb__progress {
  color: #1976d2;
  opacity: 0.75;
  font-variant-numeric: tabular-nums;
  font-size: 9px;
  flex-shrink: 0;
  margin-left: auto;
  padding-left: 8px;
}

.m-pager {
  margin-left: auto;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 9px;
  color: var(--fg-muted, #888);
}
.m-pager__btn {
  background: transparent;
  border: 1px solid var(--border-color, #ddd);
  border-radius: 3px;
  padding: 0 6px;
  font-size: 11px;
  cursor: pointer;
}
.m-pager__btn:disabled { opacity: 0.4; cursor: not-allowed; }

.m-queue__body {
  border-top: 1px solid var(--border-color, #eee);
  padding: 4px 8px 6px;
}
.m-queue__empty {
  padding: 8px;
  font-size: 11px;
  color: var(--fg-muted, #888);
  font-style: italic;
}
.m-queue__rows {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.m-queue-row {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 6px;
  border: 1px solid transparent;
  border-radius: 4px;
  position: relative;
  background: var(--surface, #fff);
  font-size: 11px;
}
.m-queue-row--sending,
.m-queue-row--running {
  border-color: rgba(33, 150, 243, 0.4);
  background: rgba(33, 150, 243, 0.06);
}
.m-queue-row--locked .m-queue-row__btn--send,
.m-queue-row--locked .m-queue-row__btn--auto { opacity: 0.4; cursor: not-allowed; }

.m-queue-row__handle { color: var(--fg-muted, #aaa); cursor: grab; user-select: none; font-size: 10px; }
.m-queue-row--sending .m-queue-row__handle,
.m-queue-row--running .m-queue-row__handle { opacity: 0.4; cursor: default; }
.m-queue-row__num { color: var(--fg-muted, #888); width: 14px; text-align: right; font-variant-numeric: tabular-nums; font-size: 10px; }
.m-queue-row__name { min-width: 0; flex: 1; display: flex; flex-direction: column; gap: 1px; }
.m-queue-row__title {
  font-weight: 500;
  font-size: 11px;
  color: var(--textPrimary, #222);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.m-queue-row__meta {
  font-size: 9px;
  color: var(--fg-muted, #888);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.m-queue-row__btn {
  background: transparent;
  border: 1px solid var(--border-color, #ddd);
  border-radius: 3px;
  width: 22px;
  height: 22px;
  font-size: 11px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--textSecondary, #555);
}
.m-queue-row__btn:hover:not(:disabled) { background: var(--alt-background, #f4f4f4); }
.m-queue-row__btn--send { color: #1976d2; }
.m-queue-row__btn--auto { color: #b85100; }
.m-queue-row__btn--cancel { color: #c0392b; }
.m-queue-row__spacer { flex: 1 1 0; }
.m-queue-row__send-opt {
  padding: 2px 8px;
  border: 1px solid var(--border-color, #ddd);
  border-radius: 3px;
  background: transparent;
  font-size: 10px;
  font-weight: 500;
  cursor: pointer;
  color: var(--textPrimary, #222);
}
.m-queue-row__send-opt--mem {
  border-color: #639922;
  color: #4d7717;
}
.m-queue-row__spinner {
  width: 10px;
  height: 10px;
  border: 2px solid rgba(33,150,243,0.25);
  border-top-color: #1976d2;
  border-radius: 50%;
  animation: m-spin 0.8s linear infinite;
  display: inline-block;
}
@keyframes m-spin { to { transform: rotate(360deg); } }

.m-overflow-menu {
  position: absolute;
  right: 4px;
  top: calc(100% + 2px);
  background: var(--surface, #fff);
  border: 1px solid var(--border-color, #ddd);
  border-radius: 4px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.08);
  z-index: 6;
  min-width: 200px;
  padding: 4px;
}
.m-overflow__item {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
  padding: 6px 8px;
  font-size: 11px;
  background: transparent;
  border: 0;
  cursor: pointer;
  border-radius: 3px;
  color: var(--textPrimary, #222);
  text-align: left;
}
.m-overflow__item:hover { background: var(--alt-background, #f4f4f4); }
.m-overflow__item .material-icons { font-size: 14px; }
.m-overflow__item--danger { color: #c0392b; }
.m-overflow__sep { height: 1px; background: var(--border-color, #eee); margin: 2px 4px; }
</style>
