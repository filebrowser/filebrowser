<template>
  <div>
    <header-bar v-if="error || req.type == null" showMenu showLogo />

    <breadcrumbs base="/files" />

    <errors v-if="error" :errorCode="error.status" />
    <component v-else-if="currentView" :is="currentView"></component>
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
import { mapState, mapActions, mapWritableState } from "pinia";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

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
    };
  },
  computed: {
    ...mapWritableState(useFileStore, [
      "req",
      "reload",
      "selected",
      "multiple",
    ]),
    ...mapState(useLayoutStore, ["show", "showShell"]),
    ...mapWritableState(useLayoutStore, ["loading"]),
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
  beforeUnmount() {
    window.removeEventListener("keydown", this.keyEvent);
  },
  unmounted() {
    if (this.showShell) {
      this.toggleShell();
    }
    this.updateRequest({});
  },
  methods: {
    ...mapActions(useLayoutStore, ["toggleShell", "showHover", "closeHovers"]),
    ...mapActions(useFileStore, ["updateRequest"]),
    async fetchData() {
      // Reset view information.
      this.reload = false;
      this.selected = [];
      this.multiple = false;
      this.closeHovers();

      // Set loading to true and reset the error.
      this.loading = true;
      this.error = null;

      let url = this.$route.path;
      if (url === "") url = "/";
      if (url[0] !== "/") url = "/" + url;

      try {
        const res = await api.fetch(url);

        if (
          clean(res.path) !==
          clean(`/${this.$route.params.path}`).replace(/,/g, "/")
        ) {
          throw new Error("Data Mismatch!");
        }

        this.updateRequest(res);
        document.title = `${res.name} - ${document.title}`;
      } catch (e) {
        this.error = e;
      } finally {
        this.loading = false;
      }
    },
    keyEvent(event) {
      // F1!
      if (event.keyCode === 112) {
        event.preventDefault();
        this.showHover("help");
      }
    },
  },
};
</script>
@/stores/file@/stores/layout
