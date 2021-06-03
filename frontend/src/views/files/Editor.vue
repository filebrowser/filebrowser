<template>
  <div id="editor-container">
    <header-bar>
      <action icon="close" :label="$t('buttons.close')" @action="close()" />
      <title>{{ req.name }}</title>

      <template #actions>
        <action
          v-if="user.perm.modify"
          id="save-button"
          icon="save"
          :label="$t('buttons.save')"
          @action="save()"
        />
      </template>
    </header-bar>
    <form id="editor"></form>
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

export default {
  name: "editor",
  components: {
    HeaderBar,
    Action,
  },
  data: function () {
    return {};
  },
  computed: {
    ...mapState(["req", "user"]),
  },
  created() {
    window.addEventListener("keydown", this.keyEvent);
  },
  beforeDestroy() {
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
