<template>
  <div class="csv-viewer">
    <div class="csv-header">
      <div class="header-select">
        <label for="columnSeparator">{{ $t("files.columnSeparator") }}</label>
        <select
          id="columnSeparator"
          class="input input--block"
          v-model="columnSeparator"
        >
          <option :value="[',']">
            {{ $t("files.csvSeparators.comma") }}
          </option>
          <option :value="[';']">
            {{ $t("files.csvSeparators.semicolon") }}
          </option>
          <option :value="[',', ';']">
            {{ $t("files.csvSeparators.both") }}
          </option>
        </select>
      </div>
      <div class="header-select" v-if="isEncodedContent">
        <label for="fileEncoding">{{ $t("files.fileEncoding") }}</label>
        <DropdownModal
          v-model="isEncondingDropdownOpen"
          :close-on-click="false"
        >
          <div>
            <span class="selected-encoding">{{ selectedEncoding }}</span>
          </div>
          <template v-slot:list>
            <input
              v-model="encodingSearch"
              :placeholder="$t('search.search')"
              class="input input--block"
              name="encoding"
            />
            <div class="encoding-list">
              <div v-if="encodingList.length == 0" class="message">
                <i class="material-icons">sentiment_dissatisfied</i>
                <span>{{ $t("files.lonely") }}</span>
              </div>
              <button
                v-for="encoding in encodingList"
                :value="encoding"
                :key="encoding"
                class="encoding-button"
                @click="selectedEncoding = encoding"
              >
                {{ encoding }}
              </button>
            </div>
          </template>
        </DropdownModal>
      </div>
    </div>
    <div v-if="displayError" class="csv-error">
      <i class="material-icons">error</i>
      <p>{{ displayError }}</p>
    </div>
    <div v-else-if="parsed.headers.length === 0" class="csv-empty">
      <i class="material-icons">description</i>
      <p>{{ $t("files.lonely") }}</p>
    </div>
    <div v-else class="csv-table-container" @wheel.stop @touchmove.stop>
      <table class="csv-table">
        <thead>
          <tr>
            <th v-for="(header, index) in parsed.headers" :key="index">
              {{ header || `Column ${index + 1}` }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, rowIndex) in parsed.rows" :key="rowIndex">
            <td v-for="(cell, cellIndex) in row" :key="cellIndex">
              {{ cell }}
            </td>
          </tr>
        </tbody>
      </table>
      <div class="csv-footer">
        <div class="csv-info" v-if="parsed.rows.length > 100">
          <i class="material-icons">info</i>
          <span>
            {{ $t("files.showingRows", { count: parsed.rows.length }) }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, watchEffect } from "vue";
import { parse } from "csv-parse/browser/esm";
import { useI18n } from "vue-i18n";
import { availableEncodings, decode } from "@/utils/encodings";
import DropdownModal from "../DropdownModal.vue";

const { t } = useI18n({});

interface Props {
  content: ArrayBuffer | string;
  error?: string;
}

const props = withDefaults(defineProps<Props>(), {
  error: "",
});

const isEncondingDropdownOpen = ref(false);

const encodingSearch = ref<string>("");

const encodingList = computed(() => {
  return availableEncodings.filter((e) =>
    e.toLowerCase().includes(encodingSearch.value.toLowerCase())
  );
});

const columnSeparator = ref([",", ";"]);

const selectedEncoding = ref("utf-8");

const parsed = ref<CsvData>({ headers: [], rows: [] });

const displayError = ref<string | null>(null);

const isEncodedContent = computed(() => {
  return props.content instanceof ArrayBuffer;
});

watchEffect(() => {
  if (props.content !== "" && columnSeparator.value.length > 0) {
    const content = isEncodedContent.value
      ? decode(props.content as ArrayBuffer, selectedEncoding.value)
      : props.content;
    parse(
      content as string,
      { delimiter: columnSeparator.value, skip_empty_lines: true },
      (error, output) => {
        if (error) {
          console.error("Failed to parse CSV:", error);
          parsed.value = { headers: [], rows: [] };
          displayError.value = t("files.csvLoadFailed", {
            error: error.toString(),
          });
        } else {
          parsed.value = {
            headers: output[0],
            rows: output.slice(1),
          };
          displayError.value = null;
        }
      }
    );
  }
});

