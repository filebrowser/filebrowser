<template>
  <form class="rules small">
    <div v-for="(rule, index) in rules" :key="index">
      <input v-model="rule.regex" type="checkbox"><label>Regex</label>
      <input v-model="rule.allow" type="checkbox"><label>Allow</label>

      <input
        v-if="rule.regex"
        v-model="rule.regexp.raw"
        type="text"
        :placeholder="$t('settings.insertRegex')"
        @keypress.enter.prevent
      >
      <input
        v-else
        v-model="rule.path"
        type="text"
        :placeholder="$t('settings.insertPath')"
        @keypress.enter.prevent
      >

      <button class="button button--red" @click="remove($event, index)">-</button>
    </div>

    <div>
      <button class="button" default="false" @click="create">{{ $t('buttons.new') }}</button>
    </div>
  </form>
</template>

<script>
export default {
  name: 'RulesTextarea',
  props: ['rules'],
  methods: {
    remove(event, index) {
      event.preventDefault()
      const rules = [...this.rules]
      rules.splice(index, 1)
      this.$emit('update:rules', [...rules])
    },
    create(event) {
      event.preventDefault()

      this.$emit('update:rules', [
        ...this.rules,
        {
          allow: true,
          path: '',
          regex: false,
          regexp: {
            raw: ''
          }
        }
      ])
    }
  }
}
</script>
