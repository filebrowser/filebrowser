<template>
  <header>
    <img v-if="showLogo !== undefined" :src="logoURL" />
    <action
      v-if="showMenu !== undefined"
      class="menu-button"
      icon="menu"
      :label="$t('buttons.toggleSidebar')"
      @action="showHover('sidebar')"
    />

    <slot />

    <div id="dropdown" :class="{ active: this.show === 'more' }">
      <slot name="actions" />
    </div>

    <action
      v-if="this.$slots.actions"
      id="more"
      icon="more_vert"
      :label="$t('buttons.more')"
      @action="showHover('more')"
    />

    <div class="overlay" v-show="this.show == 'more'" @click="closeHovers" />
  </header>
</template>

<script>
import { mapActions, mapState } from "pinia";
import { useLayoutStore } from "@/stores/layout";

import { logoURL } from "@/utils/constants";

import Action from "@/components/header/Action.vue";

export default {
  name: "header-bar",
  props: ["showLogo", "showMenu"],
  components: {
    Action,
  },
  data: function () {
    return {
      logoURL,
    };
  },
  computed: {
    ...mapState(useLayoutStore, ["show"]),
  },
  methods: {
    ...mapActions(useLayoutStore, ["showHover", "closeHovers"]),
  },
};
</script>

<style></style>
@/stores/layout
