<template>
  <div>
    <header-bar v-if="error || req.type == null" showMenu showLogo />

    <breadcrumbs base="/files" />

    <errors v-if="error" :errorCode="error.status" />
    <component
      v-else-if="currentView"
      :is="currentView"
      @changed="setChanged"
    ></component>
    <div v-else>
      <h2 class="message delayed">
        <div class="spinner">
          <div class="bounce1"></div>
          <div class="bounce2"></div>
          <div class="bounce3"></div>
        </div>
        <span>{{ $t("files.loading") }}</span>
      </h2>
    </div>
  </div>
</template>

<script>
import { files as api } from "@/api";
import { mapState, mapMutations } from "vuex";

import HeaderBar from "@/components/header/HeaderBar.vue";
import Breadcrumbs from "@/components/Breadcrumbs.vue";
import Errors from "@/views/Errors.vue";
import Preview from "@/views/files/Preview.vue";
import Listing from "@/views/files/Listing.vue";

function clean(path) {
  return path.endsWith("/") ? path.slice(0, -1) : path;
}

export default {
  name: "files",
  components: {
    HeaderBar,
    Breadcrumbs,
    Errors,
    Preview,
    Listing,
    Editor: () => import("@/views/files/Editor.vue"),
  },
  data: function () {
    return {
      error: null,
      width: window.innerWidth,
      unsavedChanges: false,
    };
  },
  computed: {
    ...mapState(["req", "reload", "loading"]),
    currentView() {
      if (this.req.type == undefined) {
        return null;
      }

      if (this.req.isDir) {
        return "listing";
      } else if (
        this.req.type === "text" ||
        this.req.type === "textImmutable"
      ) {
        return "editor";
      } else {
        return "preview";
      }
    },
  },
  beforeRouteUpdate(to, from, next) {
    this.verifyRouteChange(next);
  },
  beforeRouteLeave(to, from, next) {
    this.verifyRouteChange(next);
  },
  created() {
    this.fetchData();
  },
  watch: {
    $route: "fetchData",
    reload: function (value) {
      if (value === true) {
        this.fetchData();
      }
    },
  },
  mounted() {
    window.addEventListener("keydown", this.keyEvent);
  },
  beforeDestroy() {
    window.removeEventListener("keydown", this.keyEvent);
  },
  destroyed() {
    if (this.$store.state.showShell) {
      this.$store.commit("toggleShell");
    }
    this.$store.commit("updateRequest", {});
  },
  methods: {
    verifyRouteChange(next) {
      if (
        this.currentView === "editor" &&
        this.unsavedChanges &&
        !confirm(this.$t("prompts.unsavedChanges"))
      ) {
        next(false);
      } else {
        this.unsavedChanges = false;
        next();
      }
    },
    setChanged(value) {
      this.unsavedChanges = value;
    },
    ...mapMutations(["setLoading"]),
    async fetchData() {
      // Reset view information.
      this.$store.commit("setReload", false);
      this.$store.commit("resetSelected");
      this.$store.commit("multiple", false);
      this.$store.commit("closeHovers");

      // Set loading to true and reset the error.
      this.setLoading(true);
      this.error = null;

      let url = this.$route.path;
      if (url === "") url = "/";
      if (url[0] !== "/") url = "/" + url;

      try {
        const res = await api.fetch(url);

        if (clean(res.path) !== clean(`/${this.$route.params.pathMatch}`)) {
          return;
        }

        this.$store.commit("updateRequest", res);
        document.title = `${res.name} - ${document.title}`;
      } catch (e) {
        this.error = e;
      } finally {
        this.setLoading(false);
      }
    },
    keyEvent(event) {
      // F1!
      if (event.keyCode === 112) {
        event.preventDefault();
        this.$store.commit("showHover", "help");
      }
    },
  },
};
</script>
