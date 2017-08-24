<template>
  <div class="dashboard">
    <ul id="nav" v-if="user.admin">
      <li>
        <router-link to="/settings/global">
          {{ $t('settings.globalSettings') }} <i class="material-icons">keyboard_arrow_right</i>
        </router-link>
      </li>
    </ul>

    <h1>{{ $t('settings.profileSettings') }}</h1>

    <form @submit="updateSettings">
      <h3>{{ $t('settings.language') }}</h3>
      <p><languages id="locale" :selected.sync="locale"></languages></p>
      <h3>{{ $t('settings.customStylesheet') }}</h3>
      <textarea v-model="css" name="css"></textarea>
      <p><input type="submit" :value="$t('buttons.update')"></p>
    </form>

    <form v-if="!user.lockPassword" @submit="updatePassword">
      <h3>{{ $t('settings.changePassword') }}</h3>
      <p><input :class="passwordClass" type="password" :placeholder="$t('settings.newPassword')" v-model="password" name="password"></p>
      <p><input :class="passwordClass" type="password" :placeholder="$t('settings.newPasswordConfirm')" v-model="passwordConf" name="password"></p>
      <p><input type="submit" :value="$t('buttons.update')"></p>
    </form>
  </div>
</template>

<script>
import { mapState } from 'vuex'
import { updateUser } from '@/utils/api'
import Languages from '@/components/Languages'

export default {
  name: 'settings',
  components: {
    Languages
  },
  data: function () {
    return {
      password: '',
      passwordConf: '',
      css: '',
      locale: ''
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
    this.locale = this.user.locale
  },
  methods: {
    updatePassword (event) {
      event.preventDefault()

      if (this.password !== this.passwordConf) {
        return
      }

      let user = {
        ID: this.$store.state.user.ID,
        password: this.password
      }

      updateUser(user, 'password').then(location => {
        this.$showSuccess(this.$t('settings.passwordUpdated'))
      }).catch(e => {
        this.$showError(e)
      })
    },
    updateSettings (event) {
      event.preventDefault()

      let user = {...this.$store.state.user}
      user.css = this.css
      user.locale = this.locale

      updateUser(user, 'partial').then(location => {
        this.$store.commit('setUser', user)
        this.$emit('css-updated')
        this.$showSuccess(this.$t('settings.settingsUpdated'))
      }).catch(e => {
        this.$showError(e)
      })
    }
  }
}
</script>
