<template>
  <div class="prompt">
    <h3>{{ $t('prompts.newFile') }}</h3>
    <p>{{ $t('prompts.newArchetype') }}</p>
    <input autofocus type="text" @keyup.enter="submit" v-model.trim="name">
    <input type="text" @keyup.enter="submit" v-model.trim="archetype">
    <div>
      <button class="ok"
        @click="submit"
        :aria-label="$t('buttons.create')"
        :title="$t('buttons.create')">{{ $t('buttons.create') }}</button>
      <button class="cancel"
        @click="$store.commit('closeHovers')"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')">{{ $t('buttons.cancel') }}</button>
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
        .catch(error => {
          this.$store.commit('showError', error)
        })
    },
    new (url, type) {
      url = removePrefix(url)

      return new Promise((resolve, reject) => {
        let request = new window.XMLHttpRequest()
        request.open('POST', `${this.$store.state.baseURL}/api/resource${url}`, true)
        request.setRequestHeader('Authorization', `Bearer ${this.$store.state.jwt}`)
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

