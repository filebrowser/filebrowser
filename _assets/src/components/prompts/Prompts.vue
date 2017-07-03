<template>
  <div>
    <help v-show="showHelp" :class="{ active: showHelp }"></help>
    <download v-if="showDownload" :class="{ active: showDownload }"></download>
    <new-file v-if="showNewFile" :class="{ active: showNewFile }"></new-file>
    <new-dir v-if="showNewDir" :class="{ active: showNewDir }"></new-dir>
    <rename v-if="showRename" :class="{ active: showRename }"></rename>
    <delete v-if="showDelete" :class="{ active: showDelete }"></delete>
    <info v-if="showInfo" :class="{ active: showInfo }"></info>
    <move v-if="showMove" :class="{ active: showMove }"></move>
    <div v-show="showOverlay" @click="resetPrompts" class="overlay" :class="{ active: showOverlay }"></div>
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
import {mapGetters} from 'vuex'

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
