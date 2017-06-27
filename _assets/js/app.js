'use strict'

var data = (window.data || window.alert('Something is wrong, please refresh!'))
var ssl = (window.location.protocol === 'https:')

// Remove the last directory of an url
var removeLastDirectoryPartOf = function (url) {
  var arr = url.split('/')
  if (arr.pop() === '') {
    arr.pop()
  }
  return (arr.join('/'))
}

var search = new window.Vue({
  el: '#search',
  data: {
    hover: false,
    focus: false,
    scrollable: null,
    box: null,
    input: null
  },
  mounted: function () {
    this.scrollable = document.querySelector('#search > div')
    this.box = document.querySelector('#search > div div')
    this.input = document.querySelector('#search input')
    this.reset()
  },
  methods: {
    reset: function () {
      if (data.user.AllowCommands && data.user.Commands.length > 0) {
        this.box.innerHTML = `Search or use one of your supported commands: ${data.user.Commands.join(", ")}.`
      } else {
        this.box.innerHTML = 'Type and press enter to search.'
      }
    },
    supported: function () {
      let value = this.input.value
      let pieces = value.split(' ')

      for (let i = 0; i < data.user.Commands.length; i++) {
        if (pieces[0] === data.user.Commands[0]) {
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

      if (!this.supported() || !data.user.AllowCommands) {
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
        url = removeLastDirectoryPartOf(url)
      }

      let protocol = ssl ? 'wss:' : 'ws:'

      if (this.supported() && data.user.AllowCommands) {
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
          // TODO: if is listing!
          // listing.reload()
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
      }
    }
  }
})

console.log(search)