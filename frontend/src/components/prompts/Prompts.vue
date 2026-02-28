<template>
  <base-modal v-if="modal != null" :prompt="currentPromptName" @closed="close">
    <keep-alive>
      <component :is="modal" />
    </keep-alive>
  </base-modal>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { storeToRefs } from "pinia";
import { useLayoutStore } from "@/stores/layout";

import BaseModal from "./BaseModal.vue";
import Help from "./Help.vue";
import Info from "./Info.vue";
import Delete from "./Delete.vue";
import DeleteUser from "./DeleteUser.vue";
import Download from "./Download.vue";
import Rename from "./Rename.vue";
import Move from "./Move.vue";
import Copy from "./Copy.vue";
import Unzip from "./Unzip.vue";
import NewFile from "./NewFile.vue";
import NewDir from "./NewDir.vue";
import Replace from "./Replace.vue";
import Share from "./Share.vue";
import ShareDelete from "./ShareDelete.vue";
import Upload from "./Upload.vue";
import DiscardEditorChanges from "./DiscardEditorChanges.vue";
import ResolveConflict from "./ResolveConflict.vue";

const layoutStore = useLayoutStore();

const { currentPromptName } = storeToRefs(layoutStore);

const components = new Map<string, any>([
  ["info", Info],
  ["help", Help],
  ["delete", Delete],
  ["rename", Rename],
  ["move", Move],
  ["copy", Copy],
  ["unzip", Unzip],
  ["newFile", NewFile],
  ["newDir", NewDir],
  ["download", Download],
  ["replace", Replace],
  ["share", Share],
  ["upload", Upload],
  ["share-delete", ShareDelete],
  ["deleteUser", DeleteUser],
  ["discardEditorChanges", DiscardEditorChanges],
  ["resolve-conflict", ResolveConflict],
]);

const modal = computed(() => {
  const modal = components.get(currentPromptName.value!);
  if (!modal) null;

  return modal;
});

const close = () => {
  if (!layoutStore.currentPrompt) return;
  layoutStore.closeHovers();
};
</script>
