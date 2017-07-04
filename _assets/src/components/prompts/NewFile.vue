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

      let uri = this.$route.path
      if (this.$store.state.req.kind !== 'listing') {
        uri = url.removeLastDir(uri) + '/'
      }

      uri += this.name
      uri = uri.replace('//', '/')

      api.post(uri)
        .then(() => {
          this.$router.push({ path: uri })
        })
        .catch(error => {
          // TODO: show error message in a box
          console.log(error)
        })

      this.$store.commit('closeHovers')
    }
  }
}
</script>

