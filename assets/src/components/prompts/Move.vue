<template>
  <div class="prompt">
    <h3>Move</h3>
    <p>Choose new house for your file(s)/folder(s):</p>

    <file-list @update:selected="val => dest = val"></file-list>

    <div>
      <button class="ok" @click="move">Move</button>
      <button class="cancel" @click="$store.commit('closeHovers')">Cancel</button>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex'
import FileList from './FileList'
import api from '@/utils/api'
import buttons from '@/utils/buttons'

export default {
  name: 'move',
  components: { FileList },
  data: function () {
    return {
      current: window.location.pathname,
      dest: null
    }
  },
  computed: mapState(['req', 'selected']),
  methods: {
    move: function (event) {
      event.preventDefault()
      buttons.loading('move')
      let items = []

      // Create a new promise for each file.
      for (let item of this.selected) {
        items.push({
          from: this.req.items[item].url,
          to: this.dest + encodeURIComponent(this.req.items[item].name)
        })
      }

      // Execute the promises.
      api.move(items)
        .then(() => {
          buttons.done('move')
          this.$router.push({ path: this.dest })
        })
        .catch(error => {
          buttons.done('move')
          this.$store.commit('showError', error)
        })

      event.preventDefault()
    }
  }
}
</script>
