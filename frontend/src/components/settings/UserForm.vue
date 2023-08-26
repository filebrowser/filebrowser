<template>
  <div>
    <p v-if="!isDefault">
      <label for="username">{{ $t("settings.username") }}</label>
      <input
        class="input input--block"
        type="text"
        v-model="user.username"
        id="username"
      />
    </p>

    <p v-if="!isDefault">
      <label for="password">{{ $t("settings.password") }}</label>
      <input
        class="input input--block"
        type="password"
        :placeholder="passwordPlaceholder"
        v-model="user.password"
        id="password"
      />
    </p>

    <p>
      <label for="scope">{{ $t("settings.scope") }}</label>
      <input
        :disabled="createUserDirData"
        :placeholder="scopePlaceholder"
        class="input input--block"
        type="text"
        v-model="user.scope"
        id="scope"
      />
    </p>
    <p class="small" v-if="displayHomeDirectoryCheckbox">
      <input type="checkbox" v-model="createUserDirData" />
      {{ $t("settings.createUserHomeDirectory") }}
    </p>

    <p>
      <label for="locale">{{ $t("settings.language") }}</label>
      <languages
        class="input input--block"
        id="locale"
        :locale.sync="user.locale"
      ></languages>
    </p>

    <p v-if="!isDefault">
      <input
        type="checkbox"
        :disabled="user.perm.admin"
        v-model="user.lockPassword"
      />
      {{ $t("settings.lockPassword") }}
    </p>

    <permissions :perm.sync="user.perm" />
    <commands v-if="isExecEnabled" :commands.sync="user.commands" />

    <div v-if="!isDefault">
      <h3>{{ $t("settings.rules") }}</h3>
      <p class="small">{{ $t("settings.rulesHelp") }}</p>
      <rules :rules.sync="user.rules" />
    </div>
  </div>
</template>

<script>
import Languages from "./Languages.vue";
import Rules from "./Rules.vue";
import Permissions from "./Permissions.vue";
import Commands from "./Commands.vue";
import { enableExec } from "@/utils/constants";

export default {
  name: "user",
  data: () => {
    return {
      createUserDirData: false,
      originalUserScope: "/",
    };
  },
  components: {
    Permissions,
    Languages,
    Rules,
    Commands,
  },
  props: ["user", "createUserDir", "isNew", "isDefault"],
  created() {
    this.originalUserScope = this.user.scope;
    this.createUserDirData = this.createUserDir;
  },
  computed: {
    passwordPlaceholder() {
      return this.isNew ? "" : this.$t("settings.avoidChanges");
    },
    scopePlaceholder() {
      return this.createUserDir
        ? this.$t("settings.userScopeGenerationPlaceholder")
        : "";
    },
    displayHomeDirectoryCheckbox() {
      return this.isNew && this.createUserDir;
    },
    isExecEnabled: () => enableExec,
  },
  watch: {
    "user.perm.admin": function () {
      if (!this.user.perm.admin) return;
      this.user.lockPassword = false;
    },
    createUserDirData() {
      this.user.scope = this.createUserDirData ? "" : this.originalUserScope;
    },
  },
};
</script>
