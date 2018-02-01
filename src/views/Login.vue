<template>
  <div id="login" :class="{ recaptcha: recaptcha.length > 0 }">
    <form @submit="submit">
      <img src="../assets/logo.svg" alt="File Browser">
      <h1>File Browser</h1>
      <div v-if="wrong" class="wrong">{{ $t("login.wrongCredentials") }}</div>
      <input type="text" v-model="username" :placeholder="$t('login.username')">
      <input type="password" v-model="password" :placeholder="$t('login.password')">
      <div v-if="recaptcha.length" id="recaptcha"></div>
      <input type="submit" :value="$t('login.submit')">
    </form>
  </div>
</template>

<script>
import auth from '@/utils/auth'
import { mapState } from 'vuex'

export default {
  name: 'login',
  props: ['dependencies'],
  computed: mapState(['recaptcha']),
  data: function () {
    return {
      wrong: false,
      username: '',
      password: ''
    }
  },
  mounted () {
    if (this.dependencies) this.setup()
  },
  watch: {
    dependencies: function (val) {
      if (val) this.setup()
    }
  },
  methods: {
    setup () {
      if (this.recaptcha.length === 0) return

      window.grecaptcha.render('recaptcha', {
        sitekey: this.recaptcha
      })
    },
    submit (event) {
      event.preventDefault()
      event.stopPropagation()

      let redirect = this.$route.query.redirect
      if (redirect === '' || redirect === undefined || redirect === null) {
        redirect = '/files/'
      }

      let captcha = ''
      if (this.recaptcha.length > 0) {
        captcha = window.grecaptcha.getResponse()

        if (captcha === '') {
          this.wrong = true
          return
        }
      }

      auth.login(this.username, this.password, captcha)
        .then(() => { this.$router.push({ path: redirect }) })
        .catch(() => { this.wrong = true })
    }
  }
}
</script>
