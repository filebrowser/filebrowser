<template>
  <div>
    <ul class="file-list">
      <li @click="select"
        @touchstart="touchstart"
        @dblclick="next"
        role="button"
        tabindex="0"
        :aria-label="item.name"
        :aria-selected="selected == item.url"
        :key="item.name" v-for="item in items"
        :data-url="item.url">{{ item.name }}</li>
    </ul>

    <p>{{ $t('prompts.currentlyNavigating') }} <code>{{ nav }}</code>.</p>
  </div>
</template>

<script>
import { mapState } from 'vuex'
import url from '@/utils/url'
import { files } from '@/api'

export default {
  name: 'file-list',
  data: function () {
    return {
      items: [],
      touches: {
        id: '',
        count: 0
      },
      selected: null,
      current: window.location.pathname
    }
  },
  computed: {
    ...mapState([ 'req' ]),
    nav () {
      return decodeURIComponent(this.current)
    }
  },
  mounted () {
    this.fillOptions(this.req)
  },
  methods: {
    fillOptions (req) {
      // Sets the current path and resets
      // the current items.
      this.current = req.url
      this.items = []

      this.$emit('update:selected', this.current)

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
    next: function (event) {
      // Retrieves the URL of the directory the user
      // just clicked in and fill the options with its
      // content.
      let uri = event.currentTarget.dataset.url

      files.fetch(uri)
        .then(this.fillOptions)
        .catch(this.$showError)
    },
    touchstart (event) {
      let url = event.currentTarget.dataset.url

      // In 300 milliseconds, we shall reset the count.
      setTimeout(() => {
        this.touches.count = 0
      }, 300)

      // If the element the user is touching
      // is different from the last one he touched,
      // reset the count.
      if (this.touches.id !== url) {
        this.touches.id = url
        this.touches.count = 1
        return
      }

      this.touches.count++

      // If there is more than one touch already,
      // open the next screen.
      if (this.touches.count > 1) {
        this.next(event)
      }
    },
    select: function (event) {
      // If the element is already selected, unselect it.
      if (this.selected === event.currentTarget.dataset.url) {
        this.selected = null
        this.$emit('update:selected', this.current)
        return
      }

      // Otherwise select the element.
      this.selected = event.currentTarget.dataset.url
      this.$emit('update:selected', this.selected)
    }
  }
}
</script>
