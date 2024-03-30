<template>
  <div id="search" @click="open" v-bind:class="{ active, ongoing }">
    <div id="input">
      <button
        v-if="active"
        class="action"
        @click="close"
        :aria-label="$t('buttons.close')"
        :title="$t('buttons.close')"
      >
        <i class="material-icons">arrow_back</i>
      </button>
      <i v-else class="material-icons">search</i>
      <input
        type="text"
        @keyup.exact="keyup"
        @keyup.enter="submit"
        ref="input"
        :autofocus="active"
        v-model.trim="prompt"
        :aria-label="$t('search.search')"
        :placeholder="$t('search.search')"
      />
    </div>

    <div id="result" ref="result">
      <div>
        <template v-if="isEmpty">
          <p>{{ text }}</p>

          <template v-if="prompt.length === 0">
            <div class="boxes">
              <h3>{{ $t("search.types") }}</h3>
              <div>
                <div
                  tabindex="0"
                  v-for="(v, k) in boxes"
                  :key="k"
                  role="button"
                  @click="init('type:' + k)"
                  :aria-label="$t('search.' + v.label)"
                >
                  <i class="material-icons">{{ v.icon }}</i>
                  <p>{{ $t("search." + v.label) }}</p>
                </div>
              </div>
            </div>
          </template>
        </template>
        <ul v-show="results.length > 0">
          <li v-for="(s, k) in filteredResults" :key="k">
            <router-link v-on:click="close" :to="s.url">
              <i v-if="s.dir" class="material-icons">folder</i>
              <i v-else class="material-icons">insert_drive_file</i>
              <span>./{{ s.path }}</span>
            </router-link>
          </li>
        </ul>
      </div>
      <p id="renew">
        <i class="material-icons spin">autorenew</i>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

import url from "@/utils/url";
import { search } from "@/api";
import { computed, inject, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";
import { storeToRefs } from "pinia";

const boxes = {
  image: { label: "images", icon: "insert_photo" },
  audio: { label: "music", icon: "volume_up" },
  video: { label: "video", icon: "movie" },
  pdf: { label: "pdf", icon: "picture_as_pdf" },
};

const layoutStore = useLayoutStore();
const fileStore = useFileStore();

const { currentPromptName } = storeToRefs(layoutStore);

const prompt = ref<string>("");
const active = ref<boolean>(false);
const ongoing = ref<boolean>(false);
const results = ref<any[]>([]);
const reload = ref<boolean>(false);
const resultsCount = ref<number>(50);

const $showError = inject<IToastError>("$showError")!;

const input = ref<HTMLInputElement | null>(null);
const result = ref<HTMLElement | null>(null);

const { t } = useI18n();

const route = useRoute();

watch(currentPromptName, (newVal, oldVal) => {
  active.value = newVal === "search";

  if (oldVal === "search" && !active.value) {
    if (reload.value) {
      fileStore.reload = true;
    }

    document.body.style.overflow = "auto";
    reset();
    prompt.value = "";
    active.value = false;
    input.value?.blur();
  } else if (active.value) {
    reload.value = false;
    input.value?.focus();
    document.body.style.overflow = "hidden";
  }
});

watch(prompt, () => {
  if (results.value.length) {
    reset();
  }
});

// ...mapState(useFileStore, ["isListing"]),
// ...mapState(useLayoutStore, ["show"]),
// ...mapWritableState(useFileStore, { sReload: "reload" }),

const isEmpty = computed(() => {
  return results.value.length === 0;
});
const text = computed(() => {
  if (ongoing.value) {
    return "";
  }

  return prompt.value === ""
    ? t("search.typeToSearch")
    : t("search.pressToSearch");
});
const filteredResults = computed(() => {
  return results.value.slice(0, resultsCount.value);
});

onMounted(() => {
  if (result.value === null) {
    return;
  }
  result.value.addEventListener("scroll", (event: Event) => {
    if (
      (event.target as HTMLElement).offsetHeight +
        (event.target as HTMLElement).scrollTop >=
      (event.target as HTMLElement).scrollHeight - 100
    ) {
      resultsCount.value += 50;
    }
  });
});

const open = () => {
  !active.value && layoutStore.showHover("search");
};

const close = (event: Event) => {
  event.stopPropagation();
  event.preventDefault();
  layoutStore.closeHovers();
};

const keyup = (event: KeyboardEvent) => {
  if (event.key === "Escape") {
    close(event);
    return;
  }
  results.value.length = 0;
};

const init = (string: string) => {
  prompt.value = `${string} `;
  input.value !== null ? input.value.focus() : "";
};

const reset = () => {
  ongoing.value = false;
  resultsCount.value = 50;
  results.value = [];
};

const submit = async (event: Event) => {
  event.preventDefault();

  if (prompt.value === "") {
    return;
  }

  let path = route.path;
  if (!fileStore.isListing) {
    path = url.removeLastDir(path) + "/";
  }

  ongoing.value = true;

  try {
    results.value = await search(path, prompt.value);
  } catch (error: any) {
    $showError(error);
  }

  ongoing.value = false;
};
</script>
