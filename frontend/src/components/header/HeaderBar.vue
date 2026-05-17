<template>
  <header>
    <CncRecoveryBanner />
    <router-link
      v-if="showLogo"
      to="/files/"
      :aria-label="t('sidebar.myFiles')"
      class="logo-link"
    >
      <img :src="logoURL" />
    </router-link>
    <Action
      v-if="showMenu"
      class="menu-button"
      icon="menu"
      :label="t('buttons.toggleSidebar')"
      @action="layoutStore.showHover('sidebar')"
    />

    <slot />

    <CncStatusPill />
    <HostStatsPill />

    <div
      id="dropdown"
      :class="{ active: layoutStore.currentPromptName === 'more' }"
    >
      <slot name="actions" />
    </div>

    <Action
      v-if="ifActionsSlot"
      id="more"
      icon="more_vert"
      :label="t('buttons.more')"
      @action="layoutStore.showHover('more')"
    />

    <div
      class="overlay"
      v-show="layoutStore.currentPromptName == 'more'"
      @click="layoutStore.closeHovers"
    />
  </header>
</template>

<script setup lang="ts">
import { useLayoutStore } from "@/stores/layout";

import { logoURL } from "@/utils/constants";

import Action from "@/components/header/Action.vue";
import CncRecoveryBanner from "@/components/header/CncRecoveryBanner.vue";
import CncStatusPill from "@/components/header/CncStatusPill.vue";
import HostStatsPill from "@/components/header/HostStatsPill.vue";
import { computed, onMounted, useSlots } from "vue";
import { useI18n } from "vue-i18n";
import { useCncStore } from "@/stores/cnc";

defineProps<{
  showLogo?: boolean;
  showMenu?: boolean;
}>();

const layoutStore = useLayoutStore();
const slots = useSlots();

const { t } = useI18n();

const ifActionsSlot = computed(() => (slots.actions ? true : false));

// HeaderBar mounts on every authenticated layout, so this is the
// natural spot to kick the live CNC status feed. The store's start()
// is idempotent — extra mounts are no-ops.
const cncStore = useCncStore();
onMounted(() => {
  cncStore.start();
});
</script>

<style></style>
