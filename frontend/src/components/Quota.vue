<template>
  <div id="quota">
    <div>
      <div class="quota-label">
        <div>{{ $t("sidebar.quota.space") }}</div>
        <div v-if="loaded" class="quota-metric">{{ spaceUsageTitle }}</div>
      </div>
      <div class="quota-bar" :title="spaceProgress + '%'">
        <div
          class="quota-progress"
          :style="{ width: spaceProgress + '%' }"
        ></div>
      </div>
    </div>
    <div>
      <div class="quota-label">
        <div>{{ $t("sidebar.quota.inodes") }}</div>
        <div v-if="loaded" class="quota-metric">{{ inodeUsageTitle }}</div>
      </div>
      <div class="quota-bar" :title="inodeProgress + '%'">
        <div
          class="quota-progress"
          :style="{ width: inodeProgress + '%' }"
        ></div>
      </div>
    </div>
  </div>
</template>

<script>
import filesize from "filesize";
import { mapState } from "vuex";

export default {
  name: "quota",
  computed: {
    ...mapState("quota", {
      inodes: (state) => state.inodes,
      space: (state) => state.space,
    }),
    loaded() {
      return this.inodes !== null && this.space !== null;
    },
    spaceProgress() {
      if (!this.loaded) {
        return 0;
      }

      return this.progress(this.space);
    },
    inodeProgress() {
      if (!this.loaded) {
        return 0;
      }

      return this.progress(this.inodes);
    },
    spaceUsageTitle() {
      if (this.space === null) {
        return "- / -";
      } else {
        return filesize(this.space.usage) + " / " + filesize(this.space.quota);
      }
    },
    inodeUsageTitle() {
      if (this.inodes === null) {
        return "- / -";
      } else {
        return this.inodes.usage + " / " + this.inodes.quota;
      }
    },
  },
  mounted() {
    this.$store.dispatch("quota/fetch");
  },
  methods: {
    progress(metric) {
      let prc = (metric.usage / metric.quota) * 100;
      prc = Math.round((prc + Number.EPSILON) * 100) / 100;
      return Math.min(prc, 100);
    },
  },
};
</script>
