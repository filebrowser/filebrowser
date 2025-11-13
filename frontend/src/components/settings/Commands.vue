<template>
  <div>
    <h3>{{ $t("settings.userCommands") }}</h3>
    <p class="small">
      {{ $t("settings.userCommandsHelp") }} <i>git svn hg</i>.
    </p>
    <input class="input input--block" type="text" v-model.trim="raw" />
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";

const props = defineProps<{
  commands: string[];
}>();

const emit = defineEmits<{
  "update:commands": [commands: string[]];
}>();

const raw = computed({
  get() {
    return props.commands.join(" ");
  },
  set(value: string) {
    if (value !== "") {
      emit("update:commands", value.split(" "));
    } else {
      emit("update:commands", []);
    }
  },
});
</script>
