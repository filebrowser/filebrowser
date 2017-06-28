<template>
    <div id="listing" 
      :class="data.display" 
      @drop="drop"
      @dragenter="dragEnter" 
      @dragend="dragEnd">
        <div>
            <div class="item header">
                <div></div>
                <div>
                    <p v-bind:class="{ active: data.sort === 'name' }" class="name"><span>Name</span>
                        <a v-if="data.sort === 'name' && data.order != 'asc'" href="?sort=name&order=asc"><i class="material-icons">arrow_upward</i></a>
                        <a v-else href="?sort=name&order=desc"><i class="material-icons">arrow_downward</i></a>
                    </p>

                    <p v-bind:class="{ active: data.sort === 'size' }" class="size"><span>Size</span>
                        <a v-if="data.sort === 'size' && data.order != 'asc'" href="?sort=size&order=asc"><i class="material-icons">arrow_upward</i></a>
                        <a v-else href="?sort=size&order=desc"><i class="material-icons">arrow_downward</i></a>
                    </p>

                    <p class="modified">Last modified</p>
                </div>
            </div>
        </div>

        <h2 v-if="(data.numDirs + data.numFiles) == 0" class="message">It feels lonely here :'(</h2>

        <h2 v-if="data.numDirs > 0">Folders</h2>
        <div v-if="data.numDirs > 0">
          <item
            v-for="(item, index) in data.items"
            v-if="item.isDir"
            :key="base64(item.name)"
            :id="base64(item.name)"
            v-bind:selected="selected"
            v-bind:name="item.name"
            v-bind:isDir="item.isDir"
            v-bind:url="item.url"
            v-bind:modified="item.modified"  
            v-bind:type="item.type"
            v-bind:size="item.size">
          </item>
        </div>

        <h2 v-if="data.numFiles > 0">Files</h2>
        <div v-if="data.numFiles > 0">
          <item
            v-for="(item, index) in data.items"
            v-if="!item.isDir"
            :key="base64(item.name)"
            :id="base64(item.name)"
            v-bind:selected="selected"
            v-bind:name="item.name"
            v-bind:isDir="item.isDir"
            v-bind:url="item.url"
            v-bind:modified="item.modified"  
            v-bind:type="item.type"
            v-bind:size="item.size">
          </item>
        </div>

        <!--
          <input style="display:none" type="file" id="upload-input" onchange="listing.handleFiles(this.files, '')" value="Upload" multiple>
          -->
    </div>
</template>

<script>
import Item from './ListingItem'
import webdav from '../webdav.js'
import page from '../page.js'

export default {
  name: 'preview',
  data: function () {
    return {
      data: window.info.page.data,
      selected: [],
      multiple: false
    }
  },
  components: { Item },
  mounted: function () {
    document.addEventListener('dragover', function (event) {
      event.preventDefault()
    }, false)

    document.addEventListener('drop', this.drop, false)
  },
  beforeUpdate: function () {
    /*
      listing.redefineDownloadURLs()

  let selectedNumber = selectedItems.length,
    fileAction = document.getElementById('file-only')

  if (selectedNumber) {
    fileAction.classList.remove('disabled')

    if (selectedNumber > 1) {
      buttons.rename.classList.add('disabled')
      buttons.info.classList.add('disabled')
    }

    if (selectedNumber == 1) {
      buttons.info.classList.remove('disabled')
      buttons.rename.classList.remove('disabled')
    }

    return false
  }

  buttons.info.classList.remove('disabled')
  fileAction.classList.add('disabled')
  */
    console.log('before upding')
  },
  methods: {
    base64: function (name) {
      return window.btoa(name)
    },
    dragEnter: function (event) {
      let items = document.getElementsByClassName('item')

      Array.from(items).forEach(file => {
        file.style.opacity = 0.5
      })
    },
    dragEnd: function (event) {
      let items = document.getElementsByClassName('item')

      Array.from(items).forEach(file => {
        file.style.opacity = 1
      })
    },
    drop: function (event) {
      event.preventDefault()

      let dt = event.dataTransfer
      let files = dt.files
      let el = event.target

      for (let i = 0; i < 5; i++) {
        if (el !== null && !el.classList.contains('item')) {
          el = el.parentElement
        }
      }

      if (files.length > 0) {
        if (el !== null && el.classList.contains('item') && el.dataset.dir === 'true') {
          this.handleFiles(files, el.querySelector('.name').innerHTML + '/')
          return
        }

        this.handleFiles(files, '')
      } else {
        let items = document.getElementsByClassName('item')

        Array.from(items).forEach(file => {
          file.style.opacity = 1
        })
      }
    },
    handleFiles: function (files, base) {
      // buttons.setLoading('upload')
      let promises = []

      for (let file of files) {
        promises.push(webdav.put(window.location.pathname + base + file.name, file))
      }

      Promise.all(promises)
      .then(() => {
        page.reload()
        // buttons.setDone('upload')
      })
      .catch(e => {
        console.log(e)
        // buttons.setDone('upload', false)
      })

      return false
    }
  }
}
</script>
