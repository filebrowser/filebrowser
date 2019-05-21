<template>
  <div class="card">
    <div class="card-title">
      <h2>{{ $t('settings.users') }}</h2>
      <router-link to="/settings/users/new"><button class="button">{{ $t('buttons.new') }}</button></router-link>
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
          <td><i v-if="user.perm.admin" class="material-icons">done</i><i v-else class="material-icons">close</i></td>
          <td>{{ user.scope }}</td>
          <td class="small">
            <router-link :to="'/settings/users/' + user.id"><i class="material-icons">mode_edit</i></router-link>
          </td>
        </tr>
      </table>
    </div>
  </div>
</template>

<script>
import { users as api } from '@/api'

export default {
  name: 'users',
  data: function () {
    return {
      users: []
    }
  },
  async created () {
    try {
      this.users = await api.getAll()
    } catch (e) {
      this.$showError(e)
    }
  }
}
</script>
