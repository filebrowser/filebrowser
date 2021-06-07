<template>
  <div>
    <header-bar showMenu showLogo>
      <search /> <title />
      <action
        class="search-button"
        icon="search"
        :label="$t('buttons.search')"
        @action="openSearch()"
      />

      <template #actions>
        <template v-if="!isMobile">
          <action
            v-if="headerButtons.share"
            icon="share"
            :label="$t('buttons.share')"
            show="share"
          />
          <action
            v-if="headerButtons.rename"
            icon="mode_edit"
            :label="$t('buttons.rename')"
            show="rename"
          />
          <action
            v-if="headerButtons.copy"
            id="copy-button"
            icon="content_copy"
            :label="$t('buttons.copyFile')"
            show="copy"
          />
          <action
            v-if="headerButtons.move"
            id="move-button"
            icon="forward"
            :label="$t('buttons.moveFile')"
            show="move"
          />
          <action
            v-if="headerButtons.permissions"
            id="permissions-button"
            icon="lock"
            :label="$t('buttons.permissions')"
            show="permissions"
          />
          <action
            v-if="headerButtons.archive"
            id="archive-button"
            icon="archive"
            :label="$t('buttons.archive')"
            show="archive"
          />
          <action
            v-if="headerButtons.unarchive"
            id="unarchive-button"
            icon="unarchive"
            :label="$t('buttons.unarchive')"
            show="unarchive"
          />
          <action
            v-if="headerButtons.delete"
            id="delete-button"
            icon="delete"
            :label="$t('buttons.delete')"
            show="delete"
          />
        </template>

        <action
          v-if="headerButtons.shell"
          icon="code"
          :label="$t('buttons.shell')"
          @action="$store.commit('toggleShell')"
        />
        <action
          :icon="user.viewMode === 'mosaic' ? 'view_list' : 'view_module'"
          :label="$t('buttons.switchView')"
          @action="switchView"
        />
        <action
          v-if="headerButtons.download"
          icon="file_download"
          :label="$t('buttons.download')"
          @action="download"
          :counter="selectedCount"
        />
        <action
          v-if="headerButtons.upload"
          icon="file_upload"
          id="upload-button"
          :label="$t('buttons.upload')"
          @action="upload"
        />
        <action icon="info" :label="$t('buttons.info')" show="info" />
        <action
          icon="check_circle"
          :label="$t('buttons.selectMultiple')"
          @action="toggleMultipleSelection"
        />
      </template>
    </header-bar>

    <div v-if="isMobile" id="file-selection">
      <span v-if="selectedCount > 0">{{ selectedCount }} selected</span>
      <action
        v-if="headerButtons.share"
        icon="share"
        :label="$t('buttons.share')"
        show="share"
      />
      <action
        v-if="headerButtons.rename"
        icon="mode_edit"
        :label="$t('buttons.rename')"
        show="rename"
      />
      <action
        v-if="headerButtons.copy"
        icon="content_copy"
        :label="$t('buttons.copyFile')"
        show="copy"
      />
      <action
        v-if="headerButtons.move"
        icon="forward"
        :label="$t('buttons.moveFile')"
        show="move"
      />
      <action
        v-if="headerButtons.delete"
        icon="delete"
        :label="$t('buttons.delete')"
        show="delete"
      />
    </div>

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
    <template v-else>
      <div v-if="req.numDirs + req.numFiles == 0">
        <h2 class="message">
          <i class="material-icons">sentiment_dissatisfied</i>
          <span>{{ $t("files.lonely") }}</span>
        </h2>
        <input
          style="display: none"
          type="file"
          id="upload-input"
          @change="uploadInput($event)"
          multiple
        />
        <input
          style="display: none"
          type="file"
          id="upload-folder-input"
          @change="uploadInput($event)"
          webkitdirectory
          multiple
        />
      </div>
      <div v-else id="listing" ref="listing" :class="user.viewMode">
        <div>
          <div class="item header">
            <div></div>
            <div>
              <p
                :class="{ active: nameSorted }"
                class="name"
                role="button"
                tabindex="0"
                @click="sort('name')"
                :title="$t('files.sortByName')"
                :aria-label="$t('files.sortByName')"
              >
                <span>{{ $t("files.name") }}</span>
                <i class="material-icons">{{ nameIcon }}</i>
              </p>

              <p
                :class="{ active: sizeSorted }"
                class="size"
                role="button"
                tabindex="0"
                @click="sort('size')"
                :title="$t('files.sortBySize')"
                :aria-label="$t('files.sortBySize')"
              >
                <span>{{ $t("files.size") }}</span>
                <i class="material-icons">{{ sizeIcon }}</i>
              </p>
              <p
                :class="{ active: modifiedSorted }"
                class="modified"
                role="button"
                tabindex="0"
                @click="sort('modified')"
                :title="$t('files.sortByLastModified')"
                :aria-label="$t('files.sortByLastModified')"
              >
                <span>{{ $t("files.lastModified") }}</span>
                <i class="material-icons">{{ modifiedIcon }}</i>
              </p>
            </div>
          </div>
        </div>

        <h2 v-if="req.numDirs > 0">{{ $t("files.folders") }}</h2>
        <div v-if="req.numDirs > 0">
          <item
            v-for="item in dirs"
            :key="base64(item.name)"
            v-bind:index="item.index"
            v-bind:name="item.name"
            v-bind:isSymlink="item.isSymlink"
            v-bind:link="item.link"
            v-bind:isDir="item.isDir"
            v-bind:url="item.url"
            v-bind:mode="item.mode"
            v-bind:modified="item.modified"
            v-bind:type="item.type"
            v-bind:size="item.size"
          >
          </item>
        </div>

        <h2 v-if="req.numFiles > 0">{{ $t("files.files") }}</h2>
        <div v-if="req.numFiles > 0">
          <item
            v-for="item in files"
            :key="base64(item.name)"
            v-bind:index="item.index"
            v-bind:name="item.name"
            v-bind:isSymlink="item.isSymlink"
            v-bind:link="item.link"
            v-bind:isDir="item.isDir"
            v-bind:url="item.url"
            v-bind:mode="item.mode"
            v-bind:modified="item.modified"
            v-bind:type="item.type"
            v-bind:size="item.size"
          >
          </item>
        </div>

        <input
          style="display: none"
          type="file"
          id="upload-input"
          @change="uploadInput($event)"
          multiple
        />
        <input
          style="display: none"
          type="file"
          id="upload-folder-input"
          @change="uploadInput($event)"
          webkitdirectory
          multiple
        />

        <div :class="{ active: $store.state.multiple }" id="multiple-selection">
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
    </template>
  </div>
