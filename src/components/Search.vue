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
                  v-for="(v,k) in boxes"
                  :key="k"
                  role="button"
                  @click="init('type:'+k)"
                  :aria-label="$t('search.'+v.label)">
                  <i class="material-icons">{{v.icon}}</i>
                  <p>{{ $t('search.'+v.label) }}</p>
                </div>
              </div>
            </div>
          </template>

        </template>
        <ul v-else-if="search.length > 0">
          <li v-for="(s,k) in results" :key="k">
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

var boxes = {
  image: { label: 'images', icon: 'insert_photo' },
  audio: { label: 'music', icon: 'volume_up' },
  video: { label: 'video', icon: 'movie' },
  pdf: { label: 'pdf', icon: 'picture_as_pdf' }
}

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
      reload: false,
      resultsCount: 50,
      boxes: boxes
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
    placeholder () {
      if (this.user.allowCommands && this.user.commands.length > 0) {
        return this.$t('search.searchOrCommand')
      }

      return this.$t('search.search')
    },
    // The text that is shown on the results' box while
    // there is no search result or command output to show.
    text () {
      if (this.ongoing) {
        return ''
      }

      if (this.value.length === 0) {
        if (this.user.allowCommands && this.user.commands.length > 0) {
          return `${this.$t('search.searchOrSupportedCommand')} ${this.user.commands.join(', ')}.`
        }

        this.$t('search.typeSearch')
      }

      if (!(this.value[0] === '$') || !this.user.allowCommands) {
        return this.$t('search.pressToSearch')
      } else {
        if (this.command.length === 0) {
          return this.$t('search.typeCommand')
        }
        if (!this.supported()) {
          return this.$t('search.notSupportedCommand')
        }
        return this.$t('search.pressToExecute')
      }
    },
    // The command, without the leading symbol ('$') with or without a following space (' ')
    command () {
      return this.value[1] === ' ' ? this.value.slice(2) : this.value.slice(1)
    },
    results () {
      return this.search.slice(0, this.resultsCount)
    }
  },
  mounted () {
    // Gets the result div which will be scrollable.
    this.scrollable = document.querySelector('#search #result')

    // Adds the keydown event on window for the ESC key, so
    // when it's pressed, it closes the search window.
    window.addEventListener('keydown', (event) => {
      if (event.keyCode === 27) {
        this.$store.commit('closeHovers')
      }
    })

    this.scrollable.addEventListener('scroll', (event) => {
      if (this.scrollable.scrollTop === (this.scrollable.scrollHeight - this.scrollable.offsetHeight)) {
        this.resultsCount += 50
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
      let cmd = this.command.split(' ')[0]
      let cl = this.user.commands.length
      if (cl !== 0) {
        for (let i = 0; i < cl; i++) {
          if (cmd.match(this.user.commands[i])) {
            return true
          }
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
      this.resultsCount = 50
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
      if (this.value[0] === '$') {
        if (this.supported() && this.user.allowCommands) {
          api.command(path, this.command,
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
        this.ongoing = false
        return
      }

      let results = []

      // In case of being a search.
      api.search(path, this.value,
        (event) => {
          let response = JSON.parse(event.data)
          if (response.path[0] === '/') {
            response.path = response.path.substring(1)
          }

          results.push(response)
        },
        (event) => {
          this.ongoing = false
          this.search = results
        }
      )
    }
  }
}
</script>
