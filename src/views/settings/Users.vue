<template>
  <div class="card">
    <div class="card-title">
      <h2>{{ $t('settings.users') }}</h2>
      <router-link to="/settings/users/new"><button class="flat">{{ $t('buttons.new') }}</button></router-link>
    </div>

    <div class="card-content full">
      <table>
        <tr>
          <th>{{ $t('settings.username') }}</th>
          <th>{{ $t('settings.admin') }}</th>
          <th>{{ $t('settings.scope') }}</th>
          <th></th>
        </tr>

        <tr v-for="user in users" :key="user.id">
          <td>{{ user.username }}</td>
          <td><i v-if="user.admin" class="material-icons">done</i><i v-else class="material-icons">close</i></td>
          <td>{{ user.filesystem }}</td>
          <td class="small">
            <router-link :to="'/settings/users/' + user.ID"><i class="material-icons">mode_edit</i></router-link>
          </td>
        </tr>
      </table>
    </div>
  </div>
</template>

<script>
import * as api from '@/utils/api'

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
      this.$showError(error)
    })
  }
}
</script>
