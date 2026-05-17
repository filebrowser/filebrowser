<template>
  <div class="m-modal-backdrop" @click.self="$emit('close')">
    <div class="m-modal m-modal--narrow">
      <div class="m-modal__header">
        <span class="m-modal__title">
          {{ t("toolTable.editTitle", { slot: slotNum }) }}
        </span>
        <span class="m-modal__spacer" />
        <button class="m-modal__close" @click="$emit('close')">✕</button>
      </div>
      <div class="m-modal__body">
        <!-- Why two paths: per-field for "I have hand-measured values"
             and copy-from for "this slot is the same tool as that one".
             They're mutually exclusive — picking a copy-from source
             greys out the numeric inputs and clobbers their effects on
             save. -->
        <p class="m-edit__intro">{{ t("toolTable.editIntro") }}</p>

        <fieldset class="m-edit__group">
          <legend>{{ t("toolTable.editCopyFromLabel") }}</legend>
          <select v-model.number="copyFromSlot" class="m-edit__select">
            <option :value="0">{{ t("toolTable.editCopyFromNone") }}</option>
            <option
              v-for="src in candidateSources"
              :key="src.slot"
              :value="src.slot"
            >
              T{{ src.slot }} — ⌀{{ fmt(src.effective_diameter) }} L{{ fmt(src.effective_length) }}
            </option>
          </select>
          <p v-if="copyFromSlot !== 0" class="m-edit__hint">
            {{ t("toolTable.editCopyFromHint", { src: copyFromSlot, dst: slotNum }) }}
          </p>
        </fieldset>

        <fieldset class="m-edit__group" :disabled="copyFromSlot !== 0">
          <legend>{{ t("toolTable.editFieldsLabel") }}</legend>
          <div class="m-edit__row">
            <label>
              <span>{{ t("toolTable.lengthGeom") }}</span>
              <input
                v-model.number="lengthGeom"
                type="number"
                step="0.0001"
                inputmode="decimal"
              />
            </label>
            <label>
              <span>{{ t("toolTable.lengthWear") }}</span>
              <input
                v-model.number="lengthWear"
                type="number"
                step="0.0001"
                inputmode="decimal"
              />
            </label>
          </div>
          <div class="m-edit__row">
            <label>
              <span>{{ t("toolTable.diameterGeom") }}</span>
              <input
                v-model.number="diameterGeom"
                type="number"
                step="0.0001"
                inputmode="decimal"
              />
            </label>
            <label>
              <span>{{ t("toolTable.diameterWear") }}</span>
              <input
                v-model.number="diameterWear"
                type="number"
                step="0.0001"
                inputmode="decimal"
              />
            </label>
          </div>
        </fieldset>

        <div v-if="errMsg" class="m-modal__err">{{ errMsg }}</div>

        <div class="m-edit__actions">
          <button class="m-modal__btn" @click="$emit('close')" :disabled="saving">
            {{ t("toolTable.editCancel") }}
          </button>
          <button
            class="m-modal__btn m-modal__btn--primary"
            :disabled="saving || !canSave"
            @click="save"
          >
            {{ saving ? t("toolTable.editSaving") : t("toolTable.editSave") }}
          </button>
        </div>
        <p class="m-edit__footnote">{{ t("toolTable.editFootnote") }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useI18n } from "vue-i18n";
import { cnc as cncApi } from "@/api";
import type { ToolTableSlot } from "@/api/cnc";

const props = defineProps<{
  slotNum: number;
  machineId?: string;
  current?: ToolTableSlot;
  candidates: ToolTableSlot[];
}>();

const emit = defineEmits<{
  (e: "close"): void;
  (e: "saved"): void;
}>();

const { t } = useI18n();

// Pre-fill from the current row so the operator only re-types the
// numbers they're actually changing.
const lengthGeom = ref<number | undefined>(props.current?.length_geom);
const lengthWear = ref<number | undefined>(props.current?.length_wear);
const diameterGeom = ref<number | undefined>(props.current?.diameter_geom);
const diameterWear = ref<number | undefined>(props.current?.diameter_wear);
const copyFromSlot = ref<number>(0);

const saving = ref(false);
const errMsg = ref("");

// Only offer copy-from sources that are loaded (have at least one
// offset) and exclude the row being edited.
const candidateSources = computed(() =>
  props.candidates.filter(
    (c) =>
      c.slot !== props.slotNum &&
      (c.length_geom !== undefined ||
        c.length_wear !== undefined ||
        c.diameter_geom !== undefined ||
        c.diameter_wear !== undefined),
  ),
);

const canSave = computed(() => {
  if (copyFromSlot.value > 0) return true;
  return (
    lengthGeom.value !== undefined ||
    lengthWear.value !== undefined ||
    diameterGeom.value !== undefined ||
    diameterWear.value !== undefined
  );
});

