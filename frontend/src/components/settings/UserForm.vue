<template>
  <div>
    <p v-if="!isDefault && props.user !== null">
      <label for="username">{{ t("settings.username") }}</label>
      <input
        class="input input--block"
        type="text"
        v-model="user.username"
        id="username"
      />
    </p>

    <p v-if="!isDefault">
      <label for="password">{{ t("settings.password") }}</label>
      <input
        class="input input--block"
        type="password"
        :placeholder="passwordPlaceholder"
        v-model="user.password"
        id="password"
      />
    </p>

    <p>
      <label for="scope">{{ t("settings.scope") }}</label>
      <div
        v-for="(_, index) in scopesList"
        :key="index"
        class="scope-row"
      >
        <input
          :placeholder="index === 0 ? scopePlaceholder : ''"
          class="input input--block scope-input"
          type="text"
          :value="scopesList[index]"
          @input="updateScope(index, ($event.target as HTMLInputElement).value)"
        />
        <button
          v-if="scopesList.length > 1"
          class="button button--flat button--red scope-remove"
          type="button"
          @click="removeScope(index)"
          :title="t('buttons.delete')"
        >
          &times;
        </button>
      </div>
      <button
        v-if="!isDefault"
        class="button button--flat"
        type="button"
        @click="addScope"
      >
        + {{ t("settings.addScope") }}
      </button>
    </p>

    <p>
      <label for="locale">{{ t("settings.language") }}</label>
      <languages
        class="input input--block"
        id="locale"
        v-model:locale="user.locale"
      ></languages>
    </p>

    <p v-if="!isDefault && user.perm">
      <input
        type="checkbox"
        :disabled="user.perm.admin"
        v-model="user.lockPassword"
      />
      {{ t("settings.lockPassword") }}
    </p>

    <permissions v-model:perm="user.perm" />
    <commands v-if="enableExec" v-model:commands="user.commands" />

    <div v-if="!isDefault">
      <h3>{{ t("settings.rules") }}</h3>
      <p class="small">{{ t("settings.rulesHelp") }}</p>
      <rules v-model:rules="user.rules" />
    </div>
  </div>
</template>

<script setup lang="ts">
import Languages from "./Languages.vue";
import Rules from "./Rules.vue";
import Permissions from "./Permissions.vue";
import Commands from "./Commands.vue";
import { enableExec } from "@/utils/constants";
import { computed, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

const scopesList = ref<string[]>([""]);

const props = defineProps<{
  user: IUserForm;
  isNew: boolean;
  isDefault: boolean;
  createUserDir?: boolean;
  userHomeBasePath?: string;
}>();

onMounted(() => {
  if (props.user.scopes && props.user.scopes.length > 0) {
    scopesList.value = [...props.user.scopes];
  } else {
    scopesList.value = [props.user.scope || ""];
  }
});

function updateScope(index: number, value: string) {
  scopesList.value[index] = value;
  syncScopesToUser();
}

function addScope() {
  scopesList.value.push("");
  syncScopesToUser();
}

function removeScope(index: number) {
  scopesList.value.splice(index, 1);
  syncScopesToUser();
}

function syncScopesToUser() {
  const filtered = scopesList.value.filter((s) => s.trim() !== "");
  if (filtered.length > 1) {
    props.user.scopes = filtered;
    props.user.scope = filtered[0];
  } else {
    props.user.scopes = undefined;
    props.user.scope = filtered[0] || "";
  }
}

const passwordPlaceholder = computed(() =>
  props.isNew ? "" : t("settings.avoidChanges")
);
const scopePlaceholder = computed(() =>
  props.createUserDir ? t("settings.userScopeGenerationPlaceholder") : ""
);

const userHomePath = computed(() => {
  if (!props.createUserDir || !props.user.username) return "";
  const base = (props.userHomeBasePath || "/users").replace(/\/+$/, "");
  return `${base}/${props.user.username}`;
});

watch(
  () => props.user,
  () => {
    if (!props.user?.perm?.admin) return;
    props.user.lockPassword = false;
  }
);

watch(userHomePath, (path) => {
  if (props.isNew && props.createUserDir && path) {
    scopesList.value[0] = path;
    syncScopesToUser();
  }
});
</script>

<style scoped>
.scope-row {
  display: flex;
  align-items: center;
  gap: 0.5em;
  margin-bottom: 0.5em;
}

.scope-input {
  flex: 1;
}

.scope-remove {
  padding: 0.25em 0.5em;
  font-size: 1.2em;
  line-height: 1;
}
</style>
