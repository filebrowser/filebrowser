<template>
  <div>
    <header-bar v-if="error || req.type == null" showMenu showLogo />

    <breadcrumbs base="/files" />

    <errors v-if="error" :errorCode="errorCode" />
    <component v-else-if="currentView" :is="currentView"></component>
    <div v-else>
      <h2 class="message">
        <span>{{ $t("files.loading") }}</span>
      </h2>
    </div>
  </div>
</template>

<script>
import { files as api } from "@/api";
import { mapState, mapMutations } from "vuex";

import HeaderBar from "@/components/header/HeaderBar";
import Breadcrumbs from "@/components/Breadcrumbs";
import Errors from "@/views/Errors";
import Preview from "@/views/files/Preview";
import Listing from "@/views/files/Listing";

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
    Editor: () => import("@/views/files/Editor"),
  },
  data: function () {
    return {
      error: null,
      width: window.innerWidth,
    };
  },
  computed: {
    ...mapState(["req", "reload", "loading", "show"]),
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
    errorCode() {
      return this.error.message === "404" || this.error.message === "403"
        ? parseInt(this.error.message)
        : 500;
    },
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
        document.title = res.name;
      } catch (e) {
        this.error = e;
      } finally {
        this.setLoading(false);
      }
    },
    keyEvent(event) {
      if (this.show !== null) {
        // Esc!
        if (event.keyCode === 27) {
          this.$store.commit("closeHovers");
        }

        return;
      }

      // F1!
      if (event.keyCode === 112) {
        event.preventDefault();
        this.$store.commit("showHover", "help");
      }
    },
  },
};
</script>
