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
    <copy v-else-if="showCopy"></copy>
    <replace v-else-if="showReplace"></replace>
    <schedule v-else-if="show === 'schedule'"></schedule>
    <new-archetype v-else-if="show === 'new-archetype'"></new-archetype>
    <share v-else-if="show === 'share'"></share>
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
import NewArchetype from './NewArchetype'
import Replace from './Replace'
import Schedule from './Schedule'
import Share from './Share'
import { mapState } from 'vuex'
import buttons from '@/utils/buttons'
import * as api from '@/utils/api'

export default {
  name: 'prompts',
  components: {
    Info,
    Delete,
    NewArchetype,
    Schedule,
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
        api,
        buttons,
        'store': this.$store,
        'router': this.$router
      }
    }
  },
  computed: {
    ...mapState(['show', 'plugins']),
    showInfo: function () { return this.show === 'info' },
    showHelp: function () { return this.show === 'help' },
    showDelete: function () { return this.show === 'delete' },
    showRename: function () { return this.show === 'rename' },
    showMove: function () { return this.show === 'move' },
    showCopy: function () { return this.show === 'copy' },
    showNewFile: function () { return this.show === 'newFile' },
    showNewDir: function () { return this.show === 'newDir' },
    showDownload: function () { return this.show === 'download' },
    showReplace: function () { return this.show === 'replace' },
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
