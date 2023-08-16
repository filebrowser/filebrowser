<template>
  <div id="quota">
    <div>
      <div class="quota-label">{{ $t("sidebar.quota.space") }}</div>

      <br />

      <progress-bar
        :val="spaceProgress"
        size="small"
        :title="spaceProgress + '%'"
      ></progress-bar>

      <div v-if="loaded" class="quota-metric">{{ spaceUsageTitle }}</div>
    </div>

    <br />

    <div>
      <div class="quota-label">{{ $t("sidebar.quota.inodes") }}</div>

      <br />

      <progress-bar
        :val="inodeProgress"
        size="small"
        :title="inodeProgress + '%'"
      ></progress-bar>

      <div v-if="loaded" class="quota-metric">{{ inodeUsageTitle }}</div>
    </div>
  </div>
</template>

<script>
import { filesize } from "filesize";
import { mapState } from "vuex";
import ProgressBar from "vue-simple-progress";

export default {
  name: "quota",
  components: {
    ProgressBar,
  },
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
