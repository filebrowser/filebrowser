<template>
  <div class="dashboard">
     <form class="card" @submit="updateSettings">
      <div class="card-title">
        <h2>{{ $t('settings.profileSettings') }}</h2>
      </div>

      <div class="card-content">
        <h3>{{ $t('settings.language') }}</h3>
        <languages class="input input--block" :locale.sync="locale"></languages>
      </div>

      <div class="card-action">
        <input class="button button--flat" type="submit" :value="$t('buttons.update')">
      </div>
    </form>

    <form class="card" v-if="!user.lockPassword" @submit="updatePassword">
      <div class="card-title">
        <h2>{{ $t('settings.changePassword') }}</h2>
      </div>

      <div class="card-content">
        <input :class="passwordClass" type="password" :placeholder="$t('settings.newPassword')" v-model="password" name="password">
        <input :class="passwordClass" type="password" :placeholder="$t('settings.newPasswordConfirm')" v-model="passwordConf" name="password">
      </div>

      <div class="card-action">
        <input class="button button--flat" type="submit" :value="$t('buttons.update')">
      </div>
    </form>
  </div>
</template>

<script>
import { mapState, mapMutations } from 'vuex'
import { users as api } from '@/api'
import Languages from '@/components/settings/Languages'

export default {
  name: 'settings',
  components: {
    Languages
  },
  data: function () {
    return {
      password: '',
      passwordConf: '',
      locale: ''
    }
  },
  computed: {
    ...mapState([ 'user' ]),
    passwordClass () {
      const baseClass = 'input input--block'

      if (this.password === '' && this.passwordConf === '') {
        return baseClass
      }

      if (this.password === this.passwordConf) {
        return `${baseClass} input--green`
      }

      return `${baseClass} input--red`
    }
  },
  created () {
    this.locale = this.user.locale
  },
  methods: {
    ...mapMutations([ 'updateUser' ]),
    async updatePassword (event) {
      event.preventDefault()

      if (this.password !== this.passwordConf || this.password === '') {
        return
      }

      try {
        const data = { id: this.user.id, password: this.password }
        await api.update(data, ['password'])
        this.updateUser(data)
        this.$showSuccess(this.$t('settings.passwordUpdated'))
      } catch (e) {
        this.$showError(e)
      }
    },
    async updateSettings (event) {
      event.preventDefault()

      try {
        const data = { id: this.user.id, locale: this.locale }
        await api.update(data, ['locale'])
        this.updateUser(data)
        this.$showSuccess(this.$t('settings.settingsUpdated'))
      } catch (e) {
        this.$showError(e)
      }
    }
  }
}
</script>
