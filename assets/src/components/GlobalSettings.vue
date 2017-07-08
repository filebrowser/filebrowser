<template>
  <div class="dashboard">
    <h1>Global Settings</h1>

    <ul>
      <li><router-link v-if="user.admin" to="/users">Go to User Management</router-link></li>
    </ul>

    <form>
      <h2>Before Save</h2>
      <textarea v-model="css" name="css"></textarea>
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
    }
  },
  computed: {
    ...mapState([ 'user' ])
  },
  methods: {
    ...mapMutations([ 'showSuccess' ])
  }
}
</script>
