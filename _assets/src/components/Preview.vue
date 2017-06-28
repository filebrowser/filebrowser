<template>
    <div id="previewer">
        <div class="bar">
            <button class="action" aria-label="Close Preview" id="close">
                <i class="material-icons">close</i>
            </button>
            <!-- TODO: add more buttons -->
        </div>

        <div class="preview">
            <img v-if="Data.Type == 'image'" :src="raw()">
            <audio v-else-if="Data.Type == 'audio'" :src="raw()" controls></audio>
            <video v-else-if="Data.Type == 'video'" :src="raw()" controls>
                Sorry, your browser doesn't support embedded videos,
                but don't worry, you can <a href="?download=true">download it</a>
                and watch it with your favorite video player!
            </video>
            <object v-else-if="Data.Extension == '.pdf'" class="pdf" :data="raw()"></object>
            <a v-else-if="Data.Type == 'blob'" href="?download=true"><h2 class="message">Download <i class="material-icons">file_download</i></h2></a>
            <pre v-else >{{ Data.Content }}</pre>
        </div>
    </div>
</template>

<script>
export default {
  name: 'preview',
  data: function () {
    return window.page
  },
  methods: {
    raw: function () {
      return this.Data.URL + '?raw=true'
    }
  }
}
</script>
