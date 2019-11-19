<template>
  <div>
    <component :is="currentComponent"></component>
    <div v-show="showOverlay" @click="resetPrompts" class="overlay"></div>
  </div>
</template>

<script>
import Help from './Help'
import Info from './Info'
import Delete from './Delete'
import Rename from './Rename'
import Download from './Download'
import Move from './Move'
import Copy from './Copy'
import NewFile from './NewFile'
import NewDir from './NewDir'
import Replace from './Replace'
import Share from './Share'
import { mapState } from 'vuex'
import buttons from '@/utils/buttons'

export default {
  name: 'prompts',
  components: {
    Info,
    Delete,
    Rename,
    Download,
    Move,
    Copy,
    Share,
    NewFile,
    NewDir,
    Help,
    Replace
  },
  data: function () {
    return {
      pluginData: {
        buttons,
        'store': this.$store,
        'router': this.$router
      }
    }
  },
  computed: {
    ...mapState(['show', 'plugins']),
    currentComponent: function () {
      const matched = [
        'info',
        'help',
        'delete',
        'rename',
        'move',
        'copy',
        'newFile',
        'newDir',
        'download',
        'replace',
        'share'
      ].indexOf(this.show) >= 0;

      return matched && this.show || null;
    },
    showOverlay: function () {
      return (this.show !== null && this.show !== 'search' && this.show !== 'more')
    }
  },
  methods: {
    resetPrompts () {
      this.$store.commit('closeHovers')
    }
  }
}
</script>
