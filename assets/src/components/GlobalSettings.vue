<template>
  <div class="dashboard">
    <h1>Global Settings</h1>

    <ul>
      <li><router-link v-if="user.admin" to="/users">Go to User Management</router-link></li>
    </ul>

    <form @submit="saveHooks">
      <h2>Commands</h2>

      <p class="small">Here you can set commands that are executed in the named events. You write one command
        per line. If the event is related to files, such as before and after saving, the environment variable
        <code>file</code> will be available with the path of the file.</p>

      <h3>Before Save</h3>
      <textarea v-model="beforeSave"></textarea>

      <h3>After Save</h3>
      <textarea v-model="afterSave"></textarea>

      <p><input type="submit" value="Save"></p>
    </form>

  </div>
</template>

<script>
import { mapState, mapMutations } from 'vuex'

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
    // TODO: fetch current settings here
  },
  methods: {
    ...mapMutations([ 'showSuccess' ]),
    saveHooks (event) {
      event.preventDefault()
    }
  }
}
</script>
