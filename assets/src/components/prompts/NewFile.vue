<template>
  <div class="prompt">
    <h3>New file</h3>
    <p>Write the name of the new file.</p>
    <input autofocus type="text" @keyup.enter="submit" v-model.trim="name">
    <div>
      <button class="ok" @click="submit">Create</button>
      <button class="cancel" @click="$store.commit('closeHovers')">Cancel</button>
    </div>
  </div>
</template>

<script>
import url from '@/utils/url'
import api from '@/utils/api'

export default {
  name: 'new-file',
  data: function () {
    return {
      name: ''
    }
  },
  methods: {
    submit: function (event) {
      event.preventDefault()
      if (this.new === '') return

      // Build the path of the new file.
      let uri = this.$route.path
      if (this.$store.state.req.kind !== 'listing') {
        uri = url.removeLastDir(uri) + '/'
      }

      uri += this.name
      uri = uri.replace('//', '/')

      // Create the new file.
      api.post(uri)
        .then(() => { this.$router.push({ path: uri }) })
        .catch(error => { this.$store.commit('showError', error) })

      // Close the prompt.
      this.$store.commit('closeHovers')
    }
  }
}
</script>

