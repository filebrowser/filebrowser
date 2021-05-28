<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.unarchive") }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t("prompts.unarchiveMessage") }}</p>
      <input
        class="input input--block"
        v-focus
        type="text"
        @keyup.enter="submit"
        v-model.trim="name"
      />
    </div>

    <div class="card-action">
      <button
        class="button button--flat button--grey"
        @click="$store.commit('closeHovers')"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')"
      >
        {{ $t("buttons.cancel") }}
      </button>
      <button
        @click="submit"
        class="button button--flat"
        type="submit"
        :aria-label="$t('buttons.unarchive')"
        :title="$t('buttons.unarchive')"
        :disabled="loading"
      >
        {{ $t("buttons.unarchive") }}
      </button>
    </div>
  </div>
</template>

<script>
import { mapState, mapGetters } from "vuex";
import { files as api } from "@/api";

export default {
  name: "rename",
  data: function () {
    return {
      loading: false,
      name: "",
    };
  },
  computed: {
    ...mapState(["req", "selected", "selectedCount"]),
    ...mapGetters(["isListing", "isFiles"]),
  },
  methods: {
    cancel: function () {
      this.$store.commit("closeHovers");
    },
    submit: async function () {
      let item = this.req.items[this.selected[0]];
      let uri = this.isFiles ? this.$route.path + "/" : "/";
      let dst = uri + this.name;
      dst = dst.replace("//", "/");

      try {
        this.loading = true;
        await api.unarchive(item.url, dst, false);

        this.$store.commit("setReload", true);
      } catch (e) {
        this.$showError(e);
      } finally {
        this.loading = true;
      }

      this.$store.commit("closeHovers");
    },
  },
};
</script>
