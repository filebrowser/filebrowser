<template>
  <div id="search" @click="active = true" v-bind:class="{ active , ongoing }">
    <div id="input">
      <button v-if="active" class="action" @click="close" >
        <i class="material-icons">arrow_back</i>
      </button>
      <i v-else class="material-icons">search</i>
      <input type="text"
        v-model.trim="value"
        @keyup="keyup"
        @keyup.enter="submit"
        aria-label="Write here to search"
        :placeholder="placeholder()">
    </div>
    <div id="result">
      <div>
        <span v-if="search.length === 0 && commands.length === 0">{{ text() }}</span>
        <ul v-else-if="search.length > 0">
          <li v-for="s in search"><router-link :to="'./' + s">./{{ s }}</router-link></li>
        </ul>
        <ul v-else-if="commands.length > 0">
          <li v-for="c in commands">{{ c }}</li>
        </ul>
      </div>
      <p><i class="material-icons spin">autorenew</i></p>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex'
import url from '@/utils/url'
import api from '@/utils/api'

export default {
  name: 'search',
  data: function () {
    return {
      value: '',
      active: false,
      ongoing: false,
      scrollable: null,
      search: [],
      commands: []
    }
  },
  computed: mapState(['user']),
  mounted: function () {
    this.scrollable = document.querySelector('#search #result')

    window.addEventListener('keydown', (event) => {
      // Esc!
      if (event.keyCode === 27) {
        this.active = false
      }
    })
  },
  methods: {
    close: function (event) {
      event.stopPropagation()
      event.preventDefault()
      this.active = false
    },
    placeholder: function () {
      if (this.user.allowCommands && this.user.commands.length > 0) {
        return 'Search or execute a command...'
      }

      return 'Search...'
    },
    text: function () {
      if (this.ongoing) {
        return ''
      }

      if (this.value.length === 0) {
        if (this.user.allowCommands && this.user.commands.length > 0) {
          return `Search or use one of your supported commands: ${this.user.commands.join(', ')}.`
        }

        return 'Type and press enter to search.'
      }

      if (!this.supported() || !this.user.allowCommands) {
        return 'Press enter to search.'
      } else {
        return 'Press enter to execute.'
      }
    },
    keyup: function (event) {
      if (event.keyCode === 27) {
        this.active = false
        return
      }

      this.active = true
      this.search.length = 0
      this.commands.length = 0
    },
    supported: function () {
      let pieces = this.value.split(' ')

      for (let i = 0; i < this.user.commands.length; i++) {
        if (pieces[0] === this.user.commands[0]) {
          return true
        }
      }

      return false
    },
    click: function (event) {
      event.currentTarget.classList.add('active')
      this.$el.querySelector('input').focus()
    },
    submit: function (event) {
      this.ongoing = true

      let path = this.$route.path
      if (this.$store.state.req.kind !== 'listing') {
        path = url.removeLastDir(path) + '/'
      }

      if (this.supported() && this.user.allowCommands) {
        api.command(path, this.value,
          (event) => {
            this.commands.push(event.data)
            this.scrollable.scrollTop = this.scrollable.scrollHeight
          },
          (event) => {
            this.ongoing = false
            this.scrollable.scrollTop = this.scrollable.scrollHeight
            this.$store.commit('setReload', true)
          }
        )

        return
      }

      api.search(path, this.value,
        (event) => {
          let url = event.data
          if (url[0] === '/') url = url.substring(1)

          this.search.push(url)
          this.scrollable.scrollTop = this.scrollable.scrollHeight
        },
        (event) => {
          this.ongoing = false
          this.scrollable.scrollTop = this.scrollable.scrollHeight
        }
      )
    }
  }
}
</script>
