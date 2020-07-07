<template>
  <div id="search" @click="open" v-bind:class="{ active , ongoing }">
    <div id="input">
      <button
        v-if="active"
        class="action"
        @click="close"
        :aria-label="$t('buttons.close')"
        :title="$t('buttons.close')"
      >
        <i class="material-icons">arrow_back</i>
      </button>
      <i v-else class="material-icons">search</i>
      <input
        type="text"
        @keyup.exact="keyup"
        @keyup.enter="submit"
        ref="input"
        :autofocus="active"
        v-model.trim="value"
        :aria-label="$t('search.search')"
        :placeholder="$t('search.search')"
      >
    </div>

    <div id="result" ref="result">
      <div>
        <template v-if="isEmpty">
          <p>{{ text }}</p>

          <template v-if="value.length === 0">
            <div class="boxes">
              <h3>{{ $t('search.types') }}</h3>
              <div>
                <div
                  tabindex="0"
                  v-for="(v,k) in boxes"
                  :key="k"
                  role="button"
                  @click="init('type:'+k)"
                  :aria-label="$t('search.'+v.label)"
                >
                  <i class="material-icons">{{v.icon}}</i>
                  <p>{{ $t('search.'+v.label) }}</p>
                </div>
              </div>
            </div>
          </template>
        </template>
        <ul v-show="results.length > 0">
          <li v-for="(s,k) in filteredResults" :key="k">
            <router-link @click.native="close" :to="'./' + s.path">
              <i v-if="s.dir" class="material-icons">folder</i>
              <i v-else class="material-icons">insert_drive_file</i>
              <span>./{{ s.path }}</span>
            </router-link>
          </li>
        </ul>
      </div>
      <p id="renew">
        <i class="material-icons spin">autorenew</i>
      </p>
    </div>
  </div>
</template>

<script>
import { mapState, mapGetters, mapMutations } from "vuex"
import url from "@/utils/url"
import { search } from "@/api"

var boxes = {
  image: { label: "images", icon: "insert_photo" },
  audio: { label: "music", icon: "volume_up" },
  video: { label: "video", icon: "movie" },
  pdf: { label: "pdf", icon: "picture_as_pdf" }
}

export default {
  name: "search",
  data: function() {
    return {
      value: "",
      active: false,
      ongoing: false,
      results: [],
      reload: false,
      resultsCount: 50,
      scrollable: null
    }
  },
  watch: {
    show (val, old) {
      this.active = val === "search"

      if (old === "search" && !this.active) {
        if (this.reload) {
          this.setReload(true)
        }

        document.body.style.overflow = "auto"
        this.reset()
        this.value = ''
        this.active = false
        this.$refs.input.blur()
      } else if (this.active) {
        this.reload = false
        this.$refs.input.focus()
        document.body.style.overflow = "hidden"
      }
    },
    value () {
      if (this.results.length) {
        this.reset()
      }
    }
  },
  computed: {
    ...mapState(["user", "show"]),
    ...mapGetters(["isListing"]),
    boxes() {
      return boxes
    },
    isEmpty() {
      return this.results.length === 0
    },
    text() {
      if (this.ongoing) {
        return ""
      }

      return this.value === '' ? this.$t("search.typeToSearch") : this.$t("search.pressToSearch")
    },
    filteredResults () {
      return this.results.slice(0, this.resultsCount)
    }
  },
  mounted() {
    this.$refs.result.addEventListener('scroll', event => {
      if (event.target.offsetHeight + event.target.scrollTop >= event.target.scrollHeight - 100) {
        this.resultsCount += 50
      }
    })
  },
  methods: {
    ...mapMutations(["showHover", "closeHovers", "setReload"]),
    open() {
      this.showHover("search")
    },
    close(event) {
      event.stopPropagation()
      event.preventDefault()
      this.closeHovers()
    },
    keyup(event) {
      if (event.keyCode === 27) {
        this.close(event)
        return
      }

      this.results.length = 0
    },
    init (string) {
      this.value = `${string} `
      this.$refs.input.focus()
    },
    reset () {
      this.ongoing = false
      this.resultsCount = 50
      this.results = []
    },
    async submit(event) {
      event.preventDefault()

      if (this.value === '') {
        return
      }

      let path = this.$route.path
      if (!this.isListing) {
        path = url.removeLastDir(path) + "/"
      }

      this.ongoing = true


      this.results = await search(path, this.value)
      this.ongoing = false
    }
  }
}
</script>
