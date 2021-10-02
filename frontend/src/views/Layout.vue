<template>
  <div>
    <div id="progress">
      <div v-bind:style="{ width: this.progress + '%' }"></div>
    </div>
    <sidebar></sidebar>
    <main>
      <router-view></router-view>
      <shell v-if="isExecEnabled && isLogged && user.perm.execute" />
    </main>
    <prompts></prompts>

    <div v-if="this.progress !==0" >
       <div style=" height: 8rem;  width: 100%;  z-index: 9999;  position: absolute; top:40%; text-align:center; font-size:6em; "> {{ this.progress }}%</div>
    </div>

  </div>


  
</template>

<script>
import { mapState, mapGetters } from "vuex";
import Sidebar from "@/components/Sidebar";
import Prompts from "@/components/prompts/Prompts";
import Shell from "@/components/Shell";
import { enableExec } from "@/utils/constants";

export default {
  name: "layout",
  components: {
    Sidebar,
    Prompts,
    Shell,
  },
  computed: {
    ...mapGetters(["isLogged", "progress"]),
    ...mapState(["user"]),
    isExecEnabled: () => enableExec,
  },
  watch: {
    $route: function () {
      this.$store.commit("resetSelected");
      this.$store.commit("multiple", false);
      if (this.$store.state.show !== "success")
        this.$store.commit("closeHovers");
    },
  },
};
</script>
