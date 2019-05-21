<template>
  <div class="dashboard">
    <ul id="nav" v-if="user.perm.admin">
      <li :class="{ active: $route.path === '/settings/profile' }"><router-link to="/settings/profile">{{ $t('settings.profileSettings') }}</router-link></li>
      <li :class="{ active: $route.path === '/settings/global' }"><router-link to="/settings/global">{{ $t('settings.globalSettings') }}</router-link></li>
      <li :class="{ active: $route.path === '/settings/users' }"><router-link to="/settings/users">{{ $t('settings.userManagement') }}</router-link></li>
    </ul>

    <router-view></router-view>
  </div>
</template>

<script>
import { mapState } from 'vuex'

export default {
  name: 'settings',
  computed: mapState([ 'user' ])
}
</script>
