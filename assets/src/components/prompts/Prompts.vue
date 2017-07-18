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
    <error v-else-if="showError"></error>
    <success v-else-if="showSuccess"></success>

    <template v-for="plugin in plugins">
      <form class="prompt"
        v-for="prompt in plugin.prompts"
        :key="prompt.name"
        v-if="show === prompt.name"
        @submit="prompt.submit($event, pluginData, $route)">
        <h3>{{ prompt.title }}</h3>
        <p>{{ prompt.description }}</p>
        <input v-for="input in prompt.inputs"
          :key="input.name"
          :type="input.type"
          :name="input.name"
          :placeholder="input.placeholder">
        <div>
          <input type="submit" class="ok" :value="prompt.ok">
          <button class="cancel" @click.prevent="$store.commit('closeHovers')">Cancel</button>
        </div>
      </form>
    </template>

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
import Error from './Error'
import Success from './Success'
import NewFile from './NewFile'
import NewDir from './NewDir'
import { mapState } from 'vuex'
import buttons from '@/utils/buttons'
import api from '@/utils/api'

export default {
  name: 'prompts',
  components: {
    Info,
    Delete,
    Rename,
    Error,
    Download,
    Success,
    Move,
    NewFile,
    NewDir,
    Help
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
    showError: function () { return this.show === 'error' },
    showSuccess: function () { return this.show === 'success' },
    showInfo: function () { return this.show === 'info' },
    showHelp: function () { return this.show === 'help' },
    showDelete: function () { return this.show === 'delete' },
    showRename: function () { return this.show === 'rename' },
    showMove: function () { return this.show === 'move' },
    showNewFile: function () { return this.show === 'newFile' },
    showNewDir: function () { return this.show === 'newDir' },
    showDownload: function () { return this.show === 'download' },
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
