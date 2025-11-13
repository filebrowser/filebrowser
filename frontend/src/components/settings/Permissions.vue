<template>
  <div>
    <h3>{{ $t("settings.permissions") }}</h3>
    <p class="small">{{ $t("settings.permissionsHelp") }}</p>

    <p>
      <input type="checkbox" v-model="admin" />
      {{ $t("settings.administrator") }}
    </p>

    <p>
      <input type="checkbox" :disabled="admin" v-model="perm.create" />
      {{ $t("settings.perm.create") }}
    </p>
    <p>
      <input type="checkbox" :disabled="admin" v-model="perm.delete" />
      {{ $t("settings.perm.delete") }}
    </p>
    <p>
      <input type="checkbox" :disabled="admin" v-model="perm.download" />
      {{ $t("settings.perm.download") }}
    </p>
    <p>
      <input type="checkbox" :disabled="admin" v-model="perm.modify" />
      {{ $t("settings.perm.modify") }}
    </p>
    <p v-if="isExecEnabled">
      <input type="checkbox" :disabled="admin" v-model="perm.execute" />
      {{ $t("settings.perm.execute") }}
    </p>
    <p>
      <input type="checkbox" :disabled="admin" v-model="perm.rename" />
      {{ $t("settings.perm.rename") }}
    </p>
    <p>
      <input type="checkbox" :disabled="admin" v-model="perm.share" />
      {{ $t("settings.perm.share") }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { enableExec } from "@/utils/constants";

const props = defineProps<{
  perm: UserPermissions;
}>();

const admin = computed({
  get() {
    return props.perm.admin;
  },
  set(value: boolean) {
    if (value) {
      for (const key in props.perm) {
        props.perm[key as keyof UserPermissions] = true;
      }
    }

    props.perm.admin = value;
  },
});

const isExecEnabled = enableExec;
</script>