const fmt = (v?: number) => (typeof v === "number" ? v.toFixed(4) : "—");

const save = async () => {
  if (saving.value || !canSave.value) return;
  saving.value = true;
  errMsg.value = "";
  try {
    const opts: Parameters<typeof cncApi.editToolTableSlot>[0] = {
      slot: props.slotNum,
      machineId: props.machineId,
    };
    // copy-from path wins; in that mode the per-field inputs are
    // disabled so the values would be stale anyway. Server enforces
    // mutual exclusion too.
    if (copyFromSlot.value > 0) {
      opts.copyFromSlot = copyFromSlot.value;
    } else {
      // Only send fields the operator actually filled — the backend
      // leaves any omitted field alone. Empty-string inputs come back
      // as NaN from v-model.number; treat them as "no change."
      if (typeof lengthGeom.value === "number" && !Number.isNaN(lengthGeom.value)) {
        opts.lengthGeom = lengthGeom.value;
      }
      if (typeof lengthWear.value === "number" && !Number.isNaN(lengthWear.value)) {
        opts.lengthWear = lengthWear.value;
      }
      if (typeof diameterGeom.value === "number" && !Number.isNaN(diameterGeom.value)) {
        opts.diameterGeom = diameterGeom.value;
      }
      if (typeof diameterWear.value === "number" && !Number.isNaN(diameterWear.value)) {
        opts.diameterWear = diameterWear.value;
      }
    }
    await cncApi.editToolTableSlot(opts);
    emit("saved");
    emit("close");
  } catch (e: any) {
    errMsg.value = e?.message || String(e);
  } finally {
    saving.value = false;
  }
};
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
  width: min(440px, 92vw);
  max-height: 80vh;
  background: var(--surface, #fff);
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 12px 32px rgba(0,0,0,0.25);
}
.m-modal--narrow { width: min(440px, 92vw); }
.m-modal__header {
  display: flex;
  align-items: center;
  padding: 8px 10px;
  border-bottom: 1px solid var(--border-color, #eee);
}
.m-modal__title { font-weight: 500; font-size: 13px; }
.m-modal__spacer { flex: 1 1 0; }
.m-modal__close {
  background: transparent;
  border: 0;
  font-size: 16px;
  cursor: pointer;
  color: var(--fg-muted, #888);
  padding: 4px 8px;
}
.m-modal__body { padding: 12px 14px; overflow: auto; }
.m-modal__err {
  margin-top: 8px;
  padding: 6px 8px;
  background: rgba(198, 40, 40, 0.08);
  color: #c62828;
  font-size: 11px;
  border-radius: 4px;
}
.m-modal__btn {
  padding: 6px 14px;
  background: var(--alt-background, #f4f4f4);
  border: 1px solid var(--border-color, #ddd);
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
}
.m-modal__btn:hover:not(:disabled) { filter: brightness(0.97); }
.m-modal__btn:disabled { opacity: 0.55; cursor: not-allowed; }
.m-modal__btn--primary {
  background: #185FA5;
  border-color: #14507f;
  color: #fff;
}

.m-edit__intro {
  margin: 0 0 12px;
  font-size: 12px;
  color: var(--fg-muted, #555);
}
.m-edit__group {
  border: 1px solid var(--border-color, #ddd);
  border-radius: 4px;
  padding: 8px 10px;
  margin: 0 0 10px;
}
.m-edit__group legend {
  padding: 0 4px;
  font-size: 11px;
  font-weight: 500;
  color: var(--fg-muted, #555);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}
.m-edit__group:disabled {
  opacity: 0.5;
}
.m-edit__select {
  width: 100%;
  padding: 4px 6px;
  border: 1px solid var(--border-color, #ddd);
  border-radius: 3px;
  background: var(--surface, #fff);
  color: inherit;
  font: inherit;
}
.m-edit__hint {
  margin: 6px 0 0;
  font-size: 11px;
  color: #185FA5;
}
.m-edit__row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  margin-top: 6px;
}
.m-edit__row label {
  display: flex;
  flex-direction: column;
  gap: 2px;
  font-size: 11px;
  color: var(--fg-muted, #555);
}
.m-edit__row input {
  padding: 4px 6px;
  border: 1px solid var(--border-color, #ddd);
  border-radius: 3px;
  background: var(--surface, #fff);
  font: inherit;
  font-variant-numeric: tabular-nums;
}
.m-edit__actions {
  display: flex;
  justify-content: flex-end;
  gap: 6px;
  margin-top: 12px;
}
.m-edit__footnote {
  margin: 10px 0 0;
  font-size: 10px;
  color: var(--fg-muted, #888);
}
</style>
