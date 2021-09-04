<template>
  <nav :class="{ active }">
    <template v-if="isLogged">
      <router-link
        class="action"
        to="/files/"
        :aria-label="$t('sidebar.myFiles')"
        :title="$t('sidebar.myFiles')"
      >
        <i class="material-icons">folder</i>
        <span>{{ $t("sidebar.myFiles") }}</span>
      </router-link>

        <div v-if="user.perm.create">
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

    </template>
  </nav>
</template>

<script>
import { mapGetters, mapState } from "vuex";
import * as auth from "@/utils/auth";
import { authMethod, disableExternal, noAuth, signup, version } from "@/utils/constants";

export default {
  name: "sidebar",
  computed: {
    ...mapState(["user"]),
    ...mapGetters(["isLogged"]),
    active() {
      return this.$store.state.show === "sidebar";
    },
    signup: () => signup,
    version: () => version,
    disableExternal: () => disableExternal,
    noAuth: () => noAuth,
    authMethod: () => authMethod,
  },
  methods: {
    help() {
      this.$store.commit("showHover", "help");
    },
    logout: auth.logout,
  },
};
</script>
