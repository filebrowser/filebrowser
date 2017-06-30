<template>
  <div class="prompt">
    <h3>Move</h3>
    <p>Choose new house for your file(s)/folder(s):</p>

    <ul class="file-list">
      <li @click="select" @dblclick="next" :key="item.name" v-for="item in items" :data-url="item.url">{{ item.name }}</li>
    </ul>

    <p>Currently navigating on: <code>{{ current }}</code>.</p>

    <div>
      <button class="ok" @click="move">Move</button>
      <button class="cancel" @click="$store.commit('showMove', false)">Cancel</button>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex'
import page from '../utils/page'
import webdav from '../utils/webdav'

export default {
  name: 'move-prompt',
  data: function () {
    return {
      items: [],
      current: window.location.pathname
    }
  },
  computed: mapState(['req', 'selected', 'baseURL']),
  mounted: function () {
    if (window.location.pathname !== this.baseURL + '/') {
      this.items.push({
        name: '..',
        url: page.removeLastDir(window.location.pathname) + '/'
      })
    }

    if (this.req.kind === 'listing') {
      for (let item of this.req.data.items) {
        if (!item.isDir) continue

        this.items.push({
          name: item.name,
          url: item.url
        })
      }

      return
    }
  },
  methods: {
    move: function (event) {
      event.preventDefault()

      let el = event.currentTarget
      let promises = []
      let dest = this.current
      // buttons.setLoading('move')

      let selected = el.querySelector('li[aria-selected=true]')
      if (selected !== null) {
        dest = selected.dataset.url
      }

      for (let item of this.selected) {
        let from = this.req.data.items[item].url
        let to = dest + '/' + this.req.data.items[item].name
        to = to.replace('//', '/')

        promises.push(webdav.move(from, to))
      }

      this.$store.commit('showMove', false)

      Promise.all(promises)
        .then(() => {
          // buttons.setDone('move')
          page.open(dest)
        })
        .catch(e => {
          // buttons.setDone('move', false)
          console.log(e)
        })
    },
    next: function (event) {
      let url = event.currentTarget.dataset.url
      this.json(url)
        .then((data) => {
          this.current = url
          this.items = []

          if (url !== this.baseURL + '/') {
            this.items.push({
              name: '..',
              url: page.removeLastDir(url) + '/'
            })
          }

          let req = JSON.parse(data)
          for (let item of req.data.items) {
            if (!item.isDir) continue

            this.items.push({
              name: item.name,
              url: item.url
            })
          }
        })
        .catch(e => console.log(e))
    },
    json: function (url) {
      return new Promise((resolve, reject) => {
        let request = new XMLHttpRequest()
        request.open('GET', url)
        request.setRequestHeader('Accept', 'application/json')
        request.onload = () => {
          if (request.status === 200) {
            resolve(request.responseText)
          } else {
            reject(request.statusText)
          }
        }
        request.onerror = () => reject(request.statusText)
        request.send()
      })
    },
    select: function (event) {
      let el = event.currentTarget

      if (el.getAttribute('aria-selected') === 'true') {
        el.setAttribute('aria-selected', false)
        return
      }

      let el2 = this.$el.querySelector('li[aria-selected=true]')
      if (el2) {
        el2.setAttribute('aria-selected', false)
      }

      el.setAttribute('aria-selected', true)
      return
    }
  }
}
</script>
