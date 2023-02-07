<template>
  <div id="editor-container">
    <header-bar>
      <action icon="close" :label="$t('buttons.close')" @action="close()" />
      <title>{{ req.name }}</title>

      <action
        v-if="headerButtons.upload"
        icon="file_upload"
        id="upload-button"
        :label="$t('buttons.upload')"
        @action="upload"
      />
      <a
        v-if="previewLink"
        aria-label="preview"
        title="Preview"
        class="action"
        :href="previewLink"
        id="preview"
        target="preview"
        >
        <i class="material-icons">preview</i>
        <span>preview</span>
      </a>
      <action
        v-if="user.perm.modify"
        id="save-button"
        icon="save"
        :label="$t('buttons.save')"
        @action="save()"
      />
    </header-bar>

    <breadcrumbs base="/files" noLink />

    <form id="editor"></form>
    <input
      style="display: none"
      type="file"
      id="upload-input"
      @change="uploadInput($event)"
      multiple
    />
    <input
      style="display: none"
      type="file"
      id="upload-folder-input"
      @change="uploadInput($event)"
      webkitdirectory
      multiple
    />
  </div>
</template>

<script>
import { mapState } from "vuex";
import { files as api } from "@/api";
import { theme } from "@/utils/constants";
import buttons from "@/utils/buttons";
import url from "@/utils/url";

import ace from "ace-builds/src-min-noconflict/ace.js";
import modelist from "ace-builds/src-min-noconflict/ext-modelist.js";
import "ace-builds/webpack-resolver";

import HeaderBar from "@/components/header/HeaderBar";
import Action from "@/components/header/Action";
import Breadcrumbs from "@/components/Breadcrumbs";
import * as upload from "@/utils/upload";

export default {
  name: "editor",
  components: {
    HeaderBar,
    Action,
    Breadcrumbs,
  },
  data: function () {
    return {};
  },
  computed: {
    ...mapState(["req", "user"]),
    previewLink() {
      return this.$route.path
      .replace(/^\/files\//, "/")
      .replace(/[^/]*$/, "");
    },
    headerButtons() {
      return {
        upload: this.user.perm.create,
      };
    },
    breadcrumbs() {
      let parts = this.$route.path.split("/");

      if (parts[0] === "") {
        parts.shift();
      }

      if (parts[parts.length - 1] === "") {
        parts.pop();
      }

      let breadcrumbs = [];

      for (let i = 0; i < parts.length; i++) {
        breadcrumbs.push({ name: decodeURIComponent(parts[i]) });
      }

      breadcrumbs.shift();

      if (breadcrumbs.length > 3) {
        while (breadcrumbs.length !== 4) {
          breadcrumbs.shift();
        }

        breadcrumbs[0].name = "...";
      }

      return breadcrumbs;
    },
  },
  created() {
    window.addEventListener("keydown", this.keyEvent);
    window.addEventListener("drop", this.drop);
    window.addEventListener("paste", this.drop);
  },
  beforeDestroy() {
    window.removeEventListener("paste", this.drop);
    window.removeEventListener("drop", this.drop);
    window.removeEventListener("keydown", this.keyEvent);
    this.editor.destroy();
  },
  mounted: function () {
    const fileContent = this.req.content || "";

    this.editor = ace.edit("editor", {
      value: fileContent,
      showPrintMargin: false,
      readOnly: this.req.type === "textImmutable",
      theme: "ace/theme/chrome",
      mode: modelist.getModeForPath(this.req.name).mode,
      wrap: true,
    });

    if (theme == "dark") {
      this.editor.setTheme("ace/theme/twilight");
    }
  },
  methods: {
    back() {
      let uri = url.removeLastDir(this.$route.path) + "/";
      this.$router.push({ path: uri });
    },
    dragEnter() {
      this.dragCounter++;

      // When the user starts dragging an item, put every
      // file on the listing with 50% opacity.
      let items = document.getElementsByClassName("item");

      Array.from(items).forEach((file) => {
        file.style.opacity = 0.5;
      });
    },
    dragLeave() {
      this.dragCounter--;

      if (this.dragCounter == 0) {
        this.resetOpacity();
      }
    },
    drop: async function (event) {
      event.preventDefault();
      this.dragCounter = 0;
      this.resetOpacity();

      let dt = event.dataTransfer ?? event.clipboardData;
      let el = event.target;

      if (dt.files.length <= 0) return;

      for (let i = 0; i < 5; i++) {
        if (el !== null && !el.classList.contains("item")) {
          el = el.parentElement;
        }
      }

      let files = await upload.scanFiles(dt);
      let items = this.req.items;
      let path = this.$route.path.replace(/\/[^/]*$/, "/");

      if (
        el !== null &&
        el.classList.contains("item") &&
        el.dataset.dir === "true"
      ) {
        // Get url from ListingItem instance
        path = el.__vue__.url;

        try {
          items = (await api.fetch(path)).items;
        } catch (error) {
          this.$showError(error);
        }
      }

      let conflict = upload.checkConflict(files, items);

      if (conflict) {
        this.$store.commit("showHover", {
          prompt: "replace",
          confirm: (event) => {
            event.preventDefault();
            this.$store.commit("closeHovers");
            upload.handleFiles(files, path, true);
          },
        });

        return;
      }

      upload.handleFiles(files, path);
      const alt = files[0].name;
      const link = files[0].name;
      this.editor.insert(`![${alt}](${link})`);
    },
    uploadInput(event) {
      this.$store.commit("closeHovers");

      let files = event.currentTarget.files;
      let folder_upload =
        files[0].webkitRelativePath !== undefined &&
        files[0].webkitRelativePath !== "";

      if (folder_upload) {
        for (let i = 0; i < files.length; i++) {
          let file = files[i];
          files[i].fullPath = file.webkitRelativePath;
        }
      }

      let path = this.$route.path.replace(/\/[^/]*$/, "/");
      let conflict = upload.checkConflict(files, this.req.items);

      if (conflict) {
        this.$store.commit("showHover", {
          prompt: "replace",
          confirm: (event) => {
            event.preventDefault();
            this.$store.commit("closeHovers");
            upload.handleFiles(files, path, true);
          },
        });

        return;
      }

      upload.handleFiles(files, path);
    },
    resetOpacity() {
      let items = document.getElementsByClassName("item");

      Array.from(items).forEach((file) => {
        file.style.opacity = 1;
      });
    },
    upload: function () {
      if (
        typeof window.DataTransferItem !== "undefined" &&
        typeof DataTransferItem.prototype.webkitGetAsEntry !== "undefined"
      ) {
        this.$store.commit("showHover", "upload");
      } else {
        document.getElementById("upload-input").click();
      }
    },
    keyEvent(event) {
      if (!event.ctrlKey && !event.metaKey) {
        return;
      }

      if (String.fromCharCode(event.which).toLowerCase() !== "s") {
        return;
      }

      event.preventDefault();
      this.save();
    },
    async save() {
      const button = "save";
      buttons.loading("save");

      try {
        await api.put(this.$route.path, this.editor.getValue());
        buttons.success(button);
      } catch (e) {
        buttons.done(button);
        this.$showError(e);
      }
    },
    close() {
      this.$store.commit("updateRequest", {});

      let uri = url.removeLastDir(this.$route.path) + "/";
      this.$router.push({ path: uri });
    },
  },
};
</script>
