<template>
  <div class="prompt">
    <h3>{{ $t('prompts.newDir') }}</h3>
    <p>{{ $t('prompts.newDirMessage') }}</p>
    <input autofocus type="text" @keyup.enter="submit" v-model.trim="name">
    <div>
      <button class="ok"
        :aria-label="$t('buttons.create')"
        :title="$t('buttons.create')"
        @click="submit">{{ $t('buttons.create') }}</button>
      <button class="cancel"
        @click="$store.commit('closeHovers')"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')">{{ $t('buttons.cancel') }}</button>
    </div>
  </div>
</template>

<script>
import url from '@/utils/url'
import api from '@/utils/api'

export default {
  name: 'new-dir',
  data: function () {
    return {
      name: ''
    }
  },
  methods: {
    submit: function (event) {
      event.preventDefault()
      if (this.new === '') return

      // Build the path of the new directory.
      let uri = this.$route.path
      if (this.$store.state.req.kind !== 'listing') {
        uri = url.removeLastDir(uri) + '/'
      }

      uri += this.name + '/'
      uri = uri.replace('//', '/')

      api.post(uri)
        .then(() => { this.$router.push({ path: uri }) })
        .catch(error => { this.$store.commit('showError', error) })

      // Close the prompt
      this.$store.commit('closeHovers')
    }
  }
}
</script>

