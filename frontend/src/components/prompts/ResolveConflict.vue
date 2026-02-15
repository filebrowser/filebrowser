<template>
  <div class="card floating">
    <div class="card-title">
      <h2>
        {{
          personalized
            ? $t("prompts.resolveConflict")
            : $t("prompts.replaceOrSkip")
        }}
      </h2>
    </div>

    <div class="card-content">
      <template v-if="personalized">
        <p v-if="isUploadAction != true">
          {{ $t("prompts.singleConflictResolve") }}
        </p>
        <div class="conflict-list-container">
          <div>
            <p>
              <input
                @change="toogleCheckAll"
                type="checkbox"
                :checked="originAllChecked"
                value="origin"
              />
              {{
                isUploadAction != true
                  ? $t("prompts.filesInOrigin")
                  : $t("prompts.uploadingFiles")
              }}
            </p>
            <p>
              <input
                @change="toogleCheckAll"
                type="checkbox"
                :checked="destAllChecked"
                value="dest"
              />
              {{ $t("prompts.filesInDest") }}
            </p>
          </div>
          <div>
            <template v-for="(item, index) in conflict" :key="index">
              <div class="conflict-file-name">
                <span>{{ item.name }}</span>

                <template v-if="item.checked.length == 2">
                  <span v-if="isUploadAction != true" class="result-rename">
                    {{ $t("prompts.rename") }}
                  </span>
                  <span v-else class="result-error">
                    {{ $t("prompts.forbiddenError") }}
                  </span>
                </template>
                <span
                  v-else-if="
                    item.checked.length == 1 && item.checked[0] == 'origin'
                  "
                  class="result-override"
                >
                  {{ $t("prompts.override") }}
                </span>
                <span v-else class="result-skip">
                  {{ $t("prompts.skip") }}
                </span>
              </div>
              <div>
                <input v-model="item.checked" type="checkbox" value="origin" />
                <div>
                  <p class="conflict-file-value">
                    {{ humanTime(item.origin.lastModified) }}
                  </p>
                  <p class="conflict-file-value">
                    {{ humanSize(item.origin.size) }}
                  </p>
                </div>
              </div>
              <div>
                <input v-model="item.checked" type="checkbox" value="dest" />
                <div>
                  <p class="conflict-file-value">
                    {{ humanTime(item.dest.lastModified) }}
                  </p>
                  <p class="conflict-file-value">
                    {{ humanSize(item.dest.size) }}
                  </p>
                </div>
              </div>
            </template>
          </div>
        </div>
      </template>
      <template v-else>
        <p>
          {{ $t("prompts.fastConflictResolve", { count: conflict.length }) }}
        </p>

        <div class="result-buttons">
          <button @click="(e) => resolve(e, ['origin'])">
            <i class="material-icons">done_all</i>
            {{ $t("buttons.overrideAll") }}
          </button>
          <button
            v-if="isUploadAction != true"
            @click="(e) => resolve(e, ['origin', 'dest'])"
          >
            <i class="material-icons">folder_copy</i>
            {{ $t("buttons.renameAll") }}
          </button>
          <button @click="(e) => resolve(e, ['dest'])">
            <i class="material-icons">undo</i>
            {{ $t("buttons.skipAll") }}
          </button>
          <button @click="personalized = true">
            <i class="material-icons">checklist</i>
            {{ $t("buttons.singleDecision") }}
          </button>
        </div>
      </template>
    </div>

    <div class="card-action" style="display: flex; justify-content: end">
      <div>
        <button
          class="button button--flat button--grey"
          @click="close"
          :aria-label="$t('buttons.cancel')"
          :title="$t('buttons.cancel')"
          tabindex="4"
        >
          {{ $t("buttons.cancel") }}
        </button>
        <button
          v-if="personalized"
          id="focus-prompt"
          class="button button--flat"
          @click="(event) => currentPrompt?.confirm(event, conflict)"
          :aria-label="$t('buttons.ok')"
          :title="$t('buttons.ok')"
          tabindex="1"
        >
          {{ $t("buttons.ok") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { useLayoutStore } from "@/stores/layout";
import { filesize } from "@/utils";
import dayjs from "dayjs";

const layoutStore = useLayoutStore();
const { currentPrompt } = layoutStore;

const conflict = ref<ConflictingResource[]>(currentPrompt?.props.conflict);

const isUploadAction = ref<boolean | undefined>(
  currentPrompt?.props.isUploadAction
);

const personalized = ref(false);

const originAllChecked = computed(() => {
  for (const item of conflict.value) {
    if (!item.checked.includes("origin")) return false;
  }

  return true;
});

const destAllChecked = computed(() => {
  for (const item of conflict.value) {
    if (!item.checked.includes("dest")) return false;
  }

  return true;
});

const close = () => {
  layoutStore.closeHovers();
};

const humanSize = (size: number | undefined) => {
  return size == undefined ? "Unknown size" : filesize(size);
};

const humanTime = (modified: string | number | undefined) => {
  if (modified == undefined) return "Unknown date";

  return dayjs(modified).format("L LT");
};

const resolve = (event: Event, result: Array<"origin" | "dest">) => {
  for (const item of conflict.value) {
    item.checked = result;
  }
  currentPrompt?.confirm(event, conflict.value);
};

const toogleCheckAll = (e: Event) => {
  const target = e.currentTarget as HTMLInputElement;
  const value = target.value as "origin" | "dest" | "both";
  const checked = target.checked;

  for (const item of conflict.value) {
    if (value == "both") {
      item.checked = ["origin", "dest"];
    } else {
      if (!item.checked.includes(value)) {
        if (checked) {
          item.checked.push(value);
        }
      } else {
        if (!checked) {
          item.checked = value == "dest" ? ["origin"] : ["dest"];
        }
      }
    }
  }
};
</script>
<style scoped>
.conflict-list-container {
  max-height: 300px;
  overflow: auto;
}

.conflict-list-container > div {
  display: grid;
  grid-template-columns: 1fr 1fr;
  border-bottom: solid 1px var(--textPrimary);
  gap: 0.5rem 0.25rem;
}

.conflict-list-container > div:last-child {
  border-bottom: none;
}

.conflict-list-container > div > div {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.conflict-file-name {
  grid-column: 1 / -1;
  color: var(--textPrimary);
  font-size: 0.8rem;
  display: flex;
  justify-content: space-between;
  padding: 0.5rem 0.25rem;
}

.conflict-file-value {
  color: var(--textPrimary);
  font-size: 0.9rem;
  margin: 0;
}

.result-rename,
.result-override,
.result-error,
.result-skip {
  font-size: 0.75rem;
  line-height: 0.75rem;
  border-radius: 0.75rem;
  padding: 0.15rem 0.5rem;
}

.result-override {
  background-color: var(--input-green);
}

.result-error {
  background-color: var(--icon-red);
}
.result-rename {
  background-color: var(--icon-orange);
}
.result-skip {
  background-color: var(--icon-blue);
}

.result-buttons > button {
  padding: 0.75rem;
  color: var(--textPrimary);
  margin: 0.25rem 0;
  display: flex;
  justify-content: start;
  align-items: center;
  gap: 0.5rem;
  background: transparent;
  border: solid 1px transparent;
  width: 100%;
  transition: all ease-in-out 200ms;
  cursor: pointer;
  border-radius: 0.25rem;
}

.result-buttons > button:hover {
  border: solid 1px var(--icon-blue);
}
</style>
