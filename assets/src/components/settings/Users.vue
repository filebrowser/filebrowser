<template>
  <div class="dashboard">
    <h1>Users <router-link to="/users/new"><button>New</button></router-link></h1>

    <table>
      <tr>
        <th>Username</th>
        <th>Admin</th>
        <th>Scope</th>
        <th></th>
      </tr>

      <tr v-for="user in users">
        <td>{{ user.username }}</td>
        <td><i v-if="user.admin" class="material-icons">done</i><i v-else class="material-icons">close</i></td>
        <td>{{ user.filesystem }}</td>
        <td><router-link :to="'/settigns/users/' + user.ID"><i class="material-icons">mode_edit</i></router-link></td>
      </tr>

    </table>
  </div>
</template>

<script>
import api from '@/utils/api'

export default {
  name: 'users',
  data: function () {
    return {
      users: []
    }
  },
  created () {
    api.getUsers().then(users => {
      this.users = users
    }).catch(error => {
      this.$store.commit('showError', error)
    })
  }
}
</script>
