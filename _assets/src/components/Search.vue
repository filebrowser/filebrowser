<template>
  <div id="search" @click="open" v-bind:class="{ active , ongoing }">
    <div id="input">
      <button v-if="active" class="action" @click="close">
        <i class="material-icons">arrow_back</i>
      </button>
      <i v-else class="material-icons">search</i>
      <input type="text"
        @keyup="keyup"
        @keyup.enter="submit"
        v-model.trim="value"
        aria-label="Write here to search"
        :placeholder="placeholder">
    </div>

    <div id="result">
      <div>
        <span v-if="search.length === 0 && commands.length === 0">{{ text }}</span>
        <ul v-else-if="search.length > 0">
          <li v-for="s in search">
            <router-link @click.native="close" :to="'./' + s">./{{ s }}</router-link>
          </li>
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
  computed: {
    ...mapState(['user']),
    // Placeholder value.
    placeholder: function () {
      if (this.user.allowCommands && this.user.commands.length > 0) {
        return 'Search or execute a command...'
      }

      return 'Search...'
    },
    // The text that is shown on the results' box while
    // there is no search result or command output to show.
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
    }
  },
  mounted: function () {
    // Gets the result div which will be scrollable.
    this.scrollable = document.querySelector('#search #result')

    // Adds the keydown event on window for the ESC key, so
    // when it's pressed, it closes the search window.
    window.addEventListener('keydown', (event) => {
      if (event.keyCode === 27) {
        this.active = false
      }
    })
  },
  methods: {
    // Sets the search to active.
    open: function (event) {
      this.active = true
    },
    // Closes the search and prevents the event
    // of propagating so it doesn't trigger the
    // click event on #search.
    close: function (event) {
      event.stopPropagation()
      event.preventDefault()
      this.active = false
    },
    // Checks if the current input is a supported command.
    supported: function () {
      let pieces = this.value.split(' ')

      for (let i = 0; i < this.user.commands.length; i++) {
        if (pieces[0] === this.user.commands[0]) {
          return true
        }
      }

      return false
    },
    // When the user presses a key, if it is ESC
    // then it will close the search box. Otherwise,
    // it will set the search box to active and clean
    // the search results, as well as commands'.
    keyup: function (event) {
      if (event.keyCode === 27) {
        this.close(event)
        return
      }

      this.active = true
      this.search.length = 0
      this.commands.length = 0
    },
    // Submits the input to the server and sets ongoing to true.
    submit: function (event) {
      this.ongoing = true

      let path = this.$route.path
      if (this.$store.state.req.kind !== 'listing') {
        path = url.removeLastDir(path) + '/'
      }

      // In case of being a command.
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

      // In case of being a search.
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
