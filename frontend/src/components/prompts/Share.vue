<template>
  <div class="card floating share__promt__card" id="share">
    <div class="card-title">
      <h2>{{ $t("buttons.share") }}</h2>
    </div>

    <template v-if="listing">
      <div class="card-content">
        <table>
          <tr>
            <th>#</th>
            <th>{{ $t("settings.shareDuration") }}</th>
            <th></th>
            <th></th>
          </tr>

          <tr v-for="link in links" :key="link.hash">
            <td>{{ link.hash }}</td>
            <td>
              <template v-if="link.expire !== 0">{{
                humanTime(link.expire)
              }}</template>
              <template v-else>{{ $t("permanent") }}</template>
            </td>
            <td class="small">
              <button
                class="action copy-clipboard"
                :data-clipboard-text="buildLink(link)"
                :aria-label="$t('buttons.copyToClipboard')"
                :title="$t('buttons.copyToClipboard')"
              >
                <i class="material-icons">content_paste</i>
              </button>
            </td>
            <td class="small">
              <button
                class="action"
                @click="deleteLink($event, link)"
                :aria-label="$t('buttons.delete')"
                :title="$t('buttons.delete')"
              >
                <i class="material-icons">delete</i>
              </button>
            </td>
          </tr>
        </table>
      </div>

      <div class="card-action">
        <button
          class="button button--flat button--grey"
          @click="$store.commit('closeHovers')"
          :aria-label="$t('buttons.close')"
          :title="$t('buttons.close')"
        >
          {{ $t("buttons.close") }}
        </button>
        <button
          class="button button--flat button--blue"
          @click="() => switchListing()"
          :aria-label="$t('buttons.new')"
          :title="$t('buttons.new')"
        >
          {{ $t("buttons.new") }}
        </button>
      </div>
    </template>

    <template v-else>
      <div class="card-content">
        <div>
          <span class="margin-tb-1em">
            <label for="customLink" style="margin-right: 5px">{{
              $t("settings.shareCustomLink")
            }}</label>
            <input
              id="customLink"
              class="input display-inline"
              type="checkbox"
              v-model="custom"
              @change="customChange"
            />
          </span>
          <input
            class="input input--block"
            :class="{ 'disable-gray': !custom }"
            :disabled="!custom"
            type="text"
            v-model="customLink"
          />
        </div>
        <p>{{ $t("settings.shareDuration") }}</p>
        <div class="input-group input">
          <input
            v-focus
            type="number"
            max="2147483647"
            min="1"
            @keyup.enter="submit"
            v-model.trim="time"
          />
          <select class="right" v-model="unit" :aria-label="$t('time.unit')">
            <option value="seconds">{{ $t("time.seconds") }}</option>
            <option value="minutes">{{ $t("time.minutes") }}</option>
            <option value="hours">{{ $t("time.hours") }}</option>
            <option value="days">{{ $t("time.days") }}</option>
          </select>
        </div>
        <p>{{ $t("prompts.optionalPassword") }}</p>
        <input
          class="input input--block"
          type="password"
          v-model.trim="password"
          autocomplete="new-password"
        />
      </div>

      <div class="card-action">
        <button
          class="button button--flat button--grey"
          @click="() => switchListing()"
          :aria-label="$t('buttons.cancel')"
          :title="$t('buttons.cancel')"
        >
          {{ $t("buttons.cancel") }}
        </button>
        <button
          class="button button--flat button--blue"
          @click="submit"
          :aria-label="$t('buttons.share')"
          :title="$t('buttons.share')"
        >
          {{ $t("buttons.share") }}
        </button>
      </div>
    </template>
  </div>
</template>

<script>
import { mapState, mapGetters } from "vuex";
import { share as api } from "@/api";
import moment from "moment";
import Clipboard from "clipboard";

export default {
  name: "share",
  data: function () {
    return {
      time: "",
      unit: "hours",
      links: [],
      clip: null,
      password: "",
      custom: false,
      customLink: "",
      listing: true,
    };
  },
  computed: {
    ...mapState(["req", "selected", "selectedCount"]),
    ...mapGetters(["isListing"]),
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
    name() {
      if (this.selectedCount === 0 || this.selectedCount > 1) {
        // This shouldn't happen.
        return;
      }
      return this.req.items[this.selected[0]].name;
    },
  },
  async beforeMount() {
    try {
      const links = await api.get(this.url);
      this.links = links;
      this.sort();

      if (this.links.length === 0) {
        this.listing = false;
      }
    } catch (e) {
      this.$showError(e);
    }
  },
  mounted() {
    this.clip = new Clipboard(".copy-clipboard");
    this.clip.on("success", () => {
      this.$showSuccess(this.$t("success.linkCopied"));
    });
    this.customLink = this.name;
  },
  beforeDestroy() {
    this.clip.destroy();
  },
  methods: {
    submit: async function () {
      let isPermanent = !this.time || this.time === 0;
      try {
        let query = {
          url: this.url,
          password: this.password,
          custom: this.custom,
          customLink: this.customLink,
        };

        if (this.custom) {
          if (!this.checkCustomLink()) {
            alert("自定义链接，只支持英文和数字");
            return;
          }
        }

        if (!isPermanent) {
          Object.assign(query, {
            expires: this.time,
            unit: this.unit,
          });
        }

        const res = await api.create(query);

        this.links.push(res);
        this.sort();

        this.time = "";
        this.unit = "hours";
        this.password = "";

        this.listing = true;
      } catch (e) {
        this.$showError(e);
      }
    },
    checkCustomLink() {
      if (this.custom) {
        return /[\w]/.test(this.customLink);
      }
      return true;
    },
    customChange(tf) {
      console.log("数据发生变化");
      if (tf) {
        this.customLink = this.name;
      }
    },
    deleteLink: async function (event, link) {
      event.preventDefault();
      try {
        await api.remove(link.hash);
        this.links = this.links.filter((item) => item.hash !== link.hash);

        if (this.links.length == 0) {
          this.listing = false;
        }
      } catch (e) {
        this.$showError(e);
      }
    },
    humanTime(time) {
      return moment(time * 1000).fromNow();
    },
    buildLink(share) {
      return api.getShareURL(share);
    },
    sort() {
      this.links = this.links.sort((a, b) => {
        if (a.expire === 0) return -1;
        if (b.expire === 0) return 1;
        return new Date(a.expire) - new Date(b.expire);
      });
    },
    switchListing() {
      if (this.links.length == 0 && !this.listing) {
        this.$store.commit("closeHovers");
      }

      this.listing = !this.listing;
    },
  },
};
</script>
<style scoped>
.display-inline {
  display: inline;
}
.margin-tb-1em {
  display: block;
  margin-bottom: 1em;
}
.disable-gray {
  background-color: #dbdbdb;
  cursor: not-allowed;
}
</style>
