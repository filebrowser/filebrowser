<template>
    <button @click="download" aria-label="Download" title="Download" class="action">
        <i class="material-icons">file_download</i>
        <span>Download</span>
        <span v-if="count() > 0" class="counter">{{ count() }}</span>
    </button>
</template>

<script>
var $ = window.info

export default {
  name: 'download-button',
  methods: {
    count: function () {
      return this.$store.getters.selectedCount
    },
    download: function (event) {
      if ($.req.kind !== 'listing') {
        window.location = window.location.pathname + '?download=true'
        return
      }

      /*
           <ul class="dropdown" id="download-drop">
            <a tabindex="0" aria-label="Download as Zip" data-format="zip" href="?download=zip"><li>zip</li></a>
            <a tabindex="0" aria-label="Download as Tar" data-format="tar" href="?download=tar"><li>tar</li></a>
            <a tabindex="0" aria-label="Download as TarGz" data-format="targz" href="?download=targz"><li>tar.gz</li></a>
            <a tabindex="0" aria-label="Download as TarBz2" data-format="tarbz2" href="?download=tarbz2"><li>tar.bz2</li></a>
            <a tabindex="0" aria-label="Download as TarXz" data-format="tarbz2" href="?download=tarxz"><li>tar.xz</li></a>
        </ul>
        */
      // document.getElementById('upload-input').click()
      // TODO
      alert('Not Implemented!')
    }
  }
}
</script>
