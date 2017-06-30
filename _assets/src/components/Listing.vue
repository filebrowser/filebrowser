<template>
  <div id="listing"
    :class="req.data.display"
    @drop="drop"
    @dragenter="dragEnter"
    @dragend="dragEnd">
    <div>
      <div class="item header">
        <div></div>
        <div>
          <p v-bind:class="{ active: req.data.sort === 'name' }" class="name"><span>Name</span>
            <a v-if="req.data.sort === 'name' && req.data.order != 'asc'" href="?sort=name&order=asc"><i class="material-icons">arrow_upward</i></a>
            <a v-else href="?sort=name&order=desc"><i class="material-icons">arrow_downward</i></a>
          </p>

          <p v-bind:class="{ active: req.data.sort === 'size' }" class="size"><span>Size</span>
            <a v-if="req.data.sort === 'size' && req.data.order != 'asc'" href="?sort=size&order=asc"><i class="material-icons">arrow_upward</i></a>
            <a v-else href="?sort=size&order=desc"><i class="material-icons">arrow_downward</i></a>
          </p>

          <p class="modified">Last modified</p>
        </div>
      </div>
    </div>

    <h2 v-if="(req.data.numDirs + req.data.numFiles) == 0" class="message">It feels lonely here :'(</h2>

    <h2 v-if="req.data.numDirs > 0">Folders</h2>
    <div v-if="req.data.numDirs > 0">
      <item v-for="(item, index) in req.data.items"
        v-if="item.isDir"
        :key="base64(item.name)"
        v-bind:index="index"
        v-bind:name="item.name"
        v-bind:isDir="item.isDir"
        v-bind:url="item.url"
        v-bind:modified="item.modified"
        v-bind:type="item.type"
        v-bind:size="item.size">
      </item>
    </div>

    <h2 v-if="req.data.numFiles > 0">Files</h2>
    <div v-if="req.data.numFiles > 0">
      <item v-for="(item, index) in req.data.items"
        v-if="!item.isDir"
        :key="base64(item.name)"
        v-bind:index="index"
        v-bind:name="item.name"
        v-bind:isDir="item.isDir"
        v-bind:url="item.url"
        v-bind:modified="item.modified"
        v-bind:type="item.type"
        v-bind:size="item.size">
      </item>
    </div>

    <input style="display:none" type="file" id="upload-input" @change="uploadInput($event)" value="Upload" multiple>

    <div v-show="$store.state.multiple" :class="{ active: $store.state.multiple }" id="multiple-selection">
    <p>Multiple selection enabled</p>
      <div @click="$store.commit('multiple', false)" tabindex="0" role="button" title="Clear" aria-label="Clear" class="action">
        <i class="material-icons" title="Clear">clear</i>
      </div>
    </div>
  </div>
</template>

<script>
import {mapState} from 'vuex'
import Item from './ListingItem'
import webdav from '../utils/webdav'
import page from '../utils/page'

export default {
  name: 'listing',
  components: { Item },
  computed: mapState(['req']),
  mounted: function () {
    document.addEventListener('dragover', function (event) {
      event.preventDefault()
    }, false)

    document.addEventListener('drop', this.drop, false)
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
      this.resetOpacity()
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
        this.resetOpacity()
      }
    },
    uploadInput: function (event) {
      this.handleFiles(event.currentTarget.files, '')
    },
    resetOpacity: function () {
      let items = document.getElementsByClassName('item')

      Array.from(items).forEach(file => {
        file.style.opacity = 1
      })
    },
    handleFiles: function (files, base) {
      this.resetOpacity()

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
