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

    <form @submit="savePlugin" v-if="plugins.length > 0">
      <template v-for="plugin in plugins">
        <h2>{{ capitalize(plugin.name) }}</h2>

        <p v-for="field in plugin.fields" :key="field.variable">
          <label v-if="field.type !== 'checkbox'">{{ field.name }}</label>
          <input v-if="field.type === 'text'" type="text" v-model.trim="field.value">
          <input v-else-if="field.type === 'checkbox'" type="checkbox" v-model.trim="field.value">
          <template v-if="field.type === 'checkbox'">{{ capitalize(field.name, 'caps') }}</template>
        </p>
      </template>

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
      plugins: []
    }
  },
  computed: {
    ...mapState([ 'user' ])
  },
  created () {
    getSettings()
      .then(settings => {
        for (let key in settings.plugins) {
          this.plugins.push(this.parsePlugin(key, settings.plugins[key]))
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
    savePlugin (event) {
      event.preventDefault()
      let plugins = {}

      for (let plugin of this.plugins) {
        let p = {}

        for (let field of plugin.fields) {
          p[field.variable] = field.value

          if (field.original === 'array') {
            let val = field.value.split(' ')
            if (val[0] === '') {
              val.shift()
            }

            p[field.variable] = val
          }
        }

        plugins[plugin.name] = p
      }

      updateSettings(plugins, 'plugins')
        .then(() => { this.showSuccess(this.$t('settings.pluginsUpdated')) })
        .catch(error => { this.showError(error) })
    },
    parsePlugin (name, plugin) {
      let obj = {
        name: name,
        fields: []
      }

      for (let option of plugin) {
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

          obj.fields.push(field)
          continue
        }

        switch (typeof value) {
          case 'boolean':
            field.type = 'checkbox'
            field.original = 'boolean'
            break
        }

        obj.fields.push(field)
      }

      return obj
    }
  }
}
</script>
