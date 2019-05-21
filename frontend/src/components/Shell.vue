<template>
  <div @click="focus" class="shell" ref="scrollable" :class="{ ['shell--hidden']: !showShell}">
    <div v-for="(c, index) in content" :key="index" class="shell__result" >
      <div class="shell__prompt"><i class="material-icons">chevron_right</i></div>
      <pre class="shell__text">{{ c.text }}</pre>
    </div>

    <div class="shell__result" :class="{ 'shell__result--hidden': !canInput }" >
      <div class="shell__prompt"><i class="material-icons">chevron_right</i></div>
      <pre
        tabindex="0"
        ref="input"
        class="shell__text"
        contenteditable="true"
        @keydown.prevent.38="historyUp"
        @keydown.prevent.40="historyDown"
        @keypress.prevent.enter="submit" />
    </div>
  </div>
</template>

<script>
import { mapMutations, mapState, mapGetters } from 'vuex'
import { commands } from '@/api'

export default {
  name: 'shell',
  computed: {
    ...mapState([ 'user', 'showShell' ]),
    ...mapGetters([ 'isFiles', 'isLogged' ]),
    path: function () {
      if (this.isFiles) {
        return this.$route.path
      }

      return ''
    }
  },
  data: () => ({
    content: [],
    history: [],
    historyPos: 0,
    canInput: true
  }),
  methods: {
    ...mapMutations([ 'toggleShell' ]),
    scroll: function () {
      this.$refs.scrollable.scrollTop = this.$refs.scrollable.scrollHeight
    },
    focus: function () {
      this.$refs.input.focus()
    },
    historyUp () {
      if (this.historyPos > 0) {
        this.$refs.input.innerText = this.history[--this.historyPos]
        this.focus()
      }
    },
    historyDown () {
      if (this.historyPos >= 0 && this.historyPos < this.history.length - 1) {
        this.$refs.input.innerText = this.history[++this.historyPos]
        this.focus()
      } else {
        this.historyPos = this.history.length
        this.$refs.input.innerText = ''
      }
    },
    submit: function (event) {
      const cmd = event.target.innerText.trim()

      if (cmd === '') {
        return
      }

      if (cmd === 'clear') {
        this.content = []
        event.target.innerHTML = ''
        return
      }

      if (cmd === 'exit') {
        event.target.innerHTML = ''
        this.toggleShell()
        return
      }

      this.canInput = false
      event.target.innerHTML = ''

      let results = {
        text: `${cmd}\n\n`
      }
      
      this.history.push(cmd)
      this.historyPos = this.history.length
      this.content.push(results)

      commands(
        this.path,
        cmd,
        event => {
          results.text += `${event.data}\n`
          this.scroll()
        },
        () => {
          results.text = results.text.trimEnd()
          this.canInput = true
          this.$refs.input.focus()
          this.scroll()
        }
      )
    }
  }
}
</script>