</template>

<script>
import Vue from "vue";
import { mapState, mapGetters, mapMutations } from "vuex";
import { users, files as api } from "@/api";
import { enableExec } from "@/utils/constants";
import * as upload from "@/utils/upload";
import css from "@/utils/css";
import throttle from "lodash.throttle";

import HeaderBar from "@/components/header/HeaderBar";
import Action from "@/components/header/Action";
import Search from "@/components/Search";
import Item from "@/components/files/ListingItem";

export default {
  name: "listing",
  components: {
    HeaderBar,
    Action,
    Search,
    Item,
  },
  data: function () {
    return {
      showLimit: 50,
      dragCounter: 0,
      width: window.innerWidth,
      itemWeight: 0,
    };
  },
  computed: {
    ...mapState([
      "req",
      "selected",
      "user",
      "show",
      "multiple",
      "selected",
      "loading",
    ]),
    ...mapGetters(["selectedCount", "onlyArchivesSelected"]),
    nameSorted() {
      return this.req.sorting.by === "name";
    },
    sizeSorted() {
      return this.req.sorting.by === "size";
    },
    modifiedSorted() {
      return this.req.sorting.by === "modified";
    },
    ascOrdered() {
      return this.req.sorting.asc;
    },
    items() {
      const dirs = [];
      const files = [];

      this.req.items.forEach((item) => {
        if (item.isDir) {
          dirs.push(item);
        } else {
          files.push(item);
        }
      });

      return { dirs, files };
    },
    dirs() {
      return this.items.dirs.slice(0, this.showLimit);
    },
    files() {
      let showLimit = this.showLimit - this.items.dirs.length;

      if (showLimit < 0) showLimit = 0;

      return this.items.files.slice(0, showLimit);
    },
    nameIcon() {
      if (this.nameSorted && !this.ascOrdered) {
        return "arrow_upward";
      }

      return "arrow_downward";
    },
    sizeIcon() {
      if (this.sizeSorted && this.ascOrdered) {
        return "arrow_downward";
      }

      return "arrow_upward";
    },
    modifiedIcon() {
      if (this.modifiedSorted && this.ascOrdered) {
        return "arrow_downward";
      }

      return "arrow_upward";
    },
    headerButtons() {
      return {
        upload: this.user.perm.create,
        download: this.user.perm.download,
        shell: this.user.perm.execute && enableExec,
        delete: this.selectedCount > 0 && this.user.perm.delete,
        rename: this.selectedCount === 1 && this.user.perm.rename,
        share: this.selectedCount === 1 && this.user.perm.share,
        move: this.selectedCount > 0 && this.user.perm.rename,
        copy: this.selectedCount > 0 && this.user.perm.create,
        permissions: this.selectedCount === 1 && this.user.perm.modify,
        archive: this.selectedCount > 0 && this.user.perm.create,
        unarchive: this.selectedCount === 1 && this.onlyArchivesSelected,
      };
    },
    isMobile() {
      return this.width <= 736;
    },
  },
  watch: {
    req: function () {
      // Reset the show value
      this.showLimit = 50;

      // Ensures that the listing is displayed
      Vue.nextTick(() => {
        // How much every listing item affects the window height
        this.setItemWeight();

        // Fill and fit the window with listing items
        this.fillWindow(true);
      });
    },
  },
  mounted: function () {
    // Check the columns size for the first time.
    this.colunmsResize();

    // How much every listing item affects the window height
    this.setItemWeight();

    // Fill and fit the window with listing items
    this.fillWindow(true);

    // Add the needed event listeners to the window and document.
    window.addEventListener("keydown", this.keyEvent);
    window.addEventListener("scroll", this.scrollEvent);
    window.addEventListener("resize", this.windowsResize);

    if (!this.user.perm.create) return;
    document.addEventListener("dragover", this.preventDefault);
    document.addEventListener("dragenter", this.dragEnter);
    document.addEventListener("dragleave", this.dragLeave);
    document.addEventListener("drop", this.drop);
  },
  beforeDestroy() {
    // Remove event listeners before destroying this page.
    window.removeEventListener("keydown", this.keyEvent);
    window.removeEventListener("scroll", this.scrollEvent);
    window.removeEventListener("resize", this.windowsResize);

    if (this.user && !this.user.perm.create) return;
    document.removeEventListener("dragover", this.preventDefault);
    document.removeEventListener("dragenter", this.dragEnter);
    document.removeEventListener("dragleave", this.dragLeave);
    document.removeEventListener("drop", this.drop);
  },
  methods: {
    ...mapMutations(["updateUser", "addSelected"]),
    base64: function (name) {
      return window.btoa(unescape(encodeURIComponent(name)));
    },
    keyEvent(event) {
      // No prompts are shown
      if (this.show !== null) {
        return;
      }

      // Esc!
      if (event.keyCode === 27) {
        // Reset files selection.
        this.$store.commit("resetSelected");
        this.$store.commit("hideContextMenu");
      }

      // Del!
      if (event.keyCode === 46) {
        if (!this.user.perm.delete || this.selectedCount == 0) return;

        // Show delete prompt.
        this.$store.commit("showHover", "delete");
      }

      // F2!
      if (event.keyCode === 113) {
        if (!this.user.perm.rename || this.selectedCount !== 1) return;

        // Show rename prompt.
        this.$store.commit("showHover", "rename");
      }

      // Ctrl is pressed
      if (!event.ctrlKey && !event.metaKey) {
        return;
      }

      let key = String.fromCharCode(event.which).toLowerCase();

      switch (key) {
        case "f":
          event.preventDefault();
          this.$store.commit("showHover", "search");
          break;
        case "c":
        case "x":
          this.copyCut(event, key);
          break;
        case "v":
          this.paste(event);
          break;
        case "a":
          event.preventDefault();
          for (let file of this.items.files) {
            if (this.$store.state.selected.indexOf(file.index) === -1) {
              this.addSelected(file.index);
            }
          }
          for (let dir of this.items.dirs) {
            if (this.$store.state.selected.indexOf(dir.index) === -1) {
              this.addSelected(dir.index);
            }
          }
          break;
        case "s":
          event.preventDefault();
          document.getElementById("download-button").click();
          break;
      }
    },
    preventDefault(event) {
      // Wrapper around prevent default.
      event.preventDefault();
    },
    copyCut(event, key) {
      if (event.target.tagName.toLowerCase() === "input") {
        return;
      }

      let items = [];

      for (let i of this.selected) {
        items.push({
          from: this.req.items[i].url,
          name: this.req.items[i].name,
        });
      }

      if (items.length == 0) {
        return;
      }

      this.$store.commit("updateClipboard", {
        key: key,
        items: items,
        path: this.$route.path,
      });
    },
    paste(event) {
      if (event.target.tagName.toLowerCase() === "input") {
        return;
      }

      let items = [];

      for (let item of this.$store.state.clipboard.items) {
        const from = item.from.endsWith("/")
          ? item.from.slice(0, -1)
          : item.from;
        const to = this.$route.path + encodeURIComponent(item.name);
        items.push({ from, to, name: item.name });
      }

      if (items.length === 0) {
        return;
      }

      let action = (overwrite, rename) => {
        api
          .copy(items, overwrite, rename)
          .then(() => {
            this.$store.commit("setReload", true);
          })
          .catch(this.$showError);
      };

      if (this.$store.state.clipboard.key === "x") {
        action = (overwrite, rename) => {
          api
            .move(items, overwrite, rename)
            .then(() => {
              this.$store.commit("resetClipboard");
              this.$store.commit("setReload", true);
            })
            .catch(this.$showError);
        };
      }

      if (this.$store.state.clipboard.path == this.$route.path) {
        action(false, true);

        return;
      }

      let conflict = upload.checkConflict(items, this.req.items);

      let overwrite = false;
      let rename = false;

      if (conflict) {
        this.$store.commit("showHover", {
          prompt: "replace-rename",
          confirm: (event, option) => {
            overwrite = option == "overwrite";
            rename = option == "rename";

            event.preventDefault();
            this.$store.commit("closeHovers");
            action(overwrite, rename);
          },
        });

        return;
      }

      action(overwrite, rename);
    },
    colunmsResize() {
      // Update the columns size based on the window width.
      let columns = Math.floor(
        document.querySelector("main").offsetWidth / 300
      );
      let items = css(["#listing.mosaic .item", ".mosaic#listing .item"]);
      if (columns === 0) columns = 1;
      items.style.width = `calc(${100 / columns}% - 1em)`;
    },
    scrollEvent: throttle(function () {
      const totalItems = this.req.numDirs + this.req.numFiles;

      // All items are displayed
      if (this.showLimit >= totalItems) return;

      const currentPos = window.innerHeight + window.scrollY;

      // Trigger at the 75% of the window height
      const triggerPos = document.body.offsetHeight - window.innerHeight * 0.25;

      if (currentPos > triggerPos) {
        // Quantity of items needed to fill 2x of the window height
        const showQuantity = Math.ceil(
          (window.innerHeight * 2) / this.itemWeight
        );

        // Increase the number of displayed items
        this.showLimit += showQuantity;
      }
    }, 100),
    dragEnter() {
      this.dragCounter++;

      // When the user starts dragging an item, put every
      // file on the listing with 50% opacity.
      let items = document.getElementsByClassName("item");

      Array.from(items).forEach((file) => {
        file.style.opacity = 0.5;
      });
    },
    dragLeave() {
      this.dragCounter--;

      if (this.dragCounter == 0) {
        this.resetOpacity();
      }
    },
    drop: async function (event) {
      event.preventDefault();
      this.dragCounter = 0;
      this.resetOpacity();

      let dt = event.dataTransfer;
      let el = event.target;

      if (dt.files.length <= 0) return;

      for (let i = 0; i < 5; i++) {
        if (el !== null && !el.classList.contains("item")) {
          el = el.parentElement;
        }
      }

      let files = await upload.scanFiles(dt);
      let items = this.req.items;
      let path = this.$route.path.endsWith("/")
        ? this.$route.path
        : this.$route.path + "/";

      if (
        el !== null &&
        el.classList.contains("item") &&
        el.dataset.dir === "true"
      ) {
        // Get url from ListingItem instance
        path = el.__vue__.url;

        try {
          items = (await api.fetch(path)).items;
        } catch (error) {
          this.$showError(error);
        }
      }

      let conflict = upload.checkConflict(files, items);

      if (conflict) {
        this.$store.commit("showHover", {
          prompt: "replace",
          confirm: (event) => {
            event.preventDefault();
            this.$store.commit("closeHovers");
            upload.handleFiles(files, path, true);
          },
        });

        return;
      }

      upload.handleFiles(files, path);
    },
    uploadInput(event) {
      this.$store.commit("closeHovers");

      let files = event.currentTarget.files;
      let folder_upload =
        files[0].webkitRelativePath !== undefined &&
        files[0].webkitRelativePath !== "";

      if (folder_upload) {
        for (let i = 0; i < files.length; i++) {
          let file = files[i];
          files[i].fullPath = file.webkitRelativePath;
        }
      }

      let path = this.$route.path.endsWith("/")
        ? this.$route.path
        : this.$route.path + "/";
      let conflict = upload.checkConflict(files, this.req.items);

      if (conflict) {
        this.$store.commit("showHover", {
          prompt: "replace",
          confirm: (event) => {
            event.preventDefault();
            this.$store.commit("closeHovers");
            upload.handleFiles(files, path, true);
          },
        });

        return;
      }

      upload.handleFiles(files, path);
    },
    resetOpacity() {
      let items = document.getElementsByClassName("item");

      Array.from(items).forEach((file) => {
        file.style.opacity = 1;
      });
    },
    async sort(by) {
      let asc = false;

      if (by === "name") {
        if (this.nameIcon === "arrow_upward") {
          asc = true;
        }
      } else if (by === "size") {
        if (this.sizeIcon === "arrow_upward") {
          asc = true;
        }
      } else if (by === "modified") {
        if (this.modifiedIcon === "arrow_upward") {
          asc = true;
        }
      }

      try {
        await users.update({ id: this.user.id, sorting: { by, asc } }, [
          "sorting",
        ]);
      } catch (e) {
        this.$showError(e);
      }

      this.$store.commit("setReload", true);
    },
    openSearch() {
      this.$store.commit("showHover", "search");
    },
    toggleMultipleSelection() {
      this.$store.commit("multiple", !this.multiple);
      this.$store.commit("closeHovers");
    },
    windowsResize: throttle(function () {
      this.colunmsResize();
      this.width = window.innerWidth;

      // Listing element is not displayed
      if (this.$refs.listing == null) return;

      // How much every listing item affects the window height
      this.setItemWeight();

      // Fill but not fit the window
      this.fillWindow();
    }, 100),
    download() {
      if (this.selectedCount === 1 && !this.req.items[this.selected[0]].isDir) {
        api.download(null, this.req.items[this.selected[0]].url);
        return;
      }

      this.$store.commit("showHover", {
        prompt: "download",
        confirm: (format) => {
          this.$store.commit("closeHovers");

          let files = [];

          if (this.selectedCount > 0) {
            for (let i of this.selected) {
              files.push(this.req.items[i].url);
            }
          } else {
            files.push(this.$route.path);
          }

          api.download(format, ...files);
        },
      });
    },
    switchView: async function () {
      this.$store.commit("closeHovers");

      const data = {
        id: this.user.id,
        viewMode: this.user.viewMode === "mosaic" ? "list" : "mosaic",
      };

      users.update(data, ["viewMode"]).catch(this.$showError);

      // Await ensures correct value for setItemWeight()
      await this.$store.commit("updateUser", data);

      this.setItemWeight();
      this.fillWindow();
    },
    upload: function () {
      if (
        typeof window.DataTransferItem !== "undefined" &&
        typeof DataTransferItem.prototype.webkitGetAsEntry !== "undefined"
      ) {
        this.$store.commit("showHover", "upload");
      } else {
        document.getElementById("upload-input").click();
      }
    },
    setItemWeight() {
      // Listing element is not displayed
      if (this.$refs.listing == null) return;

      let itemQuantity = this.req.numDirs + this.req.numFiles;
      if (itemQuantity > this.showLimit) itemQuantity = this.showLimit;

      // How much every listing item affects the window height
      this.itemWeight = this.$refs.listing.offsetHeight / itemQuantity;
    },
    fillWindow(fit = false) {
      const totalItems = this.req.numDirs + this.req.numFiles;

      // More items are displayed than the total
      if (this.showLimit >= totalItems && !fit) return;

      const windowHeight = window.innerHeight;

      // Quantity of items needed to fill 2x of the window height
      const showQuantity = Math.ceil(
        (windowHeight + windowHeight * 2) / this.itemWeight
      );

      // Less items to display than current
      if (this.showLimit > showQuantity && !fit) return;

      // Set the number of displayed items
      this.showLimit = showQuantity > totalItems ? totalItems : showQuantity;
    },
  },
};
</script>
