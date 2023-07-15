<template>
  <form class="rules small">
    <div v-for="(rule, index) in rules" :key="index">
      <input type="checkbox" v-model="rule.regex" /><label>Regex</label>
      <input type="checkbox" v-model="rule.allow" /><label>Allow</label>

      <input @keypress.enter.prevent type="text" style="margin-right: 8px;" v-if="rule.regex" v-model="rule.regexp.raw"
        :placeholder="$t('settings.insertRegex')" />
      <input @keypress.enter.prevent type="text" style="margin-right: 8px;" v-else v-model="rule.path"
        :placeholder="$t('settings.insertPath')" />

      <input type="checkbox" :checked="rule.perm && rule.perm.includes('read')"
        @input="changePermOfRule('read', index)" /><label>Read</label>
      <input type="checkbox" :checked="rule.perm && rule.perm.includes('write')"
        @input="changePermOfRule('write', index)" /><label>Write</label>

      <button class="button button--red" @click="remove($event, index)">
        -
      </button>
    </div>

    <div>
      <button class="button" @click="create" default="false">
        {{ $t("buttons.new") }}
      </button>
    </div>
  </form>
</template>

<script>
export default {
  name: "rules-textarea",
  props: ["rules"],
  methods: {
    changePermOfRule(ruleName, index) {
      const rule = this.rules[index];
      const isRead = ruleName === "read" ? rule.perm.includes("read") : !rule.perm.includes("read");
      const isWrite = ruleName === "write" ? rule.perm.includes("write") : !rule.perm.includes("write");
      this.rules[index].perm = `${!isRead ? "read" : ""}${!isRead && !isWrite ? "|" : ""}${!isWrite ? "write" : ""}`
    },
    remove(event, index) {
      event.preventDefault();
      let rules = [...this.rules];
      rules.splice(index, 1);
      this.$emit("update:rules", [...rules]);
    },
    create(event) {
      event.preventDefault();

      this.$emit("update:rules", [
        ...this.rules,
        {
          allow: true,
          path: "",
          regex: false,
          regexp: {
            raw: "",
          },
        },
      ]);
    },
  },
};
</script>
