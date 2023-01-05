<template>
  <div>
    <h3>{{ $t("settings.userCommands") }}</h3>
    <p class="small">
      {{ $t("settings.userCommandsHelp") }} <i>git svn hg</i>.
    </p>
    <input class="input input--block" type="text" v-model.trim="raw" />
  </div>
</template>

<script>
export default {
  name: "permissions",
  props: ["commands"],
  computed: {
    raw: {
      get() {
        return this.commands.join(" ");
      },
      set(value) {
        if (value !== "") {
          this.$emit("update:commands", value.split(" "));
        } else {
          this.$emit("update:commands", []);
        }
      },
    },
  },
};
</script>
