<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ t("prompts.permissions") }}</h2>
    </div>

    <div class="card-content" id="permissions">
      <table>
        <thead>
          <tr>
            <td></td>
            <td>{{ t("prompts.read") }}</td>
            <td>{{ t("prompts.write") }}</td>
            <td>{{ t("prompts.execute") }}</td>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>{{ t("prompts.owner") }}</td>
            <td>
              <input type="checkbox" v-model="permissions.owner.read" />
            </td>
            <td>
              <input type="checkbox" v-model="permissions.owner.write" />
            </td>
            <td>
              <input type="checkbox" v-model="permissions.owner.execute" />
            </td>
          </tr>
          <tr>
            <td>{{ t("prompts.group") }}</td>
            <td>
              <input type="checkbox" v-model="permissions.group.read" />
            </td>
            <td>
              <input type="checkbox" v-model="permissions.group.write" />
            </td>
            <td>
              <input type="checkbox" v-model="permissions.group.execute" />
            </td>
          </tr>
          <tr>
            <td>{{ t("prompts.others") }}</td>
            <td>
              <input type="checkbox" v-model="permissions.others.read" />
            </td>
            <td>
              <input type="checkbox" v-model="permissions.others.write" />
            </td>
            <td>
              <input type="checkbox" v-model="permissions.others.execute" />
            </td>
          </tr>
        </tbody>
      </table>
      <p>
        <code>{{ permModeString }} ({{ permMode.toString(8) }})</code>
      </p>
      <template v-if="dirSelected">
        <p>
          <input type="checkbox" v-model="recursive" />
          {{ t("prompts.recursive") }}:
        </p>
        <div class="recursion-types">
          <p>
            <input
              type="radio"
              id="recursive-all"
              value="all"
              :disabled="!recursive"
              v-model="recursionType"
            />
            <label for="recursive-all">
              {{ t("prompts.directoriesAndFiles") }}
            </label>
          </p>
          <p>
            <input
              type="radio"
              id="recursive-directories"
              value="directories"
              :disabled="!recursive"
              v-model="recursionType"
            />
            <label for="recursive-directories">
              {{ t("prompts.directories") }}
            </label>
          </p>
          <p>
            <input
              type="radio"
              id="recursive-files"
              value="files"
              :disabled="!recursive"
              v-model="recursionType"
            />
            <label for="recursive-files">
              {{ t("prompts.files") }}
            </label>
          </p>
        </div>
      </template>
    </div>

    <div class="card-action">
      <button
        class="button button--flat"
        @click="chmod"
        :disabled="loading"
        :aria-label="$t('buttons.update')"
        :title="$t('buttons.update')"
      >
        {{ t("buttons.update") }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, inject, onMounted, ref } from "vue";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";
import { useI18n } from "vue-i18n";
import { files as api } from "@/api";

const fileStore = useFileStore();
const layoutStore = useLayoutStore();

const { t } = useI18n();

const $showError = inject<IToastError>("$showError")!;

const loading = ref<boolean>(false);
const recursive = ref<boolean>(false);
const recursionType = ref<string>("all");
const permissions = ref<FilePermissions>({
  owner: {
    read: false,
    write: false,
    execute: false,
  },
  group: {
    read: false,
    write: false,
    execute: false,
  },
  others: {
    read: false,
    write: false,
    execute: false,
  },
});

const masks = {
  permissions: 511,
  owner: {
    read: 256,
    write: 128,
    execute: 64,
  },
  group: {
    read: 32,
    write: 16,
    execute: 8,
  },
  others: {
    read: 4,
    write: 2,
    execute: 1,
  },
};

onMounted(() => {
  let item = fileStore.req?.items[fileStore.selected[0]];
  if (!item) {
    return;
  }

  let perms = item.mode & masks.permissions;

  // OWNER PERMS
  permissions.value.owner.read = (perms & masks.owner.read) != 0;
  permissions.value.owner.write = (perms & masks.owner.write) != 0;
  permissions.value.owner.execute = (perms & masks.owner.execute) != 0;
  // GROUP PERMS
  permissions.value.group.read = (perms & masks.group.read) != 0;
  permissions.value.group.write = (perms & masks.group.write) != 0;
  permissions.value.group.execute = (perms & masks.group.execute) != 0;
  // OTHERS PERMS
  permissions.value.others.read = (perms & masks.others.read) != 0;
  permissions.value.others.write = (perms & masks.others.write) != 0;
  permissions.value.others.execute = (perms & masks.others.execute) != 0;
});

const permMode = computed((): number => {
  let mode = 0;
  mode |= masks.owner.read * (permissions.value.owner.read ? 1 : 0);
  mode |= masks.owner.write * (permissions.value.owner.write ? 1 : 0);
  mode |= masks.owner.execute * (permissions.value.owner.execute ? 1 : 0);
  mode |= masks.group.read * (permissions.value.group.read ? 1 : 0);
  mode |= masks.group.write * (permissions.value.group.write ? 1 : 0);
  mode |= masks.group.execute * (permissions.value.group.execute ? 1 : 0);
  mode |= masks.others.read * (permissions.value.others.read ? 1 : 0);
  mode |= masks.others.write * (permissions.value.others.write ? 1 : 0);
  mode |= masks.others.execute * (permissions.value.others.execute ? 1 : 0);
  return mode;
});

const permModeString = computed((): string => {
  let perms = permMode;
  let s = "";
  s += (perms.value & masks.owner.read) != 0 ? "r" : "-";
  s += (perms.value & masks.owner.write) != 0 ? "w" : "-";
  s += (perms.value & masks.owner.execute) != 0 ? "x" : "-";
  s += (perms.value & masks.group.read) != 0 ? "r" : "-";
  s += (perms.value & masks.group.write) != 0 ? "w" : "-";
  s += (perms.value & masks.group.execute) != 0 ? "x" : "-";
  s += (perms.value & masks.others.read) != 0 ? "r" : "-";
  s += (perms.value & masks.others.write) != 0 ? "w" : "-";
  s += (perms.value & masks.others.execute) != 0 ? "x" : "-";
  return s;
});

const dirSelected = computed((): boolean => {
  let item = fileStore.req?.items[fileStore.selected[0]];
  if (!item) {
    return false;
  }

  return item.isDir;
});

const chmod = async () => {
  let item = fileStore.req?.items[fileStore.selected[0]];
  if (!item) {
    return;
  }

  try {
    loading.value = true;

    await api.chmod(
      item.url,
      permMode.value,
      recursive.value,
      recursionType.value
    );

    layoutStore.closeHovers();
    fileStore.reload = true;
  } catch (e: any) {
    $showError(e);
  } finally {
    loading.value = false;
  }
};
</script>
