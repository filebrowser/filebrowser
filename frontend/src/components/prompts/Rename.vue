<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.rename") }}</h2>
    </div>

    <div class="card-content">
      <p>
        {{ $t("prompts.renameMessage") }} <code>{{ oldName() }}</code
        >:
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

<script setup lang="ts">
import { ref, onMounted, inject } from "vue";
import { storeToRefs } from "pinia";
import { useRouter } from "vue-router";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import url from "@/utils/url";
import { files as api } from "@/api";
import { removePrefix } from "@/api/utils";

const router = useRouter();
const $showError = inject<(error: unknown) => void>("$showError");

const fileStore = useFileStore();
const layoutStore = useLayoutStore();

const { req, selected, selectedCount, isListing } = storeToRefs(fileStore);
const { closeHovers } = layoutStore;

const name = ref("");

const oldName = (): string => {
  if (!isListing.value) {
    return req.value?.name ?? "";
  }

  if (selectedCount.value === 0 || selectedCount.value > 1) {
    // This shouldn't happen.
    return "";
  }

  return req.value?.items[selected.value[0]].name ?? "";
};

onMounted(() => {
  name.value = oldName();
});

const submit = async () => {
  let oldLink = "";
  let newLink = "";

  if (!req.value) {
    return;
  }

  if (!isListing.value) {
    oldLink = req.value.url;
  } else {
    oldLink = req.value.items[selected.value[0]].url;
  }

  newLink = url.removeLastDir(oldLink) + "/" + encodeURIComponent(name.value);

  try {
    await api.move([{ from: oldLink, to: newLink }]);
    if (!isListing.value) {
      router.push({ path: newLink });
      return;
    }

    fileStore.preselect = removePrefix(newLink);

    fileStore.reload = true;
  } catch (e) {
    $showError?.(e);
  }

  closeHovers();
};
</script>
