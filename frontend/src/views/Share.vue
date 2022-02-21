<template>
  <div>
    <header-bar showMenu showLogo>
      <title />

      <action
        v-if="selectedCount"
        icon="file_download"
        :label="$t('buttons.download')"
        @action="download"
        :counter="selectedCount"
      />
      <action
        icon="check_circle"
        :label="$t('buttons.selectMultiple')"
        @action="toggleMultipleSelection"
      />
    </header-bar>

    <breadcrumbs :base="'/share/' + hash" />

    <div v-if="loading">
      <h2 class="message delayed">
        <div class="spinner">
          <div class="bounce1"></div>
          <div class="bounce2"></div>
          <div class="bounce3"></div>
        </div>
        <span>{{ $t("files.loading") }}</span>
      </h2>
    </div>
    <div v-else-if="error">
      <div v-if="error.message === '401'">
        <div class="card floating" id="password">
          <div v-if="attemptedPasswordLogin" class="share__wrong__password">
            {{ $t("login.wrongCredentials") }}
          </div>
          <div class="card-title">
            <h2>{{ $t("login.password") }}</h2>
          </div>

          <div class="card-content">
            <input
              v-focus
              type="password"
              :placeholder="$t('login.password')"
              v-model="password"
              @keyup.enter="fetchData"
            />
          </div>
          <div class="card-action">
            <button
              class="button button--flat"
              @click="fetchData"
              :aria-label="$t('buttons.submit')"
              :title="$t('buttons.submit')"
            >
              {{ $t("buttons.submit") }}
            </button>
          </div>
        </div>
      </div>
      <errors v-else :errorCode="error.message" />
    </div>
    <div v-else>
      <div class="share">
        <div class="share__box share__box__info">
          <div class="share__box__header">
            {{
              req.isDir
                ? $t("download.downloadFolder")
                : $t("download.downloadFile")
            }}
          </div>
          <div class="share__box__element share__box__center share__box__icon">
            <i class="material-icons">{{ icon }}</i>
          </div>
          <div class="share__box__element">
            <strong>{{ $t("prompts.displayName") }}</strong> {{ req.name }}
          </div>
          <div class="share__box__element" :title="modTime">
            <strong>{{ $t("prompts.lastModified") }}:</strong> {{ humanTime }}
          </div>
          <div class="share__box__element">
            <strong>{{ $t("prompts.size") }}:</strong> {{ humanSize }}
          </div>
          <div class="share__box__element share__box__center">
            <a target="_blank" :href="link" class="button button--flat">
              <div>
                <i class="material-icons">file_download</i
                >{{ $t("buttons.download") }}
              </div>
            </a>
            <a
              target="_blank"
              :href="inlineLink"
              class="button button--flat"
              v-if="!req.isDir"
            >
              <div>
                <i class="material-icons">open_in_new</i
                >{{ $t("buttons.openFile") }}
              </div>
            </a>
          </div>
          <div class="share__box__element share__box__center">
            <qrcode-vue :value="fullLink" size="200" level="M"></qrcode-vue>
          </div>
        </div>
        <div
          v-if="req.isDir && req.items.length > 0"
          class="share__box share__box__items"
        >
          <div class="share__box__header" v-if="req.isDir">
            {{ $t("files.files") }}
          </div>
          <div id="listing" class="list file-icons">
            <item
              v-for="item in req.items.slice(0, this.showLimit)"
              :key="base64(item.name)"
              v-bind:index="item.index"
              v-bind:name="item.name"
              v-bind:isDir="item.isDir"
              v-bind:url="item.url"
              v-bind:modified="item.modified"
              v-bind:type="item.type"
              v-bind:size="item.size"
              readOnly
            >
            </item>
            <div
              v-if="req.items.length > showLimit"
              class="item"
              @click="showLimit += 100"
            >
              <div>
                <p class="name">+ {{ req.items.length - showLimit }}</p>
              </div>
            </div>

            <div
              :class="{ active: $store.state.multiple }"
              id="multiple-selection"
            >
              <p>{{ $t("files.multipleSelectionEnabled") }}</p>
              <div
                @click="$store.commit('multiple', false)"
                tabindex="0"
                role="button"
                :title="$t('files.clear')"
                :aria-label="$t('files.clear')"
                class="action"
              >
                <i class="material-icons">clear</i>
              </div>
            </div>
          </div>
        </div>
        <div
          v-else-if="req.isDir && req.items.length === 0"
          class="share__box share__box__items"
        >
          <h2 class="message">
            <i class="material-icons">sentiment_dissatisfied</i>
            <span>{{ $t("files.lonely") }}</span>
          </h2>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { mapState, mapMutations, mapGetters } from "vuex";
