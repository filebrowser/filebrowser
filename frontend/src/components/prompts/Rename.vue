<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.rename") }}</h2>
    </div>

    <div class="card-content">
      <p>
        {{ $t("prompts.renameMessage") }} <code>{{ oldName() }}</code>:
      </p>
      <input
        id="focus-prompt"
        class="input input--block"
        type="text"
        @keyup.enter="submit"
        v-model.trim="name"
      />
    </div>

    <div class="card-action">
      <button
        class="button button--flat button--grey"
        @click="closeHovers"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')"
      >
        {{ $t("buttons.cancel") }}
      </button>
      <button
        @click="submit"
        class="button button--flat"
        type="submit"
        :aria-label="$t('buttons.rename')"
        :title="$t('buttons.rename')"
      >
        {{ $t("buttons.rename") }}
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, inject, onMounted } from "vue";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import url from "@/utils/url";
import { files as api } from "@/api";
import { removePrefix } from "@/api/utils";
import { useRouter } from "vue-router";
import { useI18n } from "vue-i18n";

const { t } = useI18n();
const $showError = inject("$showError");
const router = useRouter();

const fileStore = useFileStore();
const layoutStore = useLayoutStore();

const name = ref("");

const oldName = () => {
  if (!fileStore.isListing) {
    return fileStore.req.name;
  }

  if (fileStore.selectedCount === 0 || fileStore.selectedCount > 1) {
    return;
  }

  return fileStore.req.items[fileStore.selected[0]].name;
};

onMounted(() => {
  name.value = oldName();
});

const submit = async () => {
  if (name.value.includes("/")) {
    $showError(new Error(t("errors.invalidName")));
    layoutStore.closeHovers();
    return;
  }

  let oldLink = "";
  let newLink = "";

  if (!fileStore.isListing) {
    oldLink = fileStore.req.url;
  } else {
    oldLink = fileStore.req.items[fileStore.selected[0]].url;
  }

  newLink = url.removeLastDir(oldLink) + "/" + encodeURIComponent(name.value);

  try {
    await api.move([{ from: oldLink, to: newLink }]);

    if (!fileStore.isListing) {
      router.push({ path: newLink });
      return;
    }

    fileStore.preselect = removePrefix(newLink);
    fileStore.reload = true;
  } catch (e) {
    $showError(e);
  }

  layoutStore.closeHovers();
};

const closeHovers = layoutStore.closeHovers;
</script>
