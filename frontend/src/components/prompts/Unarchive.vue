<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ t("prompts.unarchive") }}</h2>
    </div>

    <div class="card-content">
      <form ref="unarchiveForm">
        <p>{{ t("prompts.unarchiveFolderNameMessage") }}</p>
        <input
          class="input input--block"
          v-focus
          type="text"
          @keyup.enter="submit"
          v-model.trim="name"
          required
        />
      </form>

      <p>{{ t("prompts.unarchiveDestinationLocationMessage") }}</p>
      <file-list @update:selected="(val: any) => (dest = val)"></file-list>

      <p v-if="overwriteAvailable">
        <input type="checkbox" v-model="overwriteExisting" />
        {{ t("prompts.unarchiveOverwriteExisting") }}
      </p>
    </div>

    <div class="card-action">
      <button
        class="button button--flat button--grey"
        @click="cancel"
        :aria-label="t('buttons.cancel')"
        :title="t('buttons.cancel')"
      >
        {{ t("buttons.cancel") }}
      </button>
      <button
        @click="submit"
        class="button button--flat"
        type="submit"
        :aria-label="t('buttons.unarchive')"
        :title="t('buttons.unarchive')"
      >
        {{ t("buttons.unarchive") }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { inject, ref } from "vue";
import { useI18n } from "vue-i18n";
import { files as api } from "@/api";
import { useFileStore } from "@/stores/file";
import { useQuotaStore } from "@/stores/quota";
import { useLayoutStore } from "@/stores/layout";
import { computed } from "vue";
import FileList from "./FileList.vue";
import buttons from "@/utils/buttons";

const fileStore = useFileStore();
const quotaStore = useQuotaStore();
const layoutStore = useLayoutStore();

const { t } = useI18n();

const $showError = inject<IToastError>("$showError")!;

let overwriteExisting = false;
let dest: string | null = null;
let name = "";

const unarchiveForm = ref<HTMLFormElement | null>(null);

const overwriteAvailable = computed((): boolean => {
  let item = fileStore.req?.items[fileStore.selected[0]];
  if (!item) {
    return false;
  }

  return [".zip", ".rar", ".tar", ".bz2", ".gz", ".xz", ".lz4", ".sz"].includes(
    item.extension
  );
});

const cancel = () => {
  layoutStore.closeHovers();
};

const submit = async () => {
  if (!unarchiveForm.value?.reportValidity()) {
    return;
  }

  let item = fileStore.req?.items[fileStore.selected[0]];
  if (!item) {
    return;
  }

  let dst = dest + encodeURIComponent(name);

  try {
    buttons.loading("unarchive");
    layoutStore.closeHovers();
    await api.unarchive(item.url, dst, overwriteExisting);
    fileStore.reload = true;
    quotaStore.fetchQuota(3000);
  } catch (e: any) {
    $showError(e);
  } finally {
    buttons.done("unarchive");
  }
};
</script>
