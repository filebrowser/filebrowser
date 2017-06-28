<template>
    <div id="search" v-on:mouseleave="hover = false" v-on:click="click" v-bind:class="{ active: focus || hover }">
        <i class="material-icons" title="Search">search</i>
        <input type="text"
            v-on:focus="focus = true"
            v-on:blur="focus = false"
            v-on:keyup="keyup"
            v-on:keyup.enter="submit"
            aria-label="Write here to search" 
            placeholder="Search or execute a command...">
            <div v-on:mouseover="hover = true">
        <div>Loading...</div>
            <p><i class="material-icons spin">autorenew</i></p>
        </div>
    </div>
</template>

<script>
import page from '../page'

var $ = window.info

export default {
  name: 'search',
  data: function () {
    return {
      hover: false,
      focus: false,
      scrollable: null,
      box: null,
      input: null
    }
  },
  mounted: function () {
    this.scrollable = document.querySelector('#search > div')
    this.box = document.querySelector('#search > div div')
    this.input = document.querySelector('#search input')
    this.reset()
  },
  methods: {
    reset: function () {
      if ($.user.allowCommands && $.user.commands.length > 0) {
        this.box.innerHTML = `Search or use one of your supported commands: ${$.user.commands.join(', ')}.`
      } else {
        this.box.innerHTML = 'Type and press enter to search.'
      }
    },
    supported: function () {
      let value = this.input.value
      let pieces = value.split(' ')

      for (let i = 0; i < $.user.commands.length; i++) {
        if (pieces[0] === $.user.commands[0]) {
          return true
        }
      }

      return false
    },
    click: function (event) {
      event.currentTarget.classList.add('active')
      this.$el.querySelector('input').focus()
    },
    keyup: function (event) {
      let el = event.currentTarget

      if (el.value.length === 0) {
        this.reset()
        return
      }

      if (!this.supported() || !$.user.allowCommands) {
        this.box.innerHTML = 'Press enter to search.'
      } else {
        this.box.innerHTML = 'Press enter to execute.'
      }
    },
    submit: function (event) {
      this.box.innerHTML = ''
      this.$el.classList.add('ongoing')

      let url = window.location.host + window.location.pathname

      if (document.getElementById('editor')) {
        url = page.removeLastDir(url)
      }

      let protocol = $.ssl ? 'wss:' : 'ws:'

      if (this.supported() && $.user.allowCommands) {
        let conn = new window.WebSocket(`${protocol}//${url}?command=true`)

        conn.onopen = () => {
          conn.send(this.input.value)
        }

        conn.onmessage = (event) => {
          this.box.innerHTML = event.data
          this.scrollable.scrollTop = this.scrollable.scrollHeight
        }

        conn.onclose = (event) => {
          this.$el.classList.remove('ongoing')
          page.reload()
        }

        return
      }

      this.box.innerHTML = '<ul></ul>'

      let ul = this.box.querySelector('ul')
      let conn = new window.WebSocket(`${protocol}//${url}?search=true`)

      conn.onopen = () => {
        conn.send(this.input.value)
      }

      conn.onmessage = (event) => {
        ul.innerHTML += `<li><a href=".${event.data}">${event.data}</a></li>`
        this.scrollable.scrollTop = this.scrollable.scrollHeight
      }

      conn.onclose = () => {
        this.$el.classList.remove('ongoing')
        page.reload()
      }
    }
  }
}
</script>
