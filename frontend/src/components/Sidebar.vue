<template>
  <nav :class="{ active }">
    <template v-if="isLogged">
      <button
        class="action"
        @click="toRoot"
        :aria-label="$t('sidebar.myFiles')"
        :title="$t('sidebar.myFiles')"
      >
        <i class="material-icons">folder</i>
        <span>{{ $t("sidebar.myFiles") }}</span>
      </button>

      <div v-if="user.perm.create">
        <button
          @click="$store.commit('showHover', 'newDir')"
          class="action"
          :aria-label="$t('sidebar.newFolder')"
          :title="$t('sidebar.newFolder')"
        >
          <i class="material-icons">create_new_folder</i>
          <span>{{ $t("sidebar.newFolder") }}</span>
        </button>

        <button
          @click="$store.commit('showHover', 'newFile')"
          class="action"
          :aria-label="$t('sidebar.newFile')"
          :title="$t('sidebar.newFile')"
        >
          <i class="material-icons">note_add</i>
          <span>{{ $t("sidebar.newFile") }}</span>
        </button>
      </div>

      <div v-if="tmpDir != '' || trashDir != ''">
        <router-link
          v-if="trashDir != ''"
          class="action"
          :to="'/files/' + trashDir"
          :aria-label="$t('sidebar.trashBin')"
          :title="$t('sidebar.trashBin')"
        >
          <i class="material-icons">delete</i>
          <span>{{ $t("sidebar.trashBin") }}</span>
        </router-link>

        <router-link
          v-if="tmpDir != ''"
          class="action"
          :to="'/files/' + tmpDir"
          :aria-label="$t('sidebar.tmpDir')"
          :title="$t('sidebar.tmpDir')"
        >
          <i class="material-icons">folder_open</i>
          <span>{{ $t("sidebar.tmpDir") }}</span>
        </router-link>
      </div>

      <quota v-if="quotaExists"></quota>

      <div>
        <button
          class="action"
          @click="toSettings"
          :aria-label="$t('sidebar.settings')"
          :title="$t('sidebar.settings')"
        >
          <i class="material-icons">settings_applications</i>
          <span>{{ $t("sidebar.settings") }}</span>
        </button>

        <button
          v-if="
            authMethod == 'json' ||
            (authMethod == 'proxy' && authLogoutURL != '') ||
            canLogout
          "
          @click="logout"
          class="action"
          id="logout"
          :aria-label="$t('sidebar.logout')"
          :title="$t('sidebar.logout')"
        >
          <i class="material-icons">exit_to_app</i>
          <span>{{ $t("sidebar.logout") }}</span>
        </button>
      </div>
    </template>
    <template v-else>
      <router-link
        class="action"
        to="/login"
        :aria-label="$t('sidebar.login')"
        :title="$t('sidebar.login')"
      >
        <i class="material-icons">exit_to_app</i>
        <span>{{ $t("sidebar.login") }}</span>
      </router-link>

      <router-link
        v-if="signup"
        class="action"
        to="/login"
        :aria-label="$t('sidebar.signup')"
        :title="$t('sidebar.signup')"
      >
        <i class="material-icons">person_add</i>
        <span>{{ $t("sidebar.signup") }}</span>
      </router-link>
    </template>

    <div
      class="credits"
      v-if="
        $router.currentRoute.path.includes('/files/') && !disableUsedPercentage
      "
      style="width: 90%; margin: 2em 2.5em 3em 2.5em"
    >
      <progress-bar :val="usage.usedPercentage" size="small"></progress-bar>
      <br />
      {{ usage.used }} of {{ usage.total }} used
    </div>

    <p class="credits">
      <span>
        <span v-if="disableExternal">File Browser</span>
        <a
          v-else
          rel="noopener noreferrer"
          target="_blank"
          href="https://github.com/filebrowser/filebrowser"
          >File Browser</a
        >
        <span> {{ version }}</span>
      </span>
      <span>
        <a @click="help">{{ $t("sidebar.help") }}</a>
      </span>
    </p>
  </nav>
</template>

<script>
import { mapState, mapGetters } from "vuex";
import * as auth from "@/utils/auth";
import Quota from "./Quota.vue";
import {
  version,
  signup,
  disableExternal,
  tmpDir,
  trashDir,
  quotaExists,
  disableUsedPercentage,
  noAuth,
  authMethod,
  authLogoutURL,
  loginPage,
} from "@/utils/constants";
import { files as api } from "@/api";
import ProgressBar from "vue-simple-progress";
import prettyBytes from "pretty-bytes";

export default {
  name: "sidebar",
  components: {
    Quota,
    ProgressBar,
  },
  computed: {
    ...mapState(["user"]),
    ...mapGetters(["isLogged", "currentPrompt"]),
    active() {
      return this.currentPrompt?.prompt === "sidebar";
    },
    signup: () => signup,
    version: () => version,
    disableExternal: () => disableExternal,
    tmpDir: () => tmpDir,
    trashDir: () => trashDir,
    quotaExists: () => quotaExists,
    noAuth: () => noAuth,
    authMethod: () => authMethod,
    authLogoutURL: () => authLogoutURL,
    disableUsedPercentage: () => disableUsedPercentage,
    canLogout: () => !noAuth && loginPage,
  },
  asyncComputed: {
    usage: {
      async get() {
        let path = this.$route.path.endsWith("/")
          ? this.$route.path
          : this.$route.path + "/";
        let usageStats = { used: 0, total: 0, usedPercentage: 0 };
        if (this.disableUsedPercentage) {
          return usageStats;
        }
        try {
          let usage = await api.usage(path);
          usageStats = {
            used: prettyBytes(usage.used, { binary: true }),
            total: prettyBytes(usage.total, { binary: true }),
            usedPercentage: Math.round((usage.used / usage.total) * 100),
          };
        } catch (error) {
          this.$showError(error);
        }
        return usageStats;
      },
      default: { used: "0 B", total: "0 B", usedPercentage: 0 },
      shouldUpdate() {
        return this.$router.currentRoute.path.includes("/files/");
      },
    },
  },
  methods: {
    toRoot() {
      this.$router.push({ path: "/files/" }, () => {});
      this.$store.commit("closeHovers");
    },
    toSettings() {
      this.$router.push({ path: "/settings" }, () => {});
      this.$store.commit("closeHovers");
    },
    help() {
      this.$store.commit("showHover", "help");
    },
    logout: auth.logout,
  },
};
</script>
