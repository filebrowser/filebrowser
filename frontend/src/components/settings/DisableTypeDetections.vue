<template>
  <form class="rules small">
    <div v-for="(disableTypeDetection, index) in disableTypeDetections" :key="index">
      <input type="checkbox" v-model="disableTypeDetection.regex"><label>Regex</label>

      <input
        @keypress.enter.prevent
        type="text"
        v-if="disableTypeDetection.regex"
        v-model="disableTypeDetection.regexp.raw"
        :placeholder="$t('settings.insertRegex')" />
      <input
        @keypress.enter.prevent
        type="text"
        v-else
        v-model="disableTypeDetection.path"
        :placeholder="$t('settings.insertPath')" />

      <button class="button button--red" @click="remove($event, index)">-</button>
    </div>

    <div>
      <button class="button" @click="create" default="false">{{ $t('buttons.new') }}</button>
    </div>
  </form>
</template>

<script>
export default {
  name: 'disableTypeDetections-textarea',
  props: ['disableTypeDetections'],
  methods: {
    remove (event, index) {
      event.preventDefault()
      let disableTypeDetections = [ ...this.disableTypeDetections ]
      disableTypeDetections.splice(index, 1)
      this.$emit('update:disableTypeDetections', [ ...disableTypeDetections ])
    },
    create (event) {
      event.preventDefault()

      this.$emit('update:disableTypeDetections', [
        ...this.disableTypeDetections,
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