watch(selectedEncoding, () => {
  isEncondingDropdownOpen.value = false;
  encodingSearch.value = "";
});
</script>

<style scoped>
.csv-viewer {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: var(--surfacePrimary);
  color: var(--textSecondary);
  padding: 1rem;
  padding-top: 4em;
  box-sizing: border-box;
}

.csv-error,
.csv-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: 1rem;
  color: var(--textPrimary);
}

.csv-error i,
.csv-empty i {
  font-size: 4rem;
  opacity: 0.5;
}

.csv-error p,
.csv-empty p {
  font-size: 1.1rem;
  margin: 0;
}

.csv-table-container {
  flex: 1;
  overflow: auto;
  background-color: var(--surfacePrimary);
  border-radius: 4px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

/* Scrollbar styling for better visibility */
.csv-table-container::-webkit-scrollbar {
  width: 12px;
  height: 12px;
}

.csv-table-container::-webkit-scrollbar-track {
  background: var(--background);
  border-radius: 4px;
}

.csv-table-container::-webkit-scrollbar-thumb {
  background: var(--borderSecondary);
  border-radius: 4px;
}

.csv-table-container::-webkit-scrollbar-thumb:hover {
  background: var(--textPrimary);
}

.csv-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.875rem;
  background-color: var(--surfacePrimary);
}

.csv-table thead {
  position: sticky;
  top: 0;
  z-index: 10;
  background-color: var(--surfaceSecondary);
}

.csv-table th {
  padding: 0.875rem 1rem;
  text-align: left;
  font-weight: 600;
  border-bottom: 2px solid var(--borderSecondary);
  background-color: var(--surfaceSecondary);
  white-space: nowrap;
  color: var(--textSecondary);
  font-size: 0.875rem;
}

.csv-table td {
  padding: 0.75rem 1rem;
  text-align: left;
  border-bottom: 1px solid var(--borderPrimary);
  white-space: nowrap;
  max-width: 400px;
  overflow: hidden;
  text-overflow: ellipsis;
  color: var(--textSecondary);
}

.csv-table tbody tr:nth-child(even) {
  background-color: var(--background);
}

.csv-table tbody tr:hover {
  background-color: var(--hover);
  transition: background-color 0.15s ease;
}

.csv-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  padding: 0.5rem;
}

.csv-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  margin-top: 0.5rem;
  background-color: var(--surfaceSecondary);
  border-radius: 4px;
  border-left: 3px solid var(--blue);
  color: var(--textSecondary);
  font-size: 0.875rem;
}

.csv-header {
  display: flex;
  justify-content: space-between;
  padding: 0.25rem;
}

.header-select {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
  flex-direction: column;
  @media (width >= 640px) {
    flex-direction: row;
  }
}

.header-select > label {
  font-size: small;
  @media (width >= 640px) {
    max-width: 70px;
  }
}

.header-select > select,
.header-select > div {
  margin-bottom: 0;
}

.csv-info i {
  font-size: 1.2rem;
  color: var(--blue);
}

.encoding-list {
  max-height: 300px;
  min-width: 120px;
  overflow: auto;
  overscroll-behavior: contain;
  -webkit-overflow-scrolling: touch;
  touch-action: pan-y;
}

.encoding-button {
  background-color: transparent;
  border: none;
  outline: none;
  padding: 0.25rem 0.5rem;
  color: var(--textPrimary);
  text-align: left;
  cursor: pointer;
  border-radius: 0.2rem;
  white-space: nowrap;
  display: block;
  width: 100%;
}

.encoding-button:hover {
  background-color: var(--surfaceSecondary);
}

.selected-encoding {
  white-space: nowrap;
  text-overflow: ellipsis;
}

.message {
  font-size: 1.25em;
}
</style>
