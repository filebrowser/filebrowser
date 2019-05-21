<template>
  <div>
    <p v-if="!isDefault">
      <label for="username">{{ $t('settings.username') }}</label>
      <input class="input input--block" type="text" v-model="user.username" id="username">
    </p>

    <p v-if="!isDefault">
      <label for="password">{{ $t('settings.password') }}</label>
      <input class="input input--block" type="password" :placeholder="passwordPlaceholder" v-model="user.password" id="password">
    </p>

    <p>
      <label for="scope">{{ $t('settings.scope') }}</label>
      <input class="input input--block" type="text" v-model="user.scope" id="scope">
    </p>

    <p>
      <label for="locale">{{ $t('settings.language') }}</label>
      <languages class="input input--block" id="locale" :locale.sync="user.locale"></languages>
    </p>

    <p v-if="!isDefault">
      <input type="checkbox" :disabled="user.perm.admin" v-model="user.lockPassword"> {{ $t('settings.lockPassword') }}
    </p>

    <permissions :perm.sync="user.perm" />
    <commands :commands.sync="user.commands" />

    <div v-if="!isDefault">
      <h3>{{ $t('settings.rules') }}</h3>
      <p class="small">{{ $t('settings.rulesHelp') }}</p>
      <rules :rules.sync="user.rules" />
    </div>
  </div>
</template>

<script>
import Languages from './Languages'
import Rules from './Rules'
import Permissions from './Permissions'
import Commands from './Commands'

export default {
  name: 'user',
  components: {
    Permissions,
    Languages,
    Rules,
    Commands
  },
  props: [ 'user', 'isNew', 'isDefault' ],
  computed: {
    passwordPlaceholder () {
      return this.isNew ? '' : this.$t('settings.avoidChanges')
    }
  },
  watch: {
    'user.perm.admin': function () {
      if (!this.user.perm.admin) return
      this.user.lockPassword = false
    }
  }
}
</script>
