<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t('prompts.newFile') }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t('prompts.newArchetype') }}</p>
      <input v-focus type="text" @keyup.enter="submit" v-model.trim="name">
      <input type="text" @keyup.enter="submit" v-model.trim="archetype">
    </div>

    <div class="card-action">
      <button class="flat cancel"
        @click="$store.commit('closeHovers')"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')">{{ $t('buttons.cancel') }}</button>
      <button class="flat"
        @click="submit"
        :aria-label="$t('buttons.create')"
        :title="$t('buttons.create')">{{ $t('buttons.create') }}</button>
    </div>
  </div>
</template>

<script>
import { removePrefix } from '@/utils/api'

export default {
  name: 'new-archetype',
  data: function () {
    return {
      name: '',
      archetype: 'default'
    }
  },
  methods: {
    submit: function (event) {
      event.preventDefault()
      this.$store.commit('closeHovers')

      this.new('/' + this.name, this.archetype)
        .then((url) => {
          this.$router.push({ path: url })
        })
        .catch(this.$showError)
    },
    new (url, type) {
      url = removePrefix(url)

      if (!url.endsWith('.md') && !url.endsWith('.markdown')) {
        url += '.markdown'
      }

      return new Promise((resolve, reject) => {
        let request = new window.XMLHttpRequest()
        request.open('POST', `${this.$store.state.baseURL}/api/resource${url}`, true)
        if (!this.$store.state.noAuth) request.setRequestHeader('Authorization', `Bearer ${this.$store.state.jwt}`)
        request.setRequestHeader('Archetype', encodeURIComponent(type))

        request.onload = () => {
          if (request.status === 200) {
            resolve(request.getResponseHeader('Location'))
          } else {
            reject(request.responseText)
          }
        }

        request.onerror = (error) => reject(error)
        request.send()
      })
    }
  }
}
</script>

