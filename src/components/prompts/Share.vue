<template>
  <div class="card floating" id="share">
    <div class="card-title">
      <h2>{{ $t('buttons.share') }}</h2>
    </div>

    <div class="card-content">
      <ul>
        <li v-if="!hasPermanent">
          <a @click="getPermalink" :aria-label="$t('buttons.permalink')">{{ $t('buttons.permalink') }}</a>
        </li>

        <li v-for="link in links" :key="link.hash">
          <a :href="buildLink(link.hash)" target="_blank">
            <template v-if="link.expires">{{ humanTime(link.expireDate) }}</template>
            <template v-else>{{ $t('permanent') }}</template>
          </a>

          <button class="action"
            @click="deleteLink($event, link)"
            :aria-label="$t('buttons.delete')"
            :title="$t('buttons.delete')"><i class="material-icons">delete</i></button>

          <button class="action copy-clipboard"
            :data-clipboard-text="buildLink(link.hash)"
            :aria-label="$t('buttons.copyToClipboard')"
            :title="$t('buttons.copyToClipboard')"><i class="material-icons">content_paste</i></button>
        </li>

        <li>
          <input v-focus
            type="number"
            max="2147483647"
            min="0"
            @keyup.enter="submit"
            v-model.trim="time">
          <select v-model="unit" :aria-label="$t('time.unit')">
            <option value="seconds">{{ $t('time.seconds') }}</option>
            <option value="minutes">{{ $t('time.minutes') }}</option>
            <option value="hours">{{ $t('time.hours') }}</option>
            <option value="days">{{ $t('time.days') }}</option>
          </select>
          <button class="action"
            @click="submit"
            :aria-label="$t('buttons.create')"
            :title="$t('buttons.create')"><i class="material-icons">add</i></button>
        </li>
      </ul>
    </div>

    <div class="card-action">
      <button class="flat"
        @click="$store.commit('closeHovers')"
        :aria-label="$t('buttons.close')"
        :title="$t('buttons.close')">{{ $t('buttons.close') }}</button>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex'
import { getShare, deleteShare, share } from '@/utils/api'
import moment from 'moment'
import Clipboard from 'clipboard'

export default {
  name: 'share',
  data: function () {
    return {
      time: '',
      unit: 'hours',
      hasPermanent: false,
      links: [],
      clip: null
    }
  },
  computed: {
    ...mapState([ 'baseURL', 'req', 'selected', 'selectedCount' ]),
    url () {
      // Get the current name of the file we are editing.
      if (this.req.kind !== 'listing') {
        return this.$route.path
      }

      if (this.selectedCount === 0 || this.selectedCount > 1) {
        // This shouldn't happen.
        return
      }

      return this.req.items[this.selected[0]].url
    }
  },
  beforeMount () {
    getShare(this.url)
      .then(links => {
        this.links = links
        this.sort()

        for (let link of this.links) {
          if (!link.expires) {
            this.hasPermanent = true
            break
          }
        }
      })
      .catch(error => {
        if (error === 404) return
        this.$showError(error)
      })
  },
  mounted () {
    this.clip = new Clipboard('.copy-clipboard')
    this.clip.on('success', (e) => {
      this.$showSuccess(this.$t('success.linkCopied'))
    })
  },
  beforeDestroy () {
    this.clip.destroy()
  },
  methods: {
    submit: function (event) {
      if (!this.time) return

      share(this.url, this.time, this.unit)
        .then(result => { this.links.push(result); this.sort() })
        .catch(this.$showError)
    },
    getPermalink (event) {
      share(this.url)
        .then(result => {
          this.links.push(result)
          this.sort()
          this.hasPermanent = true
        })
        .catch(this.$showError)
    },
    deleteLink (event, link) {
      event.preventDefault()
      deleteShare(link.hash)
        .then(() => {
          if (!link.expires) this.hasPermanent = false
          this.links = this.links.filter(item => item.hash !== link.hash)
        })
        .catch(this.$showError)
    },
    humanTime (time) {
      return moment(time).fromNow()
    },
    buildLink (hash) {
      return `${window.location.origin}${this.baseURL}/share/${hash}`
    },
    sort () {
      this.links = this.links.sort((a, b) => {
        if (!a.expires) return -1
        if (!b.expires) return 1
        return new Date(a.expireDate) - new Date(b.expireDate)
      })
    }
  }
}
</script>

