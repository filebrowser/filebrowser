<template>
  <div class="dashboard">
    <h1>Global Settings</h1>

    <ul>
      <li><router-link to="/settings/profile">Go to Profile Settings</router-link></li>
      <li><router-link to="/users">Go to User Management</router-link></li>
    </ul>

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
      beforeSave: '',
      afterSave: ''
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
        console.log(plugins)
      })
      .catch(error => { this.showError(error) })
  },
  methods: {
    ...mapMutations([ 'showSuccess', 'showError' ]),
    capitalize (name) {
      let splitted = name.split('_')
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
    }
  }
}
</script>
