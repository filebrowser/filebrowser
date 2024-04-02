<template>
  <div>
    <h3>{{ $t("settings.permissions") }}</h3>
    <p class="small">{{ $t("settings.permissionsHelp") }}</p>

    <p>
      <input  id="uds-admin" type="checkbox" v-model="admin" />
      <label for="uds-admin">{{ $t("settings.administrator") }}</label>
    </p>

    <p>
      <input  id="uds-create" type="checkbox" :disabled="admin" v-model="perm.create" />
      <label for="uds-create">{{ $t("settings.perm.create") }}</label>
    </p>
    <p>
      <input  id="uds-delete" type="checkbox" :disabled="admin" v-model="perm.delete" />
      <label for="uds-delete">{{ $t("settings.perm.delete") }}</label>
    </p>
    <p>
      <input  id="uds-download" type="checkbox" :disabled="admin" v-model="perm.download" />
      <label for="uds-download">{{ $t("settings.perm.download") }}</label>
    </p>
    <p>
      <input  id="uds-modify" type="checkbox" :disabled="admin" v-model="perm.modify" />
      <label for="uds-modify">{{ $t("settings.perm.modify") }}</label>
    </p>
    <p v-if="isExecEnabled">
      <input  id="uds-execute" type="checkbox" :disabled="admin" v-model="perm.execute" />
      <label for="uds-execute">{{ $t("settings.perm.execute") }}</label>
    </p>
    <p>
      <input  id="uds-rename" type="checkbox" :disabled="admin" v-model="perm.rename" />
      <label for="uds-rename">{{ $t("settings.perm.rename") }}</label>
    </p>
    <p>
      <input  id="uds-share" type="checkbox" :disabled="admin" v-model="perm.share" />
      <label for="uds-share">{{ $t("settings.perm.share") }}</label>
    </p>
  </div>
</template>

<script>
import { enableExec } from "@/utils/constants";
export default {
  name: "permissions",
  props: ["perm"],
  computed: {
    admin: {
      get() {
        return this.perm.admin;
      },
      set(value) {
        if (value) {
          for (const key in this.perm) {
            this.perm[key] = true;
          }
        }

        this.perm.admin = value;
      },
    },
    isExecEnabled: () => enableExec,
  },
};
</script>
