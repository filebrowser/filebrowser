<template>
  <div class="dashboard">
     <form class="card" @submit="updateSettings">
      <div class="card-title">
        <h2>{{ $t('settings.profileSettings') }}</h2>
      </div>

      <div class="card-content">
        <h3>{{ $t('settings.language') }}</h3>
        <p><languages id="locale" :selected.sync="locale"></languages></p>
        <h3>{{ $t('settings.customStylesheet') }}</h3>
        <textarea v-model="css" name="css"></textarea>
      </div>

      <div class="card-action">
        <input class="flat" type="submit" :value="$t('buttons.update')">
      </div>
    </form>

    <form class="card" v-if="!user.lockPassword" @submit="updatePassword">
      <div class="card-title">
        <h2>{{ $t('settings.changePassword') }}</h2>
      </div>

      <div class="card-content">
        <p><input :class="passwordClass" type="password" :placeholder="$t('settings.newPassword')" v-model="password" name="password"></p>
        <p><input :class="passwordClass" type="password" :placeholder="$t('settings.newPasswordConfirm')" v-model="passwordConf" name="password"></p>
      </div>

      <div class="card-action">
        <input class="flat" type="submit" :value="$t('buttons.update')">
      </div>
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
        this.$emit('css')
        this.$showSuccess(this.$t('settings.settingsUpdated'))
      }).catch(e => {
        this.$showError(e)
      })
    }
  }
}
</script>
