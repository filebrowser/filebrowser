<template>
  <div v-show="active" @click="closeHovers" class="overlay"></div>
  <nav :class="{ active }">
    <template v-if="isLoggedIn">
      <button @click="toAccountSettings" class="action">
        <i class="material-icons">person</i>
        <span>{{ user?.username }}</span>
      </button>
      <button
        class="action"
        @click="toRoot"
        :aria-label="$t('sidebar.myFiles')"
        :title="$t('sidebar.myFiles')"
      >
        <i class="material-icons">folder</i>
        <span>{{ $t("sidebar.myFiles") }}</span>
      </button>

      <div v-if="user?.perm.create">
        <button
          @click="showHover('newDir')"
          class="action"
          :aria-label="$t('sidebar.newFolder')"
          :title="$t('sidebar.newFolder')"
        >
          <i class="material-icons">create_new_folder</i>
          <span>{{ $t("sidebar.newFolder") }}</span>
        </button>

        <button
          @click="showHover('newFile')"
          class="action"
          :aria-label="$t('sidebar.newFile')"
          :title="$t('sidebar.newFile')"
        >
          <i class="material-icons">note_add</i>
          <span>{{ $t("sidebar.newFile") }}</span>
        </button>
      </div>

      <div v-if="user?.perm.admin">
        <button
          class="action"
          @click="toGlobalSettings"
          :aria-label="$t('sidebar.settings')"
          :title="$t('sidebar.settings')"
        >
          <i class="material-icons">settings_applications</i>
          <span>{{ $t("sidebar.settings") }}</span>
        </button>
      </div>
      <button
        v-if="canLogout"
        @click="logout"
        class="action"
        id="logout"
        :aria-label="$t('sidebar.logout')"
        :title="$t('sidebar.logout')"
      >
        <i class="material-icons">exit_to_app</i>
        <span>{{ $t("sidebar.logout") }}</span>
      </button>
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
      v-if="isFiles && !disableUsedPercentage"
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
        <span> {{ " " }} {{ version }}</span>
      </span>
      <span>
        <a @click="help">{{ $t("sidebar.help") }}</a>
      </span>
    </p>
  </nav>
</template>

<script setup lang="ts">
import { reactive, ref, computed, watch, onUnmounted } from "vue";
import { storeToRefs } from "pinia";
import { useRoute, useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

import * as auth from "@/utils/auth";
import {
  version,
  signup,
  disableExternal,
  disableUsedPercentage,
  noAuth,
  loginPage,
} from "@/utils/constants";
import { files as api } from "@/api";
import ProgressBar from "@/components/ProgressBar.vue";
import prettyBytes from "pretty-bytes";

const USAGE_DEFAULT = { used: "0 B", total: "0 B", usedPercentage: 0 };

const route = useRoute();
const router = useRouter();

const authStore = useAuthStore();
const fileStore = useFileStore();
const layoutStore = useLayoutStore();

const { user, isLoggedIn } = storeToRefs(authStore);
const { isFiles } = storeToRefs(fileStore);
const { currentPromptName } = storeToRefs(layoutStore);
const { closeHovers, showHover } = layoutStore;

const usage = reactive(USAGE_DEFAULT);
const usageAbortController = ref(new AbortController());

const active = computed(() => {
  return currentPromptName.value === "sidebar";
});

const canLogout = !noAuth && loginPage;

const abortOngoingFetchUsage = () => {
  usageAbortController.value.abort();
};

const fetchUsage = async () => {
  const path = route.path.endsWith("/") ? route.path : route.path + "/";
  let usageStats = USAGE_DEFAULT;
  if (disableUsedPercentage) {
    return Object.assign(usage, usageStats);
  }
  try {
    abortOngoingFetchUsage();
    usageAbortController.value = new AbortController();
    const usageData = await api.usage(path, usageAbortController.value.signal);
    usageStats = {
      used: prettyBytes(usageData.used, { binary: true }),
      total: prettyBytes(usageData.total, { binary: true }),
      usedPercentage: Math.round((usageData.used / usageData.total) * 100),
    };
  } finally {
    return Object.assign(usage, usageStats);
  }
};

const toRoot = () => {
  router.push({ path: "/files" });
  closeHovers();
};

const toAccountSettings = () => {
  router.push({ path: "/settings/profile" });
  closeHovers();
};

const toGlobalSettings = () => {
  router.push({ path: "/settings/global" });
  closeHovers();
};

const help = () => {
  showHover("help");
};

const logout = () => {
  auth.logout();
};

watch(
  () => route.path,
  (newPath) => {
    if (newPath.includes("/files")) {
      fetchUsage();
    }
  },
  { immediate: true }
);

onUnmounted(() => {
  abortOngoingFetchUsage();
});
</script>
