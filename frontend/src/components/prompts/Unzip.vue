<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.unzip") }}</h2>
    </div>

    <div class="card-content">
      <p>
        {{ loading ? $t("prompts.extractingZip") : $t("prompts.unzipMessage") }}
      </p>
      <div v-if="!loading">
        <file-list
          ref="fileListRef"
          @update:selected="(val: string) => (dest = val)"
          tabindex="1"
        />
        <p>
          <input type="checkbox" v-model="overwrite" />
          {{ $t("prompts.unzipOverwrite") }}
        </p>
      </div>
      <div v-else class="animation-wrapper">
        <div id="loading">
          <div class="spinner">
            <div class="bounce1"></div>
            <div class="bounce2"></div>
            <div class="bounce3"></div>
          </div>
        </div>
      </div>
    </div>

    <div
      class="card-action"
      style="display: flex; align-items: center; justify-content: space-between"
    >
      <template v-if="user != null && user.perm.create">
        <button
          class="button button--flat"
          @click="fileListRef?.createDir()"
          :aria-label="$t('sidebar.newFolder')"
          :title="$t('sidebar.newFolder')"
          :disabled="loading"
          style="justify-self: left"
        >
          <span>{{ $t("sidebar.newFolder") }}</span>
        </button>
      </template>
      <div>
        <button
          class="button button--flat button--grey"
          @click="close"
          :aria-label="$t('buttons.cancel')"
          :title="$t('buttons.cancel')"
          :disabled="loading"
          tabindex="3"
        >
          {{ $t("buttons.cancel") }}
        </button>
        <button
          id="focus-prompt"
          class="button button--flat"
          @click="unzip"
          :aria-label="$t('buttons.unzip')"
          :title="$t('buttons.unzip')"
          :disabled="loading"
          tabindex="2"
        >
          {{ $t("buttons.unzip") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, inject } from "vue";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { useAuthStore } from "@/stores/auth";
import FileList from "./FileList.vue";
import { files as api } from "@/api";
import { useRouter } from "vue-router";
import type { ComponentPublicInstance } from "vue";

const layoutStore = useLayoutStore();

const router = useRouter();

const $showError = inject<IToastError>("$showError")!;

const dest = ref<string | null>(null);
const overwrite = ref(false);
const loading = ref(false);

const fileListRef = ref<ComponentPublicInstance<{
  createDir: () => void;
}> | null>(null);

const { req, selected } = useFileStore();
const { user } = useAuthStore();

async function unzip(event: Event) {
  event.preventDefault();
  if (
    req === null ||
    dest.value === null ||
    req.items.length === 0 ||
    selected.length === 0
  ) {
    return;
  }

  loading.value = true;
  try {
    const zipFilePath = req.items[selected[0]].url;
    await api.unzip(zipFilePath, dest.value, overwrite.value);
    router.push(dest.value);
  } catch (e) {
    if (e instanceof Error && $showError) {
      $showError(e);
    }
  } finally {
    layoutStore.closeHovers();
    loading.value = false;
  }
}

const close = () => {
  layoutStore.closeHovers();
};
</script>

<style scoped>
.animation-wrapper {
  position: relative;
  height: 30px;
  width: 100%;
}

.animation-wrapper > #loading {
  position: relative;
  height: 30px;
  background: transparent;
}
.animation-wrapper > #loading > .spinner {
  top: 9rem;
}
</style>
