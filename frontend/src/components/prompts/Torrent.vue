<template>
  <div class="card floating" style="max-width: 50em;" id="share">
    <div class="card-title">
      <h2>{{ $t("buttons.torrent") }}</h2>
      <button class="button button--flat button--blue" @click="toggleDetailedView">
        {{ detailedView ? $t("buttons.compact") : $t("buttons.detailed") }}
      </button>
    </div>

    <div class="card-content">
      <!--
      <p>{{ $t("prompts.name") }}</p>
      <input 
        class="input input--block" 
        type="text" 
        v-model.trim="name" 
        tabindex="0" />
      -->

      <p v-if="detailedView">{{ $t("prompts.pieceLength") }}</p>
      <select 
        v-if="detailedView"
        class="input input--block" 
        v-model.trim="pieceLen" 
        tabindex="1">
        <option value="0">{{ $t("prompts.auto") }}</option>
        <option value="15">32 KiB</option>
        <option value="16">64 KiB</option>
        <option value="17">128 KiB</option>
        <option value="18">256 KiB</option>
        <option value="19">512 KiB</option>
        <option value="20">1 MiB</option>
        <option value="21">2 MiB</option>
        <option value="22">4 MiB</option>
        <option value="23">8 MiB</option>
        <option value="24">16 MiB</option>
        <option value="25">32 MiB</option>
        <option value="26">64 MiB</option>
        <option value="27">128 MiB</option>
        <option value="28">256 MiB</option>
      </select>

      <p>{{ $t("prompts.trackersList") }}</p>
      <textarea 
        class="input input--block input--textarea"
        style="min-height: 4em;"
        type="text" 
        v-model.trim="announces" 
        tabindex="2"></textarea>

      <p>{{ $t("prompts.webSeeds") }}</p>
      <textarea
        class="input input--block input--textarea" 
        style="min-height: 4em;"
        type="text" 
        v-model.trim="webSeeds" 
        tabindex="3"></textarea>

      <p>{{ $t("prompts.comment") }}</p>
      <textarea 
        class="input input--block input--textarea" 
        style="min-height: 4em;"
        type="text" 
        v-model.trim="comment" 
        tabindex="4"></textarea>

      <p v-if="detailedView">{{ $t("prompts.source") }}</p>
      <input 
        v-if="detailedView"
        class="input input--block" 
        type="text" 
        v-model.trim="source" 
        tabindex="5" />

      <p v-if="detailedView">
        <input type="checkbox" v-model="date" tabindex="6" />
        {{ $t("prompts.includeDate") }}
      </p>

      <p v-if="detailedView">
        <input type="checkbox" v-model="privateFlag" tabindex="7" />
        {{ $t("prompts.privateTorrent") }}
      </p>

      <div>
        <input type="checkbox" v-model="r2Flag" tabindex="8"/>
        {{ $t("prompts.r2") }}
      </div>
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
      pieceLen: 0,
      privateFlag: false,
      r2Flag: false,
      source: "",
      webSeeds: [],
      detailedView: false
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
    api.fetchDefaultOptions().then((res) => {
      this.announces = res.announces.join("\n");
      this.comment = res.comment;
      this.date = res.date;
      this.name = res.name;
      this.pieceLen = res.pieceLen;
      this.privateFlag = res.private;
      this.r2Flag = res.r2Flag;
      this.source = res.source;
      this.webSeeds = res.webSeeds.join("\n");
    });
  },
  methods: {
    ...mapActions(useLayoutStore, ["closeHovers"]),
    toggleDetailedView() {
      this.detailedView = !this.detailedView;
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
            parseInt(this.pieceLen),
            this.privateFlag,
            this.r2Flag,
            this.source,
            this.webSeeds.split("\n").map((t) => t.trim()).filter((t) => t),
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
        this.r2Flag=false;
        this.source = "";
        this.webSeeds = [];
      } catch (e) {
        this.$showError(e);
      }
    },
  },
};
</script>
