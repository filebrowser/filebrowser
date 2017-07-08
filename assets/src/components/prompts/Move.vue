<template>
  <div class="prompt">
    <h3>Move</h3>
    <p>Choose new house for your file(s)/folder(s):</p>

    <ul class="file-list">
      <li @click="select" @dblclick="next" :aria-selected="moveTo == item.url" :key="item.name" v-for="item in items" :data-url="item.url">{{ item.name }}</li>
    </ul>

    <p>Currently navigating on: <code>{{ current }}</code>.</p>

    <div>
      <button class="ok" @click="move">Move</button>
      <button class="cancel" @click="$store.commit('closeHovers')">Cancel</button>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex'
import url from '@/utils/url'
import api from '@/utils/api'
import buttons from '@/utils/buttons'

export default {
  name: 'move',
  data: function () {
    return {
      items: [],
      current: window.location.pathname,
      moveTo: null
    }
  },
  computed: mapState(['req', 'selected', 'baseURL']),
  mounted () {
    // If we're showing this on a listing,
    // we can use the current request object
    // to fill the move options.
    if (this.req.kind === 'listing') {
      this.fillOptions(this.req)
      return
    }

    // Otherwise, we must be on a preview or editor
    // so we fetch the data from the previous directory.
    api.fetch(url.removeLastDir(this.$rute.path))
      .then(this.fillOptions)
      .catch(this.showError)
  },
  methods: {
    move: function (event) {
      event.preventDefault()

      // Set the destination and create the promises array.
      let promises = []
      let dest = (this.moveTo === null) ? this.current : this.moveTo
      buttons.loading('move')

      // Create a new promise for each file.
      for (let item of this.selected) {
        let from = this.req.items[item].url
        let to = dest + '/' + encodeURIComponent(this.req.items[item].name)
        to = to.replace('//', '/')

        promises.push(api.move(from, to))
      }

      // Execute the promises.
      Promise.all(promises)
        .then(() => {
          buttons.done('move')
          this.$router.push({ path: dest })
        })
        .catch(error => {
          buttons.done('move')
          this.$store.commit('showError', error)
        })
    },
    fillOptions (req) {
      // Sets the current path and resets
      // the current items.
      this.current = req.url
      this.items = []

      // If the path isn't the root path,
      // show a button to navigate to the previous
      // directory.
      if (req.url !== '/files/') {
        this.items.push({
          name: '..',
          url: url.removeLastDir(req.url) + '/'
        })
      }

      // If this folder is empty, finish here.
      if (req.items === null) return

      // Otherwise we add every directory to the
      // move options.
      for (let item of req.items) {
        if (!item.isDir) continue

        this.items.push({
          name: item.name,
          url: item.url
        })
      }
    },
    showError (error) {
      this.$store.commit('showError', error)
    },
    next: function (event) {
      // Retrieves the URL of the directory the user
      // just clicked in and fill the options with its
      // content.
      let uri = event.currentTarget.dataset.url

      api.fetch(uri)
        .then(this.fillOptions)
        .catch(this.showError)
    },
    select: function (event) {
      // If the element is already selected, unselect it.
      if (this.moveTo === event.currentTarget.dataset.url) {
        this.moveTo = null
        return
      }

      // Otherwise select the element.
      this.moveTo = event.currentTarget.dataset.url
    }
  }
}
</script>
