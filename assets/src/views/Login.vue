<template>
  <div id="login">
    <form @submit="submit">
      <img src="../assets/logo.svg" alt="File Manager">
      <h1>File Manager</h1>
      <div v-if="wrong" class="wrong">{{ $t("login.wrongCredentials") }}</div>
      <input type="text" v-model="username" :placeholder="$t('login.username')">
      <input type="password" v-model="password" :placeholder="$t('login.password')">
      <input type="submit" :value="$t('login.submit')">
    </form>
  </div>
</template>

<script>
import auth from '@/utils/auth'

export default {
  name: 'login',
  data: function () {
    return {
      wrong: false,
      username: '',
      password: ''
    }
  },
  methods: {
    submit: function (event) {
      event.preventDefault()
      event.stopPropagation()

      let redirect = this.$route.query.redirect
      if (redirect === '' || redirect === undefined || redirect === null) {
        redirect = '/files/'
      }

      auth.login(this.username, this.password)
        .then(() => { this.$router.push({ path: redirect }) })
        .catch(() => { this.wrong = true })
    }
  }
}
</script>
