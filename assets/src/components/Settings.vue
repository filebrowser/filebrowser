<template>
  <div class="dashboard">
    <h1>Settings</h1>

    <router-link v-if="user.admin" to="/users">Go to User Management</router-link>

    <form @submit="changePassword">
      <h2>Change Password</h2>
      <p><input :class="passwordClass" type="password" placeholder="Your new password" v-model="password" name="password"></p>
      <p><input :class="passwordClass" type="password" placeholder="Confirm your new password" v-model="passwordConf" name="password"></p>
      <p><input type="submit" value="Change Password"></p>
    </form>

    <form @submit="updateCSS">
      <h2>Costum Stylesheet</h2>
      <textarea v-model="css" name="css"></textarea>
      <p><input type="submit" value="Update"></p>
    </form>
  </div>
</template>

<script>
import { mapState, mapMutations } from 'vuex'
import api from '@/utils/api'

export default {
  name: 'settings',
  data: function () {
    return {
      password: '',
      passwordConf: '',
      css: ''
    }
  },
  computed: {
    ...mapState([ 'user' ]),
    passwordClass () {
      if (this.password === '' && this.passwordConf === '') {
        return ''
      }

      if (this.password === this.passwordConf) {
        return 'green'
      }

      return 'red'
    }
  },
  created () {
    this.css = this.user.css
  },
  methods: {
    ...mapMutations([ 'showSuccess' ]),
    changePassword (event) {
      event.preventDefault()

      if (this.password !== this.passwordConf) {
        return
      }

      api.updatePassword(this.password).then(() => {
        this.showSuccess('Password updated!')
      }).catch(e => {
        this.$store.commit('showError', e)
      })
    },
    updateCSS (event) {
      event.preventDefault()

      api.updateCSS(this.css).then(() => {
        this.$store.commit('setUserCSS', this.css)
        this.$emit('css-updated')
        this.showSuccess('Styles updated!')
      }).catch(e => {
        this.$store.commit('showError', e)
      })
    }
  }
}
</script>
