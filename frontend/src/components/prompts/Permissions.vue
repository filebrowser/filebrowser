<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t("prompts.permissions") }}</h2>
    </div>

    <div class="card-content">
      <table class="permissions">
        <thead>
          <tr>
            <td></td>
            <td>{{ $t("prompts.read") }}</td>
            <td>{{ $t("prompts.write") }}</td>
            <td>{{ $t("prompts.execute") }}</td>
          </tr>
        </thead>
        <tbody>
          <tr class="permission-row">
            <td>{{ $t("prompts.owner") }}</td>
            <td>
              <input type="checkbox" v-model="permissions.owner.read" />
            </td>
            <td>
              <input type="checkbox" v-model="permissions.owner.write" />
            </td>
            <td>
              <input type="checkbox" v-model="permissions.owner.execute" />
            </td>
          </tr>
          <tr class="permission-row">
            <td>{{ $t("prompts.group") }}</td>
            <td>
              <input type="checkbox" v-model="permissions.group.read" />
            </td>
            <td>
              <input type="checkbox" v-model="permissions.group.write" />
            </td>
            <td>
              <input type="checkbox" v-model="permissions.group.execute" />
            </td>
          </tr>
          <tr class="permission-row">
            <td>{{ $t("prompts.others") }}</td>
            <td>
              <input type="checkbox" v-model="permissions.others.read" />
            </td>
            <td>
              <input type="checkbox" v-model="permissions.others.write" />
            </td>
            <td>
              <input type="checkbox" v-model="permissions.others.execute" />
            </td>
          </tr>
        </tbody>
      </table>
      <p>
        <code>{{ permModeString }} ({{ permMode.toString(8) }})</code>
      </p>
      <p v-if="dirSelected">
        <input type="checkbox" v-model="recursive" />
        {{ $t("prompts.recursive") }}
      </p>
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
        class="button button--flat"
        @click="chmod"
        :disabled="loading"
        :aria-label="$t('buttons.update')"
        :title="$t('buttons.update')"
      >
        {{ $t("buttons.update") }}
      </button>
    </div>
  </div>
</template>

<script>
import { mapState, mapGetters } from "vuex";
import { files as api } from "@/api";

export default {
  name: "permissions",
  data: function () {
    return {
      recursive: false,
      permissions: {
        owner: {
          read: false,
          write: false,
          execute: false,
        },
        group: {
          read: false,
          write: false,
          execute: false,
        },
        others: {
          read: false,
          write: false,
          execute: false,
        },
      },
      masks: {
        permissions: 511,
        owner: {
          read: 256,
          write: 128,
          execute: 64,
        },
        group: {
          read: 32,
          write: 16,
          execute: 8,
        },
        others: {
          read: 4,
          write: 2,
          execute: 1,
        },
      },
      loading: false,
    };
  },
  computed: {
    ...mapState(["req", "selected"]),
    ...mapGetters(["isFiles", "isListing"]),
    permMode() {
      let mode = 0;
      mode |= this.masks.owner.read * this.permissions.owner.read;
      mode |= this.masks.owner.write * this.permissions.owner.write;
      mode |= this.masks.owner.execute * this.permissions.owner.execute;
      mode |= this.masks.group.read * this.permissions.group.read;
      mode |= this.masks.group.write * this.permissions.group.write;
      mode |= this.masks.group.execute * this.permissions.group.execute;
      mode |= this.masks.others.read * this.permissions.others.read;
      mode |= this.masks.others.write * this.permissions.others.write;
      mode |= this.masks.others.execute * this.permissions.others.execute;
      return mode;
    },
    permModeString() {
      let perms = this.permMode;
      let s = "";
      s += (perms & this.masks.owner.read) != 0 ? "r" : "-";
      s += (perms & this.masks.owner.write) != 0 ? "w" : "-";
      s += (perms & this.masks.owner.execute) != 0 ? "x" : "-";
      s += (perms & this.masks.group.read) != 0 ? "r" : "-";
      s += (perms & this.masks.group.write) != 0 ? "w" : "-";
      s += (perms & this.masks.group.execute) != 0 ? "x" : "-";
      s += (perms & this.masks.others.read) != 0 ? "r" : "-";
      s += (perms & this.masks.others.write) != 0 ? "w" : "-";
      s += (perms & this.masks.others.execute) != 0 ? "x" : "-";
      return s;
    },
    dirSelected() {
      return this.req.items[this.selected[0]].isDir;
    },
  },
  created() {
    let item = this.req.items[this.selected[0]];
    let perms = item.mode & this.masks.permissions;

    // OWNER PERMS
    this.permissions.owner.read = (perms & this.masks.owner.read) != 0;
    this.permissions.owner.write = (perms & this.masks.owner.write) != 0;
    this.permissions.owner.execute = (perms & this.masks.owner.execute) != 0;
    // GROUP PERMS
    this.permissions.group.read = (perms & this.masks.group.read) != 0;
    this.permissions.group.write = (perms & this.masks.group.write) != 0;
    this.permissions.group.execute = (perms & this.masks.group.execute) != 0;
    // OTHERS PERMS
    this.permissions.others.read = (perms & this.masks.others.read) != 0;
    this.permissions.others.write = (perms & this.masks.others.write) != 0;
    this.permissions.others.execute = (perms & this.masks.others.execute) != 0;
  },
  methods: {
    cancel: function () {
      this.$store.commit("closeHovers");
    },
    chmod: async function () {
      let item = this.req.items[this.selected[0]];

      try {
        this.loading = true;
        await api.chmod(item.url, this.permMode, this.recursive);

        this.$store.commit("setReload", true);
      } catch (e) {
        this.$showError(e);
      } finally {
        this.loading = false;
      }

      this.$store.commit("closeHovers");
    },
  },
};
</script>
