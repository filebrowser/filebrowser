<template>
  <div class="card floating" id="convertx-convert">
    <div class="card-title">
      <h2>{{ t("prompts.convert") }}</h2>
    </div>

    <div class="card-content">
      <p>
        {{ t("prompts.convertMessage") }}
        <code>{{ file?.name }}</code>
      </p>

      <label for="focus-prompt">{{ t("prompts.convertSearch") }}</label>
      <input
        id="focus-prompt"
        class="input input--block"
        type="search"
        autocomplete="off"
        :placeholder="t('prompts.convertSearchPlaceholder')"
        v-model.trim="filter"
      />

      <label for="convertx-target">{{ t("prompts.convertTarget") }}</label>
      <select
        id="convertx-target"
        class="input input--block"
        v-model="selectedKey"
        :disabled="filteredOptions.length === 0"
      >
        <option
          v-for="option in filteredOptions"
          :key="option.key"
          :value="option.key"
        >
          {{ option.convertTo }} — {{ option.converter }}
        </option>
      </select>

      <p v-if="selectedOption" class="small">
        {{ t("prompts.convertOutput") }}
        <code>{{ outputPreview }}</code>
      </p>
      <p v-else-if="options.length > 0" class="small">
        {{ t("prompts.convertNoMatchingTargets") }}
      </p>
      <p v-else class="small">
        {{ t("prompts.convertNoTargets") }}
      </p>
    </div>

    <div class="card-action">
      <button
        class="button button--flat button--grey"
        @click="layoutStore.closeHovers"
        :aria-label="t('buttons.cancel')"
        :title="t('buttons.cancel')"
      >
        {{ t("buttons.cancel") }}
      </button>
      <button
        class="button button--flat"
        type="submit"
        @click="submit"
        :aria-label="t('buttons.convert')"
        :title="t('buttons.convert')"
        :disabled="!selectedOption"
      >
        {{ t("buttons.convert") }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useLayoutStore } from "@/stores/layout";

interface ConvertOption {
  key: string;
  converter: string;
  convertTo: string;
}

const layoutStore = useLayoutStore();
const { t } = useI18n();

const props = computed(() => layoutStore.currentPrompt?.props || {});
const file = computed<ResourceItem | null>(() => props.value.file || null);
const options = computed<ConvertOption[]>(() => props.value.options || []);
const selectedKey = ref("");
const filter = ref("");

const filteredOptions = computed(() => {
  const query = filter.value.trim().toLowerCase();
  if (!query) return options.value;

  return options.value.filter((option) => {
    const target = option.convertTo.toLowerCase();
    const converter = option.converter.toLowerCase();
    const label = `${target} ${converter}`;

    return (
      target.includes(query) ||
      converter.includes(query) ||
      label.includes(query)
    );
  });
});

watch(
  filteredOptions,
  (currentOptions) => {
    if (currentOptions.length === 0) {
      selectedKey.value = "";
      return;
    }

    if (!currentOptions.some((option) => option.key === selectedKey.value)) {
      selectedKey.value = currentOptions[0].key;
    }
  },
  { immediate: true }
);

const selectedOption = computed(() =>
  filteredOptions.value.find((option) => option.key === selectedKey.value)
);

const outputPreview = computed(() => {
  if (!file.value || !selectedOption.value) return "";

  const name = file.value.name || "converted";
  const dot = name.lastIndexOf(".");
  const base = dot > 0 ? name.slice(0, dot) : name;
  return `${base}.${selectedOption.value.convertTo}`;
});

const submit = () => {
  if (!selectedOption.value) return;

  layoutStore.currentPrompt?.confirm({
    converter: selectedOption.value.converter,
    convertTo: selectedOption.value.convertTo,
  });
};
</script>
