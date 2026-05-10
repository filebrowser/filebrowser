<template>
  <section class="machine-card tool-card">
    <div class="card-header">
      <i class="material-icons">build</i>
      {{ t("toolTable.title") }}
      <span class="card-header__spacer" />
      <span v-if="reading" class="tool-card__hint">
        {{ t("toolTable.readingHint", { slots: slotCount }) }}
      </span>
      <span v-else-if="latest" class="tool-card__hint">
        {{ t("toolTable.lastReadHint", { ts: fmtTs(latest.read_at) }) }}
      </span>
      <label class="tool-card__slots-input" :title="t('toolTable.slotCountTitle')">
        {{ t("toolTable.slotsToRead") }}
        <input
          type="number"
          min="1"
          max="200"
          v-model.number="slotCount"
          @change="persistSlotCount"
        />
      </label>
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
        <span>{{ fmtMs(latest.duration_ms) }}</span>
        <span>·</span>
        <span>{{ latest.bridge_address }}</span>
        <span v-if="folder">·</span>
        <a v-if="folder" :href="folderHref" class="tool-card__folder">
          {{ t("toolTable.history") }}
        </a>
      </div>

      <div v-if="latest" class="tool-card__filterbar">
        <input
          class="tool-card__search"
          type="search"
          v-model="search"
          :placeholder="t('toolTable.searchPlaceholder')"
        />
        <label class="tool-card__toggle">
          <input type="checkbox" v-model="showOffline" />
          {{ t("toolTable.showOffline") }}
        </label>
        <label class="tool-card__toggle">
          <input type="checkbox" v-model="showEmpty" />
          {{ t("toolTable.showEmpty") }}
        </label>
        <span class="tool-card__visible-count">
          {{ t("toolTable.visibleCount", { n: displayRows.length, total: latest.slots.length }) }}
        </span>
      </div>

      <div v-if="latest" class="tool-card__scroll">
        <table class="tool-table">
          <thead>
            <tr>
              <th class="thumb-col"></th>
              <th
                v-for="col in columns"
                :key="col.key"
                :class="['sortable', { num: col.num }, { sorted: sortKey === col.key }]"
                @click="toggleSort(col.key)"
              >
                {{ col.label }}
                <span v-if="sortKey === col.key" class="sort-arrow">
                  {{ sortDir === 'asc' ? '↑' : '↓' }}
                </span>
              </th>
              <th>{{ t("toolTable.status") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="row in displayRows"
              :key="row.slot"
              :class="rowClass(row)"
            >
              <td class="thumb-col">
                <ToolGeometryView
                  :slot-number="row.slot"
                  :length-ratio="ratioL(row.effective_length)"
                  :diameter-ratio="ratioD(row.effective_diameter)"
                  :width="24"
                  :height="44"
                />
              </td>
              <td>{{ row.slot }}</td>
              <td class="num">{{ fmt(row.length_geom) }}</td>
              <td class="num">{{ fmt(row.length_wear) }}</td>
              <td class="num">{{ fmt(row.effective_length) }}</td>
              <td class="num">{{ fmt(row.diameter_geom) }}</td>
              <td class="num">{{ fmt(row.diameter_wear) }}</td>
              <td class="num">{{ fmt(row.effective_diameter) }}</td>
              <td class="status">
                <span v-if="hasErr(row)" class="badge badge--err" :title="firstErr(row)">
                  {{ t("toolTable.offline") }}
                </span>
                <span v-else-if="row.empty" class="badge badge--empty">
                  {{ t("toolTable.emptyPocket") }}
                </span>
                <span v-else class="badge badge--ok">
                  {{ t("toolTable.loaded") }}
                </span>
              </td>
            </tr>
            <tr v-if="displayRows.length === 0">
              <td colspan="9" class="tool-card__no-match">
                {{ t("toolTable.noMatch") }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Magazine: side-by-side scale view of every loaded tool. All
           bodies share one length and diameter scale so spotting a
           mis-loaded slot is instant — the sticking-out-too-far tool
           jumps out visually before you scan the numbers. -->
      <div v-if="latest && loadedSlots.length > 0" class="tool-card__magazine">
        <button class="link-btn" @click="showMagazine = !showMagazine">
          {{ showMagazine ? t("toolTable.hideMagazine") : t("toolTable.showMagazine", { n: loadedSlots.length }) }}
        </button>
        <div v-if="showMagazine" class="magazine-strip">
          <div
            v-for="row in loadedSlots"
            :key="row.slot"
            class="magazine-figure"
            :title="`T${row.slot} — ⌀${fmt(row.effective_diameter)} L${fmt(row.effective_length)}`"
          >
            <ToolGeometryView
              :slot-number="row.slot"
              :length-ratio="ratioL(row.effective_length)"
              :diameter-ratio="ratioD(row.effective_diameter)"
              :width="48"
              :height="180"
            />
            <div class="magazine-figure__label">
              <strong>T{{ row.slot }}</strong>
              <span>⌀ {{ fmt(row.effective_diameter) }}</span>
              <span>L {{ fmt(row.effective_length) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { cnc as cncApi } from "@/api";
import type { ToolTable, ToolTableSlot } from "@/api/cnc";
import ToolGeometryView from "@/components/ToolGeometryView.vue";

const props = defineProps<{
  machineId?: string;
  cncRunning: boolean;
}>();

const { t } = useI18n();

const SLOT_COUNT_KEY = "cncToolTableSlotCount";
const slotCount = ref<number>(
  (() => {
    const v = parseInt(localStorage.getItem(SLOT_COUNT_KEY) || "");
    return Number.isFinite(v) && v >= 1 && v <= 200 ? v : 200;
  })()
);
const persistSlotCount = () => {
  if (slotCount.value < 1) slotCount.value = 1;
  if (slotCount.value > 200) slotCount.value = 200;
  localStorage.setItem(SLOT_COUNT_KEY, String(slotCount.value));
};

const latest = ref<ToolTable | null>(null);
const reading = ref(false);
const errMsg = ref<string>("");
const folder = ref<string>("");

const search = ref("");
const showOffline = ref(true);
const showEmpty = ref(false);
const showMagazine = ref(false);
type SortKey =
  | "slot"
  | "length_geom"
  | "length_wear"
  | "effective_length"
  | "diameter_geom"
  | "diameter_wear"
  | "effective_diameter";
const sortKey = ref<SortKey>("slot");
const sortDir = ref<"asc" | "desc">("asc");

const columns = computed(() => [
  { key: "slot" as const, label: t("toolTable.slot"), num: false },
  { key: "length_geom" as const, label: t("toolTable.lengthGeom"), num: true },
  { key: "length_wear" as const, label: t("toolTable.lengthWear"), num: true },
  { key: "effective_length" as const, label: t("toolTable.lengthEff"), num: true },
  { key: "diameter_geom" as const, label: t("toolTable.diameterGeom"), num: true },
  { key: "diameter_wear" as const, label: t("toolTable.diameterWear"), num: true },
  { key: "effective_diameter" as const, label: t("toolTable.diameterEff"), num: true },
]);

const folderHref = computed(() =>
  folder.value ? `/files${folder.value}` : ""
);

const hasErr = (r: ToolTableSlot) =>
  !!r.errors && Object.keys(r.errors).length > 0 &&
  r.length_geom === undefined && r.length_wear === undefined &&
  r.diameter_geom === undefined && r.diameter_wear === undefined;
const firstErr = (r: ToolTableSlot) => {
  if (!r.errors) return "";
  const k = Object.keys(r.errors)[0];
  return `${k}: ${r.errors[k]}`;
};

const fmt = (v: number | undefined) => {
  if (v === undefined || v === null) return "—";
  return v.toFixed(4);
};

const fmtMs = (ms: number) => {
  if (ms < 1000) return `${Math.round(ms)} ms`;
  if (ms < 60_000) return `${(ms / 1000).toFixed(1)} s`;
  const m = Math.floor(ms / 60_000);
  const s = Math.round((ms % 60_000) / 1000);
  return `${m} m ${s} s`;
};

const fmtTs = (iso: string) => {
  if (!iso) return "—";
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
};

const rowClass = (r: ToolTableSlot) => ({
  "is-empty": r.empty && !hasErr(r),
  "is-offline": hasErr(r),
});

// Loaded tools only — empty pockets and offline slots have no useful
// geometry to render. The magazine view + per-row thumbnails both
// scale against the loaded population, so missing/empty rows render
// as the dashed placeholder instead of a misleadingly-tiny shape.
const loadedSlots = computed(() => {
  if (!latest.value) return [];
  return latest.value.slots.filter(
    (r) => !r.empty && !hasErr(r) &&
      typeof r.effective_length === "number" &&
      typeof r.effective_diameter === "number"
  );
});

// Magazine-wide max length / diameter — used as the denominator for
// the scaling ratios so all tools render against one shared scale.
// Floor at a small positive so a single-tool magazine still draws.
const maxLen = computed(() => {
  let m = 0;
  for (const r of loadedSlots.value) {
    if (typeof r.effective_length === "number" && r.effective_length > m) {
      m = r.effective_length;
    }
  }
  return m > 0 ? m : 1;
});
const maxDia = computed(() => {
  let m = 0;
  for (const r of loadedSlots.value) {
    if (typeof r.effective_diameter === "number" && r.effective_diameter > m) {
      m = r.effective_diameter;
    }
  }
  return m > 0 ? m : 0.5;
});

const ratioL = (len: number | undefined) =>
  typeof len === "number" && len > 0 ? len / maxLen.value : undefined;
const ratioD = (dia: number | undefined) =>
  typeof dia === "number" && dia > 0 ? dia / maxDia.value : undefined;

const displayRows = computed(() => {
  if (!latest.value) return [];
  let rows = latest.value.slots.slice();

  // Status filters first (cheap).
  rows = rows.filter((r) => {
    const offline = hasErr(r);
    const empty = r.empty && !offline;
    const loaded = !offline && !empty;
    if (offline && !showOffline.value) return false;
    if (empty && !showEmpty.value) return false;
    return loaded || (offline && showOffline.value) || (empty && showEmpty.value);
  });

  // Free-text search across slot number and any numeric value (matches
  // partial number strings — typing "0.5" finds 0.5011, etc).
  const q = search.value.trim().toLowerCase();
  if (q) {
    rows = rows.filter((r) => {
      if (String(r.slot).includes(q)) return true;
      const fields: (number | undefined)[] = [
        r.length_geom,
        r.length_wear,
        r.effective_length,
        r.diameter_geom,
        r.diameter_wear,
        r.effective_diameter,
      ];
      return fields.some(
        (f) => typeof f === "number" && f.toFixed(4).includes(q)
      );
    });
  }

  // Sort.
  const k = sortKey.value;
  rows.sort((a: any, b: any) => {
    const va = a[k];
    const vb = b[k];
    // undefined sorts last regardless of direction so empty pockets
    // group at the bottom.
    if (va === undefined && vb === undefined) return 0;
    if (va === undefined) return 1;
    if (vb === undefined) return -1;
    if (va === vb) return 0;
    const cmp = va < vb ? -1 : 1;
    return sortDir.value === "asc" ? cmp : -cmp;
  });

  return rows;
});

const toggleSort = (k: SortKey) => {
  if (sortKey.value === k) {
    sortDir.value = sortDir.value === "asc" ? "desc" : "asc";
  } else {
    sortKey.value = k;
    sortDir.value = "asc";
  }
};

const loadLatest = async () => {
  errMsg.value = "";
  try {
    const tbl = await cncApi.getLatestToolTable(props.machineId);
    latest.value = tbl;
  } catch (e: any) {
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
    const env = await cncApi.readToolTable(slotCount.value, props.machineId);
    latest.value = env.table;
    const parts: string[] = [];
    if (env.read_error) parts.push(`${t("toolTable.readPartial")}: ${env.read_error}`);
    if (env.persist_error) parts.push(`${t("toolTable.persistFailed")}: ${env.persist_error}`);
    if (parts.length) errMsg.value = parts.join(" · ");
    cncApi.getToolTableHistory(props.machineId)
      .then((h) => { folder.value = h.folder || ""; })
      .catch(() => {});
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
/* Span both columns of the /machine grid so the table has room to
   breathe — at 200 slots the dense column layout is the load-bearing
   feature. Falls back to single-column on narrow viewports via the
   parent grid's media query. */
.tool-card {
  grid-column: 1 / -1;
}

.tool-card__hint {
  font-size: 0.78rem;
  color: var(--fg-muted, #888);
  margin-right: 0.6rem;
}

.tool-card__slots-input {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.78rem;
  color: var(--fg-muted, #888);
  margin-right: 0.6rem;
}

.tool-card__slots-input input {
  width: 4.5rem;
  padding: 0.2rem 0.3rem;
  font-size: 0.85rem;
  border: 1px solid var(--border-color, #ddd);
  border-radius: 4px;
  background: var(--surface, #fff);
  color: inherit;
}

.tool-card__body {
  flex-direction: column;
  align-items: stretch;
  padding: 0.8rem 1rem;
  gap: 0.6rem;
  overflow: hidden;
}

.tool-card__err {
  padding: 0.5rem 0.7rem;
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
}

.tool-card__folder {
  color: var(--primaryColor, #2196f3);
  text-decoration: none;
}

.tool-card__folder:hover {
  text-decoration: underline;
}

.tool-card__filterbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.6rem;
  padding-bottom: 0.4rem;
  border-bottom: 1px solid var(--border-color, #eee);
}

.tool-card__search {
  flex: 1 1 12rem;
  min-width: 12rem;
  padding: 0.35rem 0.6rem;
  border: 1px solid var(--border-color, #ddd);
  border-radius: 4px;
  background: var(--surface, #fff);
  color: inherit;
  font-size: 0.85rem;
}

.tool-card__toggle {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.8rem;
  color: var(--fg-muted, #666);
  cursor: pointer;
  user-select: none;
}

.tool-card__visible-count {
  margin-left: auto;
  font-size: 0.78rem;
  color: var(--fg-muted, #888);
}

.tool-card__scroll {
  flex: 1 1 auto;
  max-height: 50vh;
  overflow: auto;
  border: 1px solid var(--border-color, #eee);
  border-radius: 4px;
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

.tool-table thead th {
  position: sticky;
  top: 0;
  background: var(--alt-background, #fafafa);
  z-index: 1;
  font-weight: 500;
  color: var(--fg-muted, #888);
  text-transform: uppercase;
  font-size: 0.7rem;
  letter-spacing: 0.04em;
  cursor: pointer;
  user-select: none;
}

.tool-table th.sortable:hover {
  color: var(--fg, #333);
}

.tool-table th.sorted {
  color: var(--primaryColor, #2196f3);
}

.sort-arrow {
  margin-left: 0.2rem;
  font-size: 0.75rem;
}

.tool-table td.num,
.tool-table th.num {
  text-align: right;
  font-variant-numeric: tabular-nums;
}

.tool-table tr.is-empty td {
  color: var(--fg-muted, #999);
}

.tool-table tr.is-offline td {
  color: var(--fg-muted, #888);
  background: rgba(198, 40, 40, 0.04);
}

.badge {
  display: inline-block;
  padding: 0.1rem 0.45rem;
  border-radius: 999px;
  font-size: 0.7rem;
  font-weight: 500;
  letter-spacing: 0.02em;
}

.badge--ok {
  background: rgba(46, 125, 50, 0.12);
  color: #2e7d32;
}

.badge--empty {
  background: rgba(158, 158, 158, 0.18);
  color: #757575;
}

.badge--err {
  background: rgba(198, 40, 40, 0.12);
  color: #c62828;
  cursor: help;
}

.tool-card__no-match {
  text-align: center;
  padding: 1rem 0;
  color: var(--fg-muted, #888);
}

.thumb-col {
  width: 28px;
  text-align: center;
  padding: 0.2rem 0.3rem;
}

/* Magazine: tools rendered to scale on one shared scale. Bottom-
   align so the cutting-tip reference line is consistent — long
   tools hang out further at the top. */
.tool-card__magazine {
  margin-top: 0.6rem;
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.magazine-strip {
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem;
  padding: 0.8rem 0.6rem 0.4rem;
  background: var(--alt-background, #fafafa);
  border-radius: 6px;
  border: 1px solid var(--border-color, #eee);
  align-items: flex-end;
}

.magazine-figure {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.2rem;
  padding: 0.2rem 0.3rem;
  border-radius: 4px;
  cursor: help;
}

.magazine-figure:hover {
  background: var(--surface-hover, rgba(33, 150, 243, 0.05));
}

.magazine-figure__label {
  display: flex;
  flex-direction: column;
  align-items: center;
  font-size: 0.7rem;
  color: var(--fg-muted, #666);
  font-variant-numeric: tabular-nums;
}

.magazine-figure__label strong {
  color: var(--fg, #333);
  font-size: 0.78rem;
}
</style>
