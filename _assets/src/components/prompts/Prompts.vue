<template>
  <div>
    <help v-if="showHelp" ></help>
    <download v-else-if="showDownload"></download>
    <new-file v-else-if="showNewFile"></new-file>
    <new-dir v-else-if="showNewDir"></new-dir>
    <rename v-else-if="showRename"></rename>
    <delete v-else-if="showDelete"></delete>
    <info v-else-if="showInfo"></info>
    <move v-else-if="showMove"></move>

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
import NewFile from './NewFile'
import NewDir from './NewDir'
import {mapGetters, mapState} from 'vuex'

export default {
  name: 'prompts',
  components: {
    Info,
    Delete,
    Rename,
    Download,
    Move,
    NewFile,
    NewDir,
    Help
  },
  computed: {
    ...mapState(['prompt']),
    ...mapGetters([
      'showOverlay',
      'showInfo',
      'showHelp',
      'showDelete',
      'showRename',
      'showMove',
      'showNewFile',
      'showNewDir',
      'showDownload'
    ])
  },
  methods: {
    resetPrompts () {
      this.$store.commit('closePrompts')
    }
  }
}
</script>
