<template>
  <div id="search" @click="open" v-bind:class="{ active , ongoing }">
    <div id="input">
      <button v-if="active" class="action" @click="close" :aria-label="$t('buttons.close')" :title="$t('buttons.close')">
        <i class="material-icons">arrow_back</i>
      </button>
      <i v-else class="material-icons">search</i>
      <input type="text"
        @keyup="keyup"
        @keyup.enter="submit"
        ref="input"
        :autofocus="active"
        v-model.trim="value"
        :aria-label="$t('search.writeToSearch')"
        :placeholder="placeholder">
    </div>

    <div id="result">
      <div>
        <template v-if="search.length === 0 && commands.length === 0">
          <p>{{ text }}</p>

          <template v-if="value.length === 0">
            <div class="boxes">
              <h3>{{ $t('search.types') }}</h3>
              <div>
                <div tabindex="0"
                  role="button"
                  @click="init('type:image')"
                  :aria-label="$t('search.images')">
                  <i class="material-icons">insert_photo</i>
                  <p>{{ $t('search.images') }}</p>
                </div>

                <div tabindex="0"
                  role="button"
                  @click="init('type:audio')"
                  :aria-label="$t('search.music')">
                  <i class="material-icons">volume_up</i>
                  <p>{{ $t('search.music') }}</p>
                </div>

                <div tabindex="0"
                  role="button"
                  @click="init('type:video')"
                  :aria-label="$t('search.video')">
                  <i class="material-icons">movie</i>
                  <p>{{ $t('search.video') }}</p>
                </div>

                <div tabindex="0"
                  role="button"
                  @click="init('type:pdf')"
                  :aria-label="$t('search.pdf')">
                  <i class="material-icons">picture_as_pdf</i>
                  <p>{{ $t('search.pdf') }}</p>
                </div>
              </div>
            </div>
          </template>

        </template>
        <ul v-else-if="search.length > 0">
          <li v-for="s in search">
            <router-link @click.native="close" :to="'./' + s.path">
              <i v-if="s.dir" class="material-icons">folder</i>
              <i v-else class="material-icons">insert_drive_file</i>
              <span>./{{ s.path }}</span>
            </router-link>
          </li>
        </ul>

        <pre v-else-if="commands.length > 0">
          <template v-for="c in commands">{{ c }}</template>
        </pre>
      </div>
      <p id="renew"><i class="material-icons spin">autorenew</i></p>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex'
import url from '@/utils/url'
import * as api from '@/utils/api'

export default {
  name: 'search',
  data: function () {
    return {
      value: '',
      active: false,
      ongoing: false,
      scrollable: null,
      search: [],
      commands: [],
      reload: false
    }
  },
  watch: {
    show (val, old) {
      this.active = (val === 'search')

      // If the hover was search and now it's something else
      // we should blur the input.
      if (old === 'search' && val !== 'search') {
        if (this.reload) {
          this.$store.commit('setReload', true)
        }

        document.body.style.overflow = 'auto'
        this.reset()
        this.$refs.input.blur()
      }

      // If we are starting to show the search box, we should
      // focus the input.
      if (val === 'search') {
        this.reload = false
        this.$refs.input.focus()
        document.body.style.overflow = 'hidden'
      }
    }
  },
  computed: {
    ...mapState(['user', 'show']),
    // Placeholder value.
    placeholder: function () {
      if (this.user.allowCommands && this.user.commands.length > 0) {
        return this.$t('search.searchOrCommand')
      }

      return this.$t('search.search')
    },
    // The text that is shown on the results' box while
    // there is no search result or command output to show.
    text: function () {
      if (this.ongoing) {
        return ''
      }

      if (this.value.length === 0) {
        if (this.user.allowCommands && this.user.commands.length > 0) {
          return `${this.$t('search.searchOrSupportedCommand')} ${this.user.commands.join(', ')}.`
        }

        this.$t('search.type')
      }

      if (!this.supported() || !this.user.allowCommands) {
        return this.$t('search.pressToSearch')
      } else {
        return this.$t('search.pressToExecute')
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
        this.$store.commit('closeHovers')
      }
    })
  },
  methods: {
    // Sets the search to active.
    open (event) {
      this.$store.commit('showHover', 'search')
    },
    // Closes the search and prevents the event
    // of propagating so it doesn't trigger the
    // click event on #search.
    close (event) {
      event.stopPropagation()
      event.preventDefault()
      this.$store.commit('closeHovers')
    },
    // Checks if the current input is a supported command.
    supported () {
      let pieces = this.value.split(' ')

      for (let i = 0; i < this.user.commands.length; i++) {
        if (pieces[0] === this.user.commands[i]) {
          return true
        }
      }

      return false
    },
    // Initializes the search with a default value.
    init (string) {
      this.value = string + ' '
      this.$refs.input.focus()
    },
    // Resets the search box value.
    reset () {
      this.value = ''
      this.active = false
      this.ongoing = false
      this.search = []
      this.commands = []
    },
    // When the user presses a key, if it is ESC
    // then it will close the search box. Otherwise,
    // it will set the search box to active and clean
    // the search results, as well as commands'.
    keyup (event) {
      if (event.keyCode === 27) {
        this.close(event)
        return
      }

      this.search.length = 0
      this.commands.length = 0
    },
    // Submits the input to the server and sets ongoing to true.
    submit (event) {
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
            this.reload = true
            this.ongoing = false
            this.scrollable.scrollTop = this.scrollable.scrollHeight
          }
        )

        return
      }

      // In case of being a search.
      api.search(path, this.value,
        (event) => {
          let response = JSON.parse(event.data)
          if (response.path[0] === '/') {
            response.path = response.path.substring(1)
          }

          this.search.push(response)
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
