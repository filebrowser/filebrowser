<template>
  <div
    class="item"
    role="button"
    tabindex="0"
    :draggable="isDraggable"
    @dragstart="dragStart"
    @dragover="dragOver"
    @drop="drop"
    @click="itemClick"
    :data-dir="isDir"
    :data-type="type"
    :aria-label="name"
    :aria-selected="isSelected"
  >
    <div>
      <img
        v-if="readOnly == undefined && type === 'image' && isThumbsEnabled"
        v-lazy="thumbnailUrl"
      />
      <i v-else class="material-icons"></i>
    </div>

    <div>
      <p class="name">{{ name }}</p>

      <p v-if="isDir" class="size" data-order="-1">&mdash;</p>
      <p v-else class="size" :data-order="humanSize()">{{ humanSize() }}</p>

      <p class="modified">
        <time :datetime="modified">{{ humanTime() }}</time>
      </p>
    </div>
  </div>
</template>

<script>
import { mapState, mapActions, mapWritableState } from "pinia";
import { useAuthStore } from "@/stores/auth";
import { useFileStore } from "@/stores/file";
import { useLayoutStore } from "@/stores/layout";

import { enableThumbs } from "@/utils/constants";
import { filesize } from "filesize";
import moment from "moment";
import { files as api } from "@/api";
import * as upload from "@/utils/upload";

export default {
  name: "item",
  compatConfig: {
    ATTR_FALSE_VALUE: "suppress-warning",
  },
  data: function () {
    return {
      touches: 0,
    };
  },
  props: [
    "name",
    "isDir",
    "url",
    "type",
    "size",
    "modified",
    "index",
    "readOnly",
    "path",
  ],
  computed: {
    ...mapState(useAuthStore, ["user", "jwt"]),
    ...mapState(useFileStore, ["req", "selectedCount", "multiple"]),
    ...mapWritableState(useFileStore, ["reload", "selected"]),
    singleClick() {
      return this.readOnly == undefined && this.user.singleClick;
    },
    isSelected() {
      return this.selected.indexOf(this.index) !== -1;
    },
    isDraggable() {
      return this.readOnly == undefined && this.user.perm.rename;
    },
    canDrop() {
      if (!this.isDir || this.readOnly !== undefined) return false;

      for (let i of this.selected) {
        if (this.req.items[i].url === this.url) {
          return false;
        }
      }

      return true;
    },
    thumbnailUrl() {
      const file = {
        path: this.path,
        modified: this.modified,
      };

      return api.getPreviewURL(file, "thumb");
    },
    isThumbsEnabled() {
      return enableThumbs;
    },
  },
  methods: {
    ...mapActions(useFileStore, ["removeSelected"]),
    ...mapActions(useLayoutStore, ["showHover", "closeHovers"]),
    humanSize: function () {
      return this.type == "invalid_link" ? "invalid link" : filesize(this.size);
    },
    humanTime: function () {
      if (this.readOnly == undefined && this.user.dateFormat) {
        return moment(this.modified).format("L LT");
      }
      return moment(this.modified).fromNow();
    },
    dragStart: function () {
      if (this.selectedCount === 0) {
        this.selected.push(this.index);
        return;
      }

      if (!this.isSelected) {
        this.selected = [];
        this.selected.push(this.index);
      }
    },
    dragOver: function (event) {
      if (!this.canDrop) return;

      event.preventDefault();
      let el = event.target;

      for (let i = 0; i < 5; i++) {
        if (!el.classList.contains("item")) {
          el = el.parentElement;
        }
      }

      el.style.opacity = 1;
    },
    drop: async function (event) {
      if (!this.canDrop) return;
      event.preventDefault();

      if (this.selectedCount === 0) return;

      let el = event.target;
      for (let i = 0; i < 5; i++) {
        if (el !== null && !el.classList.contains("item")) {
          el = el.parentElement;
        }
      }

      let items = [];

      for (let i of this.selected) {
        items.push({
          from: this.req.items[i].url,
          to: this.url + encodeURIComponent(this.req.items[i].name),
          name: this.req.items[i].name,
        });
      }

      // Get url from ListingItem instance
      let path = el.__vue__.url;
      let baseItems = (await api.fetch(path)).items;

      let action = (overwrite, rename) => {
        api
          .move(items, overwrite, rename)
          .then(() => {
            this.reload = true;
          })
          .catch(this.$showError);
      };

      let conflict = upload.checkConflict(items, baseItems);

      let overwrite = false;
      let rename = false;

      if (conflict) {
        this.showHover({
          prompt: "replace-rename",
          confirm: (event, option) => {
            overwrite = option == "overwrite";
            rename = option == "rename";

            event.preventDefault();
            this.closeHovers();
            action(overwrite, rename);
          },
        });

        return;
      }

      action(overwrite, rename);
    },
    itemClick: function (event) {
      if (this.singleClick && !this.multiple) this.open();
      else this.click(event);
    },
    click: function (event) {
      if (!this.singleClick && this.selectedCount !== 0) event.preventDefault();

      setTimeout(() => {
        this.touches = 0;
      }, 300);

      this.touches++;
      if (this.touches > 1) {
        this.open();
      }

      if (this.selected.indexOf(this.index) !== -1) {
        this.removeSelected(this.index);
        return;
      }

      if (event.shiftKey && this.selected.length > 0) {
        let fi = 0;
        let la = 0;

        if (this.index > this.selected[0]) {
          fi = this.selected[0] + 1;
          la = this.index;
        } else {
          fi = this.index;
          la = this.selected[0] - 1;
        }

        for (; fi <= la; fi++) {
          if (this.selected.indexOf(fi) == -1) {
            this.selected.push(fi);
          }
        }

        return;
      }

      if (
        !this.singleClick &&
        !event.ctrlKey &&
        !event.metaKey &&
        !this.multiple
      ) {
        this.selected = [];
      }
      this.selected.push(this.index);
    },
    open: function () {
      this.$router.push({ path: this.url });
    },
  },
};
</script>
@/stores/auth@/stores/file@/stores/layout
