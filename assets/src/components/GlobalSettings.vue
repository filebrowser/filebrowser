<template>
  <div class="dashboard">
    <h1>Global Settings</h1>

    <ul>
      <li><router-link to="/settings/profile">Go to Profile Settings</router-link></li>
      <li><router-link to="/users">Go to User Management</router-link></li>
    </ul>

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
      <h2>Commands</h2>

      <p class="small">Here you can set commands that are executed in the named events. You write one command
        per line. If the event is related to files, such as before and after saving, the environment variable
        <code>file</code> will be available with the path of the file.</p>

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
import api from '@/utils/api'

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
    api.getCommands()
      .then(commands => {
        for (let key in commands) {
          this.commands.push({
            name: key,
            value: commands[key].join('\n')
          })
        }
      })
      .catch(error => { this.showError(error) })

    api.getPlugins()
      .then(plugins => {
        for (let key in plugins) {
          this.plugins.push(this.parsePlugin(key, plugins[key]))
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

      api.updateCommands(commands)
        .then(() => { this.showSuccess('Commands updated!') })
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

      console.log(plugins)

      api.updatePlugins(plugins)
        .then(() => { this.showSuccess('Plugins settings updated!') })
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
