<template>
  <div id="search" @click="open" v-bind:class="{ active, ongoing }">
    <div id="input">
      <button
        v-if="active"
        class="action"
        @click="close"
        :aria-label="closeButtonTitle"
        :title="closeButtonTitle"
      >
        <i v-if="ongoing" class="material-icons">stop_circle</i>
        <i v-else class="material-icons">arrow_back</i>
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
      <i
        v-show="ongoing"
        class="material-icons spin"
        style="display: inline-block"
        >autorenew
      </i>
      <span style="margin-top: 5px" v-show="results.length > 0">
        {{ results.length }}
      </span>
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
    </div>
  </div>
</template>

<script setup lang="ts">
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

import url from "@/utils/url";
import { search } from "@/api";
import { computed, inject, onMounted, ref, watch, onUnmounted } from "vue";
import { useI18n } from "vue-i18n";
import { useRoute } from "vue-router";
import { storeToRefs } from "pinia";
import { StatusError } from "@/api/utils";

const boxes = {
  image: { label: "images", icon: "insert_photo" },
  audio: { label: "music", icon: "volume_up" },
  video: { label: "video", icon: "movie" },
  pdf: { label: "pdf", icon: "picture_as_pdf" },
};

const layoutStore = useLayoutStore();
const fileStore = useFileStore();
let searchAbortController = new AbortController();

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
  reset();
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

const closeButtonTitle = computed(() => {
  return ongoing.value ? t("buttons.stopSearch") : t("buttons.close");
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

onUnmounted(() => {
  abortLastSearch();
});

const open = () => {
  !active.value && layoutStore.showHover("search");
};

const close = (event: Event) => {
  if (ongoing.value) {
    abortLastSearch();
    ongoing.value = false;
  } else {
    event.stopPropagation();
    event.preventDefault();
    layoutStore.closeHovers();
  }
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
  abortLastSearch();
  ongoing.value = false;
  resultsCount.value = 50;
  results.value = [];
};

const abortLastSearch = () => {
  searchAbortController.abort();
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
    abortLastSearch();
    searchAbortController = new AbortController();
    results.value = [];
    await search(path, prompt.value, searchAbortController.signal, (item) =>
      results.value.push(item)
    );
  } catch (error: any) {
    if (error instanceof StatusError && error.is_canceled) {
      return;
    }
    $showError(error);
  }

  ongoing.value = false;
};
</script>
