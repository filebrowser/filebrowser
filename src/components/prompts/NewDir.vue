<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t('prompts.newDir') }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t('prompts.newDirMessage') }}</p>
      <input type="text" @keyup.enter="submit" v-model.trim="name" v-focus >
    </div>

    <div class="card-action">
      <button class="cancel flat"
        @click="$store.commit('closeHovers')"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')">{{ $t('buttons.cancel') }}</button>
      <button class="flat"
        :aria-label="$t('buttons.create')"
        :title="$t('buttons.create')"
        @click="submit">{{ $t('buttons.create') }}</button>
    </div>
  </div>
</template>

<script>
import url from '@/utils/url'
import * as api from '@/utils/api'

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
        .catch(this.$showError)

      // Close the prompt
      this.$store.commit('closeHovers')
    }
  }
}
</script>

