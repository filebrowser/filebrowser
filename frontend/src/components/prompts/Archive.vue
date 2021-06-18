<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.archive") }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t("prompts.archiveMessage") }}</p>
      <input
        class="input input--block"
        v-focus
        type="text"
        @keyup.enter="submit"
        v-model.trim="name"
        required
      />

      <button
        v-for="(ext, format) in formats"
        :key="format"
        class="button button--block"
        @click="archive(format)"
        v-focus
      >
        {{ ext }}
      </button>
    </div>
  </div>
</template>

<script>
import { mapState, mapGetters } from "vuex";
import { files as api } from "@/api";
import url from "@/utils/url";
import buttons from "@/utils/buttons";

export default {
  name: "archive",
  data: function () {
    return {
      name: "",
      formats: {
        zip: "zip",
        tar: "tar",
        targz: "tar.gz",
        tarbz2: "tar.bz2",
        tarxz: "tar.xz",
        tarlz4: "tar.lz4",
        tarsz: "tar.sz",
      },
    };
  },
  computed: {
    ...mapState(["req", "selected"]),
    ...mapGetters(["isFiles", "isListing"]),
  },
  methods: {
    cancel: function () {
      this.$store.commit("closeHovers");
    },
    archive: async function (format) {
      let items = [];

      for (let i of this.selected) {
        items.push(this.req.items[i].name);
      }

      let uri = this.isFiles ? this.$route.path : "/";

      if (!this.isListing) {
        uri = url.removeLastDir(uri);
      }

      uri += "/archive";
      uri = uri.replace("//", "/");

      try {
        buttons.loading("archive");
        await api.archive(uri, this.name, format, ...items);
        this.$store.commit("closeHovers");
        this.$store.commit("setReload", true);
        this.$store.dispatch("quota/fetch", 3000);
      } catch (e) {
        this.$showError(e);
      } finally {
        buttons.done("archive");
      }
    },
  },
};
</script>
