<template>
  <div id="editor-container">
    <header-bar>
      <action icon="close" :label="$t('buttons.close')" @action="close()" />
      <title>{{ req.name }}</title>
    </header-bar>
    <breadcrumbs base="/files" noLink />
    <errors v-if="error" :errorCode="error.status" />
    <div id="editor">
      <div id="onlyoffice-editor"></div>
    </div>
  </div>
</template>

<script>
import { mapState } from "vuex";
import url from "@/utils/url";
import { onlyOfficeUrl } from "@/utils/constants";

import HeaderBar from "@/components/header/HeaderBar.vue";
import Action from "@/components/header/Action.vue";
import Breadcrumbs from "@/components/Breadcrumbs.vue";
import Errors from "@/views/Errors.vue";
import { fetchJSON } from "@/api/utils";

export default {
  name: "onlyofficeeditor",
  components: {
    HeaderBar,
    Action,
    Breadcrumbs,
    Errors,
  },
  data: function () {
    return {
      error: null,
      clientConfig: null,
    };
  },
  computed: {
    ...mapState(["req", "user", "jwt"]),
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
    const isMobile = window.innerWidth <= 736;
    this.clientConfigPromise = fetchJSON(
      `/api/onlyoffice/client-config${this.req.path}?isMobile=${isMobile}`
    );
    window.addEventListener("keydown", this.keyEvent);
  },
  beforeDestroy() {
    window.removeEventListener("keydown", this.keyEvent);
    this.editor.destroyEditor();
  },
  mounted: function () {
    const scriptUrl = `${onlyOfficeUrl}/web-apps/apps/api/documents/api.js`;
    const onlyofficeScript = document.createElement("script");
    onlyofficeScript.setAttribute("src", scriptUrl);
    document.head.appendChild(onlyofficeScript);

    onlyofficeScript.onload = async () => {
      try {
        const clientConfig = await this.clientConfigPromise;
        // eslint-disable-next-line no-undef
        this.editor = new DocsAPI.DocEditor("onlyoffice-editor", clientConfig);
      } catch (e) {
        this.error = e;
      }
    };
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
    close() {
      this.$store.commit("updateRequest", {});

      let uri = url.removeLastDir(this.$route.path) + "/";
      this.$router.push({ path: uri });
    },
  },
};
</script>
