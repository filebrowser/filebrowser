<template>
  <form class="rules small">
    <div v-for="(rule, index) in rules" :key="index">
      <input type="checkbox" v-model="rule.regex" /><label>Regex</label>
      <input type="checkbox" v-model="rule.allow" /><label>Allow</label>

      <input
        @keypress.enter.prevent
        type="text"
        v-if="rule.regex"
        v-model="rule.regexp.raw"
        :placeholder="$t('settings.insertRegex')"
      />
      <input
        @keypress.enter.prevent
        type="text"
        v-else
        v-model="rule.path"
        :placeholder="$t('settings.insertPath')"
      />

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

<script setup lang="ts">
interface Rule {
  allow: boolean;
  path: string;
  regex: boolean;
  regexp: {
    raw: string;
  };
}

const props = defineProps<{
  rules: Rule[];
}>();

const emit = defineEmits<{
  "update:rules": [rules: Rule[]];
}>();

const remove = (event: Event, index: number) => {
  event.preventDefault();
  const rules = [...props.rules];
  rules.splice(index, 1);
  emit("update:rules", [...rules]);
};

const create = (event: Event) => {
  event.preventDefault();

  emit("update:rules", [
    ...props.rules,
    {
      allow: true,
      path: "",
      regex: false,
      regexp: {
        raw: "",
      },
    },
  ]);
};
</script>
