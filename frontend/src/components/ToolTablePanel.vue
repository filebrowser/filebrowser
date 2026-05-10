<template>
  <section class="machine-card tool-card">
    <div class="card-header">
      <i class="material-icons">build</i>
      {{ t("toolTable.title") }}
      <span class="card-header__spacer" />
      <span v-if="reading" class="tool-card__hint">
        {{ t("toolTable.readingHint", { slots }) }}
      </span>
      <span v-else-if="latest" class="tool-card__hint">
        {{ t("toolTable.lastReadHint", { ts: fmtTs(latest.read_at) }) }}
      </span>
      <button
        class="check-btn"
        :disabled="reading || cncRunning"
        @click="runRead"
        :title="t('toolTable.readNowTitle')"
      >
        <i class="material-icons">refresh</i>
        {{ reading ? t("toolTable.reading") : t("toolTable.readNow") }}
      </button>
    </div>
    <div class="card-body tool-card__body">
      <div v-if="errMsg" class="tool-card__err">{{ errMsg }}</div>
      <div v-if="!latest && !reading" class="tool-card__empty">
        {{ t("toolTable.empty") }}
      </div>

      <div v-if="latest" class="tool-card__meta">
        <span>{{ latest.slots_read }} / {{ latest.slots_requested }} {{ t("toolTable.slotsLabel") }}</span>
        <span>·</span>
        <span>{{ Math.round(latest.duration_ms) }} ms</span>
        <span>·</span>
        <span>{{ latest.bridge_address }}</span>
        <span v-if="folder">·</span>
        <a v-if="folder" :href="folderHref" class="tool-card__folder">
          {{ t("toolTable.history") }}
        </a>
      </div>

      <table v-if="latest" class="tool-table">
        <thead>
          <tr>
            <th>{{ t("toolTable.slot") }}</th>
            <th class="num">{{ t("toolTable.lengthGeom") }}</th>
            <th class="num">{{ t("toolTable.lengthWear") }}</th>
            <th class="num">{{ t("toolTable.lengthEff") }}</th>
            <th class="num">{{ t("toolTable.diameterGeom") }}</th>
            <th class="num">{{ t("toolTable.diameterWear") }}</th>
            <th class="num">{{ t("toolTable.diameterEff") }}</th>
            <th>{{ t("toolTable.notes") }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="row in displayRows"
            :key="row.slot"
            :class="{ 'is-empty': row.empty, 'has-err': hasErr(row) }"
          >
            <td>{{ row.slot }}</td>
            <td class="num">{{ fmt(row.length_geom) }}</td>
            <td class="num">{{ fmt(row.length_wear) }}</td>
            <td class="num">{{ fmt(row.effective_length) }}</td>
            <td class="num">{{ fmt(row.diameter_geom) }}</td>
            <td class="num">{{ fmt(row.diameter_wear) }}</td>
            <td class="num">{{ fmt(row.effective_diameter) }}</td>
            <td class="notes">
              <span v-if="row.empty">{{ t("toolTable.emptyPocket") }}</span>
              <span v-else-if="hasErr(row)" class="err">
                {{ firstErr(row) }}
              </span>
            </td>
          </tr>
        </tbody>
      </table>

      <div v-if="latest && hidableEmpty" class="tool-card__toggle">
        <button class="link-btn" @click="hideEmpty = !hideEmpty">
          {{ hideEmpty ? t("toolTable.showAll") : t("toolTable.hideEmpty") }}
        </button>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { cnc as cncApi } from "@/api";
import type { ToolTable, ToolTableSlot } from "@/api/cnc";

const props = defineProps<{
  machineId?: string;
  cncRunning: boolean;
  slots?: number;
}>();

const { t } = useI18n();

const slots = computed(() => props.slots ?? 30);
const latest = ref<ToolTable | null>(null);
const reading = ref(false);
const errMsg = ref<string>("");
const folder = ref<string>("");
const hideEmpty = ref(true);

const folderHref = computed(() =>
  folder.value ? `/files${folder.value}` : ""
);

const hasErr = (r: ToolTableSlot) => r.errors && Object.keys(r.errors).length > 0;
const firstErr = (r: ToolTableSlot) => {
  if (!r.errors) return "";
  const k = Object.keys(r.errors)[0];
  return `${k}: ${r.errors[k]}`;
};

const fmt = (v: number | undefined) => {
  if (v === undefined || v === null) return "—";
  return v.toFixed(4);
};

// "Empty" rows are pockets that read cleanly but had every value at
// 0.0 — i.e. no tool loaded. Hiding them by default keeps the visible
// table to the populated slots, which is what the operator usually
// wants. Toggle reveals all so they can verify the read covered every
// requested slot.
const hidableEmpty = computed(() => {
  if (!latest.value) return false;
  return latest.value.slots.some((r) => r.empty) &&
    latest.value.slots.some((r) => !r.empty);
});

const displayRows = computed(() => {
  if (!latest.value) return [];
  if (!hideEmpty.value) return latest.value.slots;
  return latest.value.slots.filter((r) => !r.empty || hasErr(r));
});

const fmtTs = (iso: string) => {
  if (!iso) return "—";
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
};

const loadLatest = async () => {
  errMsg.value = "";
  try {
    const tbl = await cncApi.getLatestToolTable(props.machineId);
    latest.value = tbl;
  } catch (e: any) {
    // Latest is best-effort — only surface non-trivial errors.
    if (e?.status && e.status !== 404 && e.status !== 204) {
      errMsg.value = e.message || String(e);
    }
  }
  try {
    const hist = await cncApi.getToolTableHistory(props.machineId);
    folder.value = hist.folder || "";
  } catch {
    /* ignore */
  }
};

const runRead = async () => {
  if (reading.value) return;
  reading.value = true;
  errMsg.value = "";
  try {
    const env = await cncApi.readToolTable(slots.value, props.machineId);
    latest.value = env.table;
    if (env.persist_error) {
      errMsg.value = `${t("toolTable.persistFailed")}: ${env.persist_error}`;
    }
    // Refresh history listing so the new dump shows in the folder
    // link's destination immediately.
    cncApi.getToolTableHistory(props.machineId).then((h) => {
      folder.value = h.folder || "";
    }).catch(() => {});
  } catch (e: any) {
    errMsg.value = e?.message || String(e);
  } finally {
    reading.value = false;
  }
};

watch(
  () => props.machineId,
  () => {
    latest.value = null;
    loadLatest();
  }
);

onMounted(loadLatest);
</script>

<style scoped>
.tool-card__hint {
  font-size: 0.78rem;
  color: var(--fg-muted, #888);
  margin-right: 0.6rem;
}

.tool-card__body {
  padding: 0.8rem 1rem;
}

.tool-card__err {
  padding: 0.5rem 0.7rem;
  margin-bottom: 0.6rem;
  border-radius: 4px;
  background: rgba(198, 40, 40, 0.1);
  color: #c62828;
  font-size: 0.85rem;
}

.tool-card__empty {
  padding: 0.6rem 0;
  color: var(--fg-muted, #888);
  font-size: 0.9rem;
}

.tool-card__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  font-size: 0.78rem;
  color: var(--fg-muted, #888);
  margin-bottom: 0.6rem;
}

.tool-card__folder {
  color: var(--primaryColor, #2196f3);
  text-decoration: none;
}

.tool-card__folder:hover {
  text-decoration: underline;
}

.tool-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.85rem;
}

.tool-table th,
.tool-table td {
  padding: 0.3rem 0.5rem;
  text-align: left;
  border-bottom: 1px solid var(--border-color, #eee);
}

.tool-table th {
  font-weight: 500;
  color: var(--fg-muted, #888);
  text-transform: uppercase;
  font-size: 0.7rem;
  letter-spacing: 0.04em;
}

.tool-table td.num,
.tool-table th.num {
  text-align: right;
  font-variant-numeric: tabular-nums;
}

.tool-table td.notes {
  font-size: 0.78rem;
  color: var(--fg-muted, #888);
}

.tool-table tr.has-err td.notes .err {
  color: #c62828;
}

.tool-table tr.is-empty td {
  color: var(--fg-muted, #999);
}

.tool-card__toggle {
  margin-top: 0.6rem;
  text-align: center;
}

.link-btn {
  background: none;
  border: 0;
  color: var(--primaryColor, #2196f3);
  cursor: pointer;
  font-size: 0.85rem;
  padding: 0.2rem 0.4rem;
}

.link-btn:hover {
  text-decoration: underline;
}
</style>
