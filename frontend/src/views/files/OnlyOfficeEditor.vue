<template>
  <div id="editor-container">
    <header-bar>
      <action icon="close" :label="$t('buttons.close')" @action="close()" />
      <title>{{ fileStore.req?.name ?? "" }}</title>
    </header-bar>
    <breadcrumbs base="/files" noLink />
    <errors v-if="error" :errorCode="error.status" />
    <div id="editor" v-if="clientConfig">
      <DocumentEditor
        v-if="clientConfig"
        id="onlyoffice-editor"
        :documentServerUrl="onlyOfficeUrl"
        :config="clientConfig"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import url from "@/utils/url";
import { onlyOfficeUrl } from "@/utils/constants";
import HeaderBar from "@/components/header/HeaderBar.vue";
import Action from "@/components/header/Action.vue";
import Breadcrumbs from "@/components/Breadcrumbs.vue";
import Errors from "@/views/Errors.vue";
import { fetchJSON, StatusError } from "@/api/utils";
import { useFileStore } from "@/stores/file";
import { useRoute } from "vue-router";
import { useRouter } from "vue-router";
import { onMounted, ref } from "vue";
import { DocumentEditor } from "@onlyoffice/document-editor-vue";

const fileStore = useFileStore();
const route = useRoute();
const router = useRouter();
const error = ref<StatusError | null>(null);
const clientConfig = ref<any>(null);

onMounted(async () => {
  try {
    const isMobile = window.innerWidth <= 736;
    clientConfig.value = await fetchJSON(
      `/api/onlyoffice/client-config${fileStore.req!.path}?isMobile=${isMobile}`
    );
  } catch (err) {
    if (err instanceof Error) {
      error.value = err;
    }
  }
});

const close = () => {
  fileStore.updateRequest(null);
  let uri = url.removeLastDir(route.path) + "/";
  router.push({ path: uri });
};
</script>
