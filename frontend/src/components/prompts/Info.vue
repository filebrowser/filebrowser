<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.fileInfo") }}</h2>
    </div>

    <div class="card-content">
      <p v-if="selected.length > 1">
        {{ $t("prompts.filesSelected", { count: selected.length }) }}
      </p>

      <p class="break-word" v-if="selected.length < 2">
        <strong>{{ $t("prompts.displayName") }}</strong> {{ name }}
      </p>

      <p v-if="!dir || selected.length > 1">
        <strong>{{ $t("prompts.size") }}:</strong>
        <span id="content_length"></span> {{ humanSize }}
      </p>

      <div v-if="resolution">
        <strong>{{ $t("prompts.resolution") }}:</strong>
        {{ resolution.width }} x {{ resolution.height }}
      </div>

      <p v-if="selected.length < 2" :title="modTime">
        <strong>{{ $t("prompts.lastModified") }}:</strong> {{ humanTime }}
      </p>

      <template v-if="dir && selected.length === 0">
        <p>
          <strong>{{ $t("prompts.numberFiles") }}:</strong> {{ req?.numFiles }}
        </p>
        <p>
          <strong>{{ $t("prompts.numberDirs") }}:</strong> {{ req?.numDirs }}
        </p>
      </template>

      <template v-if="!dir">
        <p>
          <strong>MD5: </strong
          ><code
            ><a
              @click="checksum($event, 'md5')"
              @keypress.enter="checksum($event, 'md5')"
              tabindex="2"
              >{{ $t("prompts.show") }}</a
            ></code
          >
        </p>
        <p>
          <strong>SHA1: </strong
          ><code
            ><a
              @click="checksum($event, 'sha1')"
              @keypress.enter="checksum($event, 'sha1')"
              tabindex="3"
              >{{ $t("prompts.show") }}</a
            ></code
          >
        </p>
        <p>
          <strong>SHA256: </strong
          ><code
            ><a
              @click="checksum($event, 'sha256')"
              @keypress.enter="checksum($event, 'sha256')"
              tabindex="4"
              >{{ $t("prompts.show") }}</a
            ></code
          >
        </p>
        <p>
          <strong>SHA512: </strong
          ><code
            ><a
              @click="checksum($event, 'sha512')"
              @keypress.enter="checksum($event, 'sha512')"
              tabindex="5"
              >{{ $t("prompts.show") }}</a
            ></code
          >
        </p>
      </template>
    </div>

    <div class="card-action">
      <button
        id="focus-prompt"
        type="submit"
        @click="closeHovers"
        class="button button--flat"
        :aria-label="$t('buttons.ok')"
        :title="$t('buttons.ok')"
      >
        {{ $t("buttons.ok") }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject } from "vue";
import { storeToRefs } from "pinia";
import { useRoute } from "vue-router";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { filesize } from "@/utils";
import dayjs from "dayjs";
import { files as api } from "@/api";

const route = useRoute();
const $showError = inject<(error: unknown) => void>("$showError");

const fileStore = useFileStore();
const layoutStore = useLayoutStore();

const { req, selected, selectedCount, isListing } = storeToRefs(fileStore);
const { closeHovers } = layoutStore;

const humanSize = computed(() => {
  if (selectedCount.value === 0 || !isListing.value) {
    return filesize(req.value?.size ?? 0);
  }

  let sum = 0;

  for (const selectedIdx of selected.value) {
    sum += req.value?.items[selectedIdx]?.size ?? 0;
  }

  return filesize(sum);
});

const humanTime = computed(() => {
  if (selectedCount.value === 0) {
    return dayjs(req.value?.modified).fromNow();
  }

  return dayjs(req.value?.items[selected.value[0]]?.modified).fromNow();
});

const modTime = computed(() => {
  if (selectedCount.value === 0) {
    return new Date(Date.parse(req.value?.modified ?? "")).toLocaleString();
  }

  return new Date(
    Date.parse(req.value?.items[selected.value[0]]?.modified ?? "")
  ).toLocaleString();
});

const name = computed(() => {
  return selectedCount.value === 0
    ? (req.value?.name ?? "")
    : (req.value?.items[selected.value[0]]?.name ?? "");
});

const dir = computed(() => {
  return (
    selectedCount.value > 1 ||
    (selectedCount.value === 0
      ? (req.value?.isDir ?? false)
      : (req.value?.items[selected.value[0]]?.isDir ?? false))
  );
});

const resolution = computed(() => {
  if (selectedCount.value === 1) {
    const selectedItem = req.value?.items[selected.value[0]];
    if (selectedItem && selectedItem.type === "image") {
      return (selectedItem as any).resolution;
    }
  } else if (req.value && req.value.type === "image") {
    return (req.value as any).resolution;
  }
  return null;
});

const checksum = async (event: Event, algo: string) => {
  event.preventDefault();

  let link;

  if (selectedCount.value) {
    link = req.value?.items[selected.value[0]]?.url ?? "";
  } else {
    link = route.path;
  }

  try {
    const hash = await api.checksum(link, algo as any);
    (event.target as HTMLElement).textContent = hash;
  } catch (e) {
    $showError?.(e);
  }
};
</script>
