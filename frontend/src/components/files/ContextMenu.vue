<template>
  <div
    id="context-menu"
    ref="contextMenu"
    class="card"
    :style="menuPosition"
    @click="close"
  >
    <p>
      <action icon="info" :label="$t('buttons.info')" show="info" />
    </p>
    <p v-if="options.share">
      <action icon="share" :label="$t('buttons.share')" show="share" />
    </p>
    <p v-if="options.rename">
      <action icon="mode_edit" :label="$t('buttons.rename')" show="rename" />
    </p>
    <p v-if="options.copy">
      <action
        id="copy-button"
        icon="content_copy"
        :label="$t('buttons.copyFile')"
        show="copy"
      />
    </p>
    <p v-if="options.move">
      <action
        id="move-button"
        icon="forward"
        :label="$t('buttons.moveFile')"
        show="move"
      />
    </p>
    <p v-if="options.permissions">
      <action
        id="permissions-button"
        icon="lock"
        :label="$t('buttons.permissions')"
        show="permissions"
      />
    </p>
    <p v-if="options.archive">
      <action
        id="archive-button"
        icon="archive"
        :label="$t('buttons.archive')"
        show="archive"
      />
    </p>
    <p v-if="options.unarchive">
      <action
        id="unarchive-button"
        icon="unarchive"
        :label="$t('buttons.unarchive')"
        show="unarchive"
      />
    </p>
    <p v-if="options.download">
      <action
        icon="file_download"
        :label="$t('buttons.download')"
        @action="download"
        :counter="selectedCount"
      />
    </p>
    <p v-if="options.delete">
      <action
        id="delete-button"
        icon="delete"
        :label="$t('buttons.delete')"
        show="delete"
      />
    </p>
  </div>
</template>

<script>
import { mapState, mapGetters } from "vuex";
import { files as api } from "@/api";
import Action from "@/components/header/Action";

export default {
  name: "context-menu",
  components: { Action },
  computed: {
    ...mapState(["req", "selected", "user", "selected", "contextMenu"]),
    ...mapGetters(["selectedCount", "onlyArchivesSelected"]),
    menuPosition() {
      if (this.contextMenu === null) {
        return { left: "0px", right: "0px" };
      }

      let style = {
        left: this.contextMenu.x + "px",
        top: this.contextMenu.y + "px",
      };

      if (window.innerWidth - this.contextMenu.x < 150) {
        style.transform = "translateX(calc(-100% - 3px))";
      }

      return style;
    },
    options() {
      return {
        download: this.user.perm.download,
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
  },
  mounted() {
    window.addEventListener("mousedown", this.windowClick);
  },
  beforeDestroy() {
    window.removeEventListener("mousedown", this.windowClick);
  },
  methods: {
    windowClick(event) {
      if (!this.$refs.contextMenu.contains(event.target)) {
        this.close();
      }
    },
    close() {
      this.$store.commit("hideContextMenu");
    },
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
  },
};
</script>
