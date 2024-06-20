<template>
  <div class="card floating" style="max-width: 50em;" id="share">
    <div class="card-title">
      <h2>{{ $t("buttons.torrent") }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t("prompts.trackersList") }}</p>
      <textarea 
        class="input input--block input--textarea"
        style="min-height: 5em;"
        type="text" 
        v-model.trim="announces" 
        tabindex="1"></textarea>

      <p>{{ $t("prompts.webSeeds") }}</p>
      <textarea 
        class="input input--block input--textarea" 
        style="min-height: 5em;"
        type="text" 
        v-model.trim="webSeeds" 
        tabindex="2"></textarea>

      <p>{{ $t("prompts.comment") }}</p>
      <textarea 
        class="input input--block input--textarea" 
        style="min-height: 5em;"
        type="text" 
        v-model.trim="comment" 
        tabindex="3"></textarea>

      <p>{{ $t("prompts.source") }}</p>
      <input 
        class="input input--block" 
        type="text" 
        v-model.trim="source" 
        tabindex="4" />

      <label>
        <input type="checkbox" v-model="privateFlag" tabindex="3" />
        {{ $t("prompts.privateTorrent") }}
      </label>
    </div>

    <div class="card-action">
      <button class="button button--flat button--grey" @click="closeHovers" :aria-label="$t('buttons.close')"
        :title="$t('buttons.close')" tabindex="2">
        {{ $t("buttons.close") }}
      </button>
      <button id="focus-prompt" class="button button--flat button--blue" @click="torrent"
        :aria-label="$t('buttons.torrent')" :title="$t('buttons.torrent')" tabindex="1">
        {{ $t("buttons.torrent") }}
      </button>
    </div>
  </div>
</template>

<script>
import { mapActions, mapState, mapWritableState } from "pinia";
import { useFileStore } from "@/stores/file";
import { torrent as api } from "@/api";
import { useLayoutStore } from "@/stores/layout";
import buttons from "@/utils/buttons";

export default {
  name: "torrent",
  data: function () {
    return {
      announces: [],
      comment: "",
      date: true,
      name: "",
      pieceLen: 18,
      privateFlag: false,
      source: "",
      webSeeds: [],
    };
  },
  inject: ["$showError", "$showSuccess"],
  computed: {
    ...mapState(useFileStore, [
      "req",
      "selected",
      "selectedCount",
      "isListing",
    ]),
    ...mapWritableState(useFileStore, ["reload"]),
    url() {
      if (!this.isListing) {
        return this.$route.path;
      }

      if (this.selectedCount === 0 || this.selectedCount > 1) {
        // This shouldn't happen.
        return;
      }

      return this.req.items[this.selected[0]].url;
    },
  },
  async beforeMount() {
    this.fetchTrackers();
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers"]),
    async fetchTrackers() {
      const url = 'https://cf.trackerslist.com/all.txt';
      try {
        const response = await fetch(url);
        if (!response.ok) {
          throw new Error(this.$t("error.failedToFetchTrackers"));
        }
        let text = await response.text();
        text = text.replace(/^\s*[\r\n]/gm, '');
        this.announces = text;
      } catch (error) {
        this.$showError(error);
        text = this.$t("error.faildToFetchTrackers");
        this.announces = text;
      }
    },
    torrent: async function (event) {
      event.preventDefault();
      try {
        if (!this.announces.length) {
          this.$showError(this.$t("error.noTrackers"));
          return;
        }

        let action = async () => {
          buttons.loading("torrent");
          const res = await api.makeTorrent(
            this.url,
            this.announces.split("\n").map((t) => t.trim()).filter((t) => t),
            this.comment,
            this.date,
            this.name,
            this.pieceLen,
            this.privateFlag,
            this.source,
            this.webSeeds,
          ).then(
            () => {
              buttons.success("torrent");
              this.closeHovers();
              this.reload = true
              this.$showSuccess(this.$t("success.torrentCreated"));
            }
          ).catch((e) => {
              buttons.done("torrent");
              this.$showError(e);
            });
        }

        this.closeHovers();
        action();

        this.announces = [];
        this.comment = "";
        this.date = true;
        this.pieceLen = 18;
        this.privateFlag = false;
        this.source = "";
        this.webSeeds = [];
      } catch (e) {
        this.$showError(e);
      }
    },
  },
};
</script>