import { pub as api } from "@/api";
import { baseURL } from "@/utils/constants";
import filesize from "filesize";
import moment from "moment";

import HeaderBar from "@/components/header/HeaderBar";
import Action from "@/components/header/Action";
import Breadcrumbs from "@/components/Breadcrumbs";
import Errors from "@/views/Errors";
import QrcodeVue from "qrcode.vue";
import Item from "@/components/files/ListingItem";

export default {
  name: "share",
  components: {
    HeaderBar,
    Action,
    Breadcrumbs,
    Item,
    QrcodeVue,
    Errors,
  },
  data: () => ({
    error: null,
    showLimit: 100,
    password: "",
    attemptedPasswordLogin: false,
    hash: null,
    token: null,
  }),
  watch: {
    $route: function () {
      this.showLimit = 100;

      this.fetchData();
    },
  },
  created: async function () {
    const hash = this.$route.params.pathMatch.split("/")[0];
    this.hash = hash;
    await this.fetchData();
  },
  mounted() {
    window.addEventListener("keydown", this.keyEvent);
  },
  beforeDestroy() {
    window.removeEventListener("keydown", this.keyEvent);
  },
  computed: {
    ...mapState(["req", "loading", "multiple", "selected"]),
    ...mapGetters(["selectedCount", "selectedCount"]),
    icon: function () {
      if (this.req.isDir) return "folder";
      if (this.req.type === "image") return "insert_photo";
      if (this.req.type === "audio") return "volume_up";
      if (this.req.type === "video") return "movie";
      return "insert_drive_file";
    },
    link: function () {
      let queryArg = "";
      if (this.token !== "") {
        queryArg = `?token=${this.token}`;
      }

      const path = this.$route.path.split("/").splice(2).join("/");
      return `${baseURL}/api/public/dl/${path}${queryArg}`;
    },
    inlineLink: function () {
      let url = new URL(this.fullLink);
      url.searchParams.set("inline", "true");
      return url.href;
    },
    fullLink: function () {
      return window.location.origin + this.link;
    },
    humanSize: function () {
      if (this.req.isDir) {
        return this.req.items.length;
      }

      return filesize(this.req.size);
    },
    humanTime: function () {
      return moment(this.req.modified).fromNow();
    },
    modTime: function () {
      return new Date(Date.parse(this.req.modified)).toLocaleString();
    },
  },
  methods: {
    ...mapMutations(["resetSelected", "updateRequest", "setLoading"]),
    base64: function (name) {
      return window.btoa(unescape(encodeURIComponent(name)));
    },
    fetchData: async function () {
      // Reset view information.
      this.$store.commit("setReload", false);
      this.$store.commit("resetSelected");
      this.$store.commit("multiple", false);
      this.$store.commit("closeHovers");

      // Set loading to true and reset the error.
      this.setLoading(true);
      this.error = null;

      if (this.password !== "") {
        this.attemptedPasswordLogin = true;
      }

      let url = this.$route.path;
      if (url === "") url = "/";
      if (url[0] !== "/") url = "/" + url;

      try {
        let file = await api.fetch(url, this.password);

        this.token = file.token || "";

        this.updateRequest(file);
        document.title = `${file.name} - ${this.$route.name}`;
      } catch (e) {
        this.error = e;
      } finally {
        this.setLoading(false);
      }
    },
    keyEvent(event) {
      // Esc!
      if (event.keyCode === 27) {
        // If we're on a listing, unselect all
        // files and folders.
        if (this.selectedCount > 0) {
          this.resetSelected();
        }
      }
    },
    toggleMultipleSelection() {
      this.$store.commit("multiple", !this.multiple);
    },
    download() {
      if (this.selectedCount === 1 && !this.req.items[this.selected[0]].isDir) {
        api.download(
          null,
          this.hash,
          this.token,
          this.req.items[this.selected[0]].path
        );
        return;
      }

      this.$store.commit("showHover", {
        prompt: "download",
        confirm: (format) => {
          this.$store.commit("closeHovers");

          let files = [];

          for (let i of this.selected) {
            files.push(this.req.items[i].path);
          }

          api.download(format, this.hash, this.token, ...files);
        },
      });
    },
  },
};
</script>
