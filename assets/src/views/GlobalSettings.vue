<template>
  <div class="dashboard">
    <ul id="nav">
      <li>
        <router-link to="/settings/profile">
          <i class="material-icons">keyboard_arrow_left</i> {{ $t('settings.profileSettings') }}
        </router-link>
      </li>
      <li>
        <router-link to="/users">
          {{ $t('settings.userManagement') }} <i class="material-icons">keyboard_arrow_right</i>
        </router-link>
      </li>
    </ul>

    <h1>{{ $t('settings.globalSettings') }}</h1>

    <form @submit="saveStaticGen" v-if="$store.state.staticGen.length > 0">
      <h2>{{ capitalize($store.state.staticGen) }}</h2>

      <p v-for="field in staticGen" :key="field.variable">
        <label v-if="field.type !== 'checkbox'">{{ field.name }}</label>
        <input v-if="field.type === 'text'" type="text" v-model.trim="field.value">
        <input v-else-if="field.type === 'checkbox'" type="checkbox" v-model.trim="field.value">
        <template v-if="field.type === 'checkbox'">{{ capitalize(field.name, 'caps') }}</template>
      </p>

      <p><input type="submit" value="Save"></p>
    </form>

    <form @submit="saveCommands">
      <h2>{{ $t('settings.commands') }}</h2>

      <p class="small">{{ $t('settings.commandsHelp') }}</p>

      <template v-for="command in commands">
        <h3>{{ capitalize(command.name) }}</h3>
        <textarea v-model.trim="command.value"></textarea>
      </template>

      <p><input type="submit" value="Save"></p>
    </form>

  </div>
</template>

<script>
import { mapState, mapMutations } from 'vuex'
import { getSettings, updateSettings } from '@/utils/api'

export default {
  name: 'settings',
  data: function () {
    return {
      commands: [],
      staticGen: []
    }
  },
  computed: {
    ...mapState([ 'user' ])
  },
  created () {
    getSettings()
      .then(settings => {
        if (this.$store.state.staticGen.length > 0) {
          this.parseStaticGen(settings.staticGen)
        }

        for (let key in settings.commands) {
          this.commands.push({
            name: key,
            value: settings.commands[key].join('\n')
          })
        }
      })
      .catch(error => { this.showError(error) })
  },
  methods: {
    ...mapMutations([ 'showSuccess', 'showError' ]),
    capitalize (name, where = '_') {
      if (where === 'caps') where = /(?=[A-Z])/
      let splitted = name.split(where)
      name = ''

      for (let i = 0; i < splitted.length; i++) {
        name += splitted[i].charAt(0).toUpperCase() + splitted[i].slice(1) + ' '
      }

      return name.slice(0, -1)
    },
    saveCommands (event) {
      event.preventDefault()

      let commands = {}

      for (let command of this.commands) {
        let value = command.value.split('\n')
        if (value.length === 1 && value[0] === '') {
          value = []
        }

        commands[command.name] = value
      }

      updateSettings(commands, 'commands')
        .then(() => { this.showSuccess(this.$t('settings.commandsUpdated')) })
        .catch(error => { this.showError(error) })
    },
    saveStaticGen (event) {
      event.preventDefault()
      let staticGen = {}

      for (let field of this.staticGen) {
        staticGen[field.variable] = field.value

        if (field.original === 'array') {
          let val = field.value.split(' ')
          if (val[0] === '') {
            val.shift()
          }

          staticGen[field.variable] = val
        }
      }

      updateSettings(staticGen, 'staticGen')
        .then(() => { this.showSuccess(this.$t('settings.settingsUpdated')) })
        .catch(error => { this.showError(error) })
    },
    parseStaticGen (staticgen) {
      for (let option of staticgen) {
        let value = option.value

        let field = {
          name: option.name,
          variable: option.variable,
          type: 'text',
          original: 'text',
          value: value
        }

        if (Array.isArray(value)) {
          field.original = 'array'
          field.value = value.join(' ')

          this.staticGen.push(field)
          continue
        }

        switch (typeof value) {
          case 'boolean':
            field.type = 'checkbox'
            field.original = 'boolean'
            break
        }

        this.staticGen.push(field)
      }
    }
  }
}
</script>
