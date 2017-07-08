<template>
  <div class="dashboard">
    <h1>Global Settings</h1>

    <ul>
      <li><router-link v-if="user.admin" to="/users">Go to User Management</router-link></li>
    </ul>

    <form @submit="saveCommands">
      <h2>Commands</h2>

      <p class="small">Here you can set commands that are executed in the named events. You write one command
        per line. If the event is related to files, such as before and after saving, the environment variable
        <code>file</code> will be available with the path of the file.</p>

      <h3>Before Save</h3>
      <textarea v-model.trim="beforeSave"></textarea>

      <h3>After Save</h3>
      <textarea v-model.trim="afterSave"></textarea>

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
        this.beforeSave = commands['before_save'].join('\n')
        this.afterSave = commands['after_save'].join('\n')
      })
      .catch(error => { this.showError(error) })
  },
  methods: {
    ...mapMutations([ 'showSuccess', 'showError' ]),
    saveCommands (event) {
      event.preventDefault()

      let commands = {
        'before_save': this.beforeSave.split('\n'),
        'after_save': this.afterSave.split('\n')
      }

      if (commands['before_save'].length === 1 && commands['before_save'][0] === '') commands['before_save'] = []
      if (commands['after_save'].length === 1 && commands['after_save'][0] === '') commands['after_save'] = []

      api.updateCommands(commands)
        .then(() => { this.showSuccess('Commands updated!') })
        .catch(error => { this.showError(error) })
    }
  }
}
</script>
