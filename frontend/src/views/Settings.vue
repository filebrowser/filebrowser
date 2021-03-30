<template>
  <div class="dashboard">
    <header-bar showMenu showLogo />

    <div id="nav">
      <div class="wrapper">
        <ul>
          <router-link to="/settings/profile"
            ><li :class="{ active: $route.path === '/settings/profile' }">
              {{ $t("settings.profileSettings") }}
            </li></router-link
          >
          <router-link to="/settings/shares" v-if="user.perm.share"
            ><li :class="{ active: $route.path === '/settings/shares' }">
              {{ $t("settings.shareManagement") }}
            </li></router-link
          >
          <router-link to="/settings/global" v-if="user.perm.admin"
            ><li :class="{ active: $route.path === '/settings/global' }">
              {{ $t("settings.globalSettings") }}
            </li></router-link
          >
          <router-link to="/settings/users" v-if="user.perm.admin"
            ><li
              :class="{
                active:
                  $route.path === '/settings/users' || $route.name === 'User',
              }"
            >
              {{ $t("settings.userManagement") }}
            </li></router-link
          >
        </ul>
      </div>
    </div>

    <router-view></router-view>
  </div>
</template>

<script>
import { mapState } from "vuex";

import HeaderBar from "@/components/header/HeaderBar";

export default {
  name: "settings",
  components: {
    HeaderBar,
  },
  computed: {
    ...mapState(["user"]),
  },
};
</script>
