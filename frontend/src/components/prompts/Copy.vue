<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.copy") }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t("prompts.copyMessage") }}</p>
      <file-list
        ref="fileList"
        @update:selected="(val: string) => (dest = val)"
        tabindex="1"
      />
    </div>

    <div
      class="card-action"
      style="display: flex; align-items: center; justify-content: space-between"
    >
      <template v-if="user?.perm.create">
        <button
          class="button button--flat"
          @click="fileList?.createDir()"
          :aria-label="$t('sidebar.newFolder')"
          :title="$t('sidebar.newFolder')"
          style="justify-self: left"
        >
          <span>{{ $t("sidebar.newFolder") }}</span>
        </button>
      </template>
      <div>
        <button
          class="button button--flat button--grey"
          @click="closeHovers"
          :aria-label="$t('buttons.cancel')"
          :title="$t('buttons.cancel')"
          tabindex="3"
        >
          {{ $t("buttons.cancel") }}
        </button>
        <button
          id="focus-prompt"
          class="button button--flat"
          @click="copy"
          :aria-label="$t('buttons.copy')"
          :title="$t('buttons.copy')"
          tabindex="2"
        >
          {{ $t("buttons.copy") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, inject } from "vue";
import { storeToRefs } from "pinia";
import { useRoute, useRouter } from "vue-router";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { useAuthStore } from "@/stores/auth";
import FileList from "./FileList.vue";
import { files as api } from "@/api";
import buttons from "@/utils/buttons";
import * as upload from "@/utils/upload";
import { removePrefix } from "@/api/utils";

const route = useRoute();
const router = useRouter();
const $showError = inject<(error: unknown) => void>("$showError");

const fileStore = useFileStore();
const layoutStore = useLayoutStore();
const authStore = useAuthStore();

const { req, selected } = storeToRefs(fileStore);
const { user } = storeToRefs(authStore);
const { showHover, closeHovers } = layoutStore;

const fileList = ref<InstanceType<typeof FileList> | null>(null);
const dest = ref<string | null>(null);

const copy = async (event: Event) => {
  event.preventDefault();
  const items: Array<{ from: string; to: string; name: string }> = [];

  // Create a new promise for each file.
  for (const item of selected.value) {
    items.push({
      from: req.value!.items[item].url,
      to: dest.value! + encodeURIComponent(req.value!.items[item].name),
      name: req.value!.items[item].name,
    });
  }

  const action = async (overwrite: boolean, rename: boolean) => {
    buttons.loading("copy");

    await api
      .copy(items, overwrite, rename)
      .then(() => {
        buttons.success("copy");
        fileStore.preselect = removePrefix(items[0].to);

        if (route.path === dest.value) {
          fileStore.reload = true;
          return;
        }

        router.push({ path: dest.value! });
      })
      .catch((e) => {
        buttons.done("copy");
        $showError?.(e);
      });
  };

  if (route.path === dest.value) {
    closeHovers();
    action(false, true);
    return;
  }

  const dstItems = (await api.fetch(dest.value!)).items;
  const conflict = upload.checkConflict(items as any, dstItems);

  let overwrite = false;
  let rename = false;

  if (conflict) {
    showHover({
      prompt: "replace-rename",
      confirm: (event: Event, option: string) => {
        overwrite = option == "overwrite";
        rename = option == "rename";

        event.preventDefault();
        closeHovers();
        action(overwrite, rename);
      },
    });

    return;
  }

  action(overwrite, rename);
};
</script>
