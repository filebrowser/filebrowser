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
      <button class="cancel" @click="$store.commit('closeHovers')">Cancel</button>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex'
import url from '@/utils/url'
import api from '@/utils/api'

export default {
  name: 'move',
  data: function () {
    return {
      items: [],
      current: window.location.pathname
    }
  },
  computed: mapState(['req', 'selected', 'baseURL']),
  mounted: function () {
    if (this.$route.path !== '/files/') {
      this.items.push({
        name: '..',
        url: url.removeLastDir(this.$route.path) + '/'
      })
    }

    if (this.req.kind === 'listing') {
      for (let item of this.req.items) {
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
        let from = this.req.items[item].url
        let to = dest + '/' + encodeURIComponent(this.req.items[item].name)
        to = to.replace('//', '/')

        promises.push(api.move(from, to))
      }

      this.$store.commit('showMove', false)

      Promise.all(promises)
        .then(() => {
          // buttons.setDone('move')
          this.$router.push({page: dest})
        })
        .catch(e => {
          // buttons.setDone('move', false)
          console.log(e)
        })
    },
    next: function (event) {
      let uri = event.currentTarget.dataset.url
      this.json(uri)
        .then((data) => {
          this.current = uri
          this.items = []

          if (uri !== this.baseURL + '/') {
            this.items.push({
              name: '..',
              url: url.removeLastDir(uri) + '/'
            })
          }

          let req = JSON.parse(data)
          for (let item of req.items) {
            if (!item.isDir) continue

            this.items.push({
              name: item.name,
              url: item.uri
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
