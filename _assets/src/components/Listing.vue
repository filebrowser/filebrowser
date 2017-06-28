<template>
    <div id="listing" :class="Data.Display">
        <div>
            <div class="item header">
                <div></div>
                <div>
                    <p v-bind:class="{ active: Data.Sort === 'name' }" class="name"><span>Name</span>
                        <a v-if="Data.Sort === 'name' && Data.Order != 'asc'" href="?sort=name&order=asc"><i class="material-icons">arrow_upward</i></a>
                        <a v-else href="?sort=name&order=desc"><i class="material-icons">arrow_downward</i></a>
                    </p>

                    <p v-bind:class="{ active: Data.Sort === 'size' }" class="size"><span>Size</span>
                        <a v-if="Data.Sort === 'size' && Data.Order != 'asc'" href="?sort=size&order=asc"><i class="material-icons">arrow_upward</i></a>
                        <a v-else href="?sort=size&order=desc"><i class="material-icons">arrow_downward</i></a>
                    </p>

                    <p class="modified">Last modified</p>
                </div>
            </div>
        </div>

        <h2 v-if="(Data.NumDirs + Data.NumFiles) == 0" class="message">It feels lonely here :'(</h2>

        <h2 v-if="Data.NumDirs !== 0">Folders</h2>
        <div v-if="Data.NumDirs !== 0">
          <item
            v-for="(item, index) in Data.Items"
            v-if="item.IsDir"
            :key="base64(item.Name)"
            :id="base64(item.Name)"
            v-bind:name="item.Name"
            v-bind:isDir="item.IsDir"
            v-bind:url="item.URL"
            v-bind:modified="item.ModTime"  
            v-bind:type="item.Type"
            v-bind:size="item.Size">
          </item>
        </div>

        <h2 v-if="Data.NumItems !== 0">Files</h2>
        <div v-if="Data.NumItems !== 0">
          <item
            v-for="(item, index) in Data.Items"
            v-if="!item.IsDir"
            :key="base64(item.Name)"
            :id="base64(item.Name)"
            v-bind:name="item.Name"
            v-bind:isDir="item.IsDir"
            v-bind:modified="item.ModTime"  
            v-bind:url="item.URL"
            v-bind:type="item.Type"
            v-bind:size="item.Size">
          </item>
        </div>

        <!--
          <input style="display:none" type="file" id="upload-input" onchange="listing.handleFiles(this.files, '')" value="Upload" multiple>
          -->
    </div>
</template>

<script>
import Item from './ListingItem'

export default {
  name: 'preview',
  components: { Item },
  data: function () {
    return window.page
  },
  methods: {
    base64: function (name) {
      return window.btoa(name)
    }
  }
}
</script>
