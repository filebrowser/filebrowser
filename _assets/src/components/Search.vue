<template>
  <div id="search" v-on:mouseleave="hover = false" v-on:click="click" v-bind:class="{ active: focus || hover, ongoing }">
    <i class="material-icons" title="Search">search</i>
    <input type="text"
      v-model.trim="value"
      v-on:focus="focus = true"
      v-on:blur="focus = false"
      v-on:keyup="keyup"
      v-on:keyup.enter="submit"
      aria-label="Write here to search"
      :placeholder="placeholder()">
    <div v-on:mouseover="hover = true">
      <div>
        <span v-if="search.length === 0 && commands.length === 0">{{ text() }}</span>
        <ul v-else-if="search.length > 0">
          <li v-for="s in search"><a :href="'.' + s">.{{ s }}</a></li>
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
import page from '../utils/page'

export default {
  name: 'search',
  data: function () {
    return {
      value: '',
      hover: false,
      focus: false,
      ongoing: false,
      scrollable: null,
      search: [],
      commands: []
    }
  },
  computed: mapState(['user']),
  mounted: function () {
    this.scrollable = document.querySelector('#search > div')
  },
  methods: {
    placeholder: function () {
      if (this.user.allowCommands && this.user.commands.length > 0) {
        return 'Search or execute a command...'
      }

      return 'Search...'
    },
    text: function () {
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
    keyup: function () {
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
      let uri = window.location.host + window.location.pathname

      if (this.$store.state.req.kind !== 'listing') {
        uri = page.removeLastDir(uri)
      }

      uri = `${(this.$store.state.ssl ? 'wss:' : 'ws:')}//${uri}`

      if (this.supported() && this.user.allowCommands) {
        let conn = new window.WebSocket(`${uri}?command=true`)

        conn.onopen = () => conn.send(this.value)

        conn.onmessage = (event) => {
          this.commands.push(event.data)
          this.scrollable.scrollTop = this.scrollable.scrollHeight
        }

        conn.onclose = (event) => {
          this.ongoing = false
          this.scrollable.scrollTop = this.scrollable.scrollHeight
          page.reload()
        }

        return
      }

      let conn = new window.WebSocket(`${uri}?search=true`)

      conn.onopen = () => conn.send(this.value)

      conn.onmessage = (event) => {
        this.search.push(event.data)
        this.scrollable.scrollTop = this.scrollable.scrollHeight
      }

      conn.onclose = () => {
        this.ongoing = false
        this.scrollable.scrollTop = this.scrollable.scrollHeight
      }
    }
  }
}
</script>
