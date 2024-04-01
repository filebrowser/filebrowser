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
      <button
        v-if="isSingleFile()"
        class="action copy-clipboard"
        :data-clipboard-text="linkSelected()"
        :aria-label="$t('buttons.copyDownloadLinkToClipboard')"
        :title="$t('buttons.copyDownloadLinkToClipboard')"
      >
        <i class="material-icons">content_paste</i>
      </button>
      <action
        icon="check_circle"
        :label="$t('buttons.selectMultiple')"
        @action="toggleMultipleSelection"
      />
    </header-bar>

    <breadcrumbs :base="'/share/' + hash" />

    <div v-if="loading">
      <h2 class="message delayed" style="padding-top: 3em !important;">
        <div class="spinner">
          <div class="bounce1"></div>
          <div class="bounce2"></div>
          <div class="bounce3"></div>
        </div>
        <span>{{ $t("files.loading") }}</span>
      </h2>
    </div>
    <div v-else-if="error">
      <div v-if="error.status === 401">
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
      <errors v-else :errorCode="error.status" />
    </div>
    <div v-else>
      <div class="share">
        <div class="share__box share__box__info"
          style="
            position: -webkit-sticky;
            position: sticky;
            top:-20.6em;
            z-index:999;"
        >
          <div class="share__box__header" style="height:3em">
            {{
              req.isDir
                ? $t("download.downloadFolder")
                : $t("download.downloadFile")
            }}
          </div>
          <div v-if="!this.req.isDir" class="share__box__element share__box__center share__box__icon">
            <i class="material-icons">{{ icon }}</i>
          </div>
          <div class="share__box__element" style="height:3em">
            <strong>{{ $t("prompts.displayName") }}</strong> {{ req.name }}
          </div>
          <div v-if="!this.req.isDir" class="share__box__element" :title="modTime">
            <strong>{{ $t("prompts.lastModified") }}:</strong> {{ humanTime }}
          </div>
          <div class="share__box__element" style="height:3em">
            <strong>{{ $t("prompts.size") }}:</strong> {{ humanSize }}
          </div>
          <div class="share__box__element share__box__center">
            <a target="_blank" :href="link" class="button button--flat" style="height:4em">
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
            <qrcode-vue v-if="this.req.isDir" :value="fullLink" size="100" level="M"></qrcode-vue>
          </div>
          <div v-if="!this.req.isDir" class="share__box__element share__box__center">
            <qrcode-vue :value="link" size="200" level="M"></qrcode-vue>
          </div>
      	  <div v-if="this.req.isDir" class="share__box__element share__box__header" style="height:3em">
            {{ $t("sidebar.preview") }}
          </div>
          <div
      	    v-if="this.req.isDir" 
      	    class="share__box__element share__box__center share__box__icon"
      	    style="padding:0em !important;height:12em !important;"
      	  >
      	    <a
              target="_blank"
              :href="raw"
              class="button button--flat"
      	      v-if= "!this.$store.state.multiple && 
      	             selectedCount === 1 && 
      	             req.items[this.selected[0]].type === 'image'" 
              style="height: 12em; padding:0; margin:0;"
            >
      	      <img 
      	        style="height: 12em;"
      	        :src="raw"
      	      >
            </a>
            <div
      	      v-else-if= "
      	        !this.$store.state.multiple && 
      	        selectedCount === 1 && 
      	        req.items[this.selected[0]].type === 'audio'" 
      	      style="height: 12em; paddingTop:1em; margin:0;"
      	    >
              <button @click="play" v-if="!this.tag" style="fontSize:6em !important; border:0px;outline:none; background: white;" class="material-icons">play_circle_filled</button>
              <button @click="play" v-if="this.tag"  style="fontSize:6em !important; border:0px;outline:none; background: white;" class="material-icons">pause_circle_filled</button>
      	      <audio id="myaudio"
      	        :src="raw"
      	        controls="controls" 
                :autoplay="tag"
      	      >
              </audio>
            </div>
      	    <video
      	      v-else-if= "
      	        !this.$store.state.multiple && 
      	        selectedCount === 1 && 
      	        req.items[this.selected[0]].type === 'video'" 
      	      style="height: 12em; padding:0; margin:0;"
      	      :src="raw"
      	      controls="controls" 
      	    >
      	      Sorry, your browser doesn't support embedded videos, but don't worry,
      	      you can <a :href="raw">download it</a>
      	      and watch it with your favorite video player!
      	    </video>
            <i 
      	      v-else-if= "
                !this.$store.state.multiple && 
                selectedCount === 1 &&
                req.items[this.selected[0]].isDir" 
              class="material-icons">folder
            </i>
            <i v-else class="material-icons">call_to_action</i>
          </div>
        </div>
        <div id="shareList"
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
import { filesize } from "@/utils";
import moment from "moment/min/moment-with-locales";

import HeaderBar from "@/components/header/HeaderBar.vue";
import Action from "@/components/header/Action.vue";
import Breadcrumbs from "@/components/Breadcrumbs.vue";
import Errors from "@/views/Errors.vue";
import QrcodeVue from "qrcode.vue";
import Item from "@/components/files/ListingItem.vue";
import Clipboard from "clipboard";

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
    clip: null,
    tag: false,
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
    this.clip = new Clipboard(".copy-clipboard");
    this.clip.on("success", () => {
      this.$showSuccess(this.$t("success.linkCopied"));
    });
  },
  beforeDestroy() {
    window.removeEventListener("keydown", this.keyEvent);
    this.clip.destroy();
  },
  computed: {
    ...mapState(["req", "loading", "multiple", "selected"]),
    ...mapGetters(["selectedCount"]),
    icon: function () {
      if (this.req.isDir) return "folder";
      if (this.req.type === "image") return "insert_photo";
      if (this.req.type === "audio") return "volume_up";
      if (this.req.type === "video") return "movie";
      return "insert_drive_file";
    },
    link: function () {
      return api.getDownloadURL(this.req);
    },
    raw: function () {
      return this.req.items[this.selected[0]].url.replace(/share/, 'api/public/dl')+'?token='+this.token;    
    },
    inlineLink: function () {
      return api.getDownloadURL(this.req, true);
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
    play() {
      var audio = document.getElementById('myaudio');
      if(this.tag){
        audio.pause();
        this.tag = false;
      } else {
        audio.play();
        this.tag = true;
      }
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
        file.hash = this.hash;

        this.token = file.token || "";

        this.updateRequest(file);
        document.title = `${file.name} - ${document.title}`;
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
    isSingleFile: function () {
      return (
        this.selectedCount === 1 && !this.req.items[this.selected[0]].isDir
      );
    },
    download() {
      if (this.isSingleFile()) {
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
    linkSelected: function () {
      return this.isSingleFile()
        ? api.getDownloadURL({
            hash: this.hash,
            path: this.req.items[this.selected[0]].path,
          })
        : "";
    },
  },
};
</script>
<style scoped>
  #listing.list{
    height: auto;
  }
  #shareList{
    overflow-y: scroll; 
  }
  @media (min-width: 930px) {
    #shareList{
      height: calc(100vh - 9.8em);   
      overflow-y: auto; 
    }
  }
</style>