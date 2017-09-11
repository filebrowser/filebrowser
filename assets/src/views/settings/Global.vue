<template>
  <div class="dashboard">
    <form class="card" v-if="staticGen.length" @submit.prevent="saveStaticGen">
      <div class="card-title">
        <h2>{{ capitalize($store.state.staticGen) }}</h2>
      </div>

      <div class="card-content">
        <p v-for="field in staticGen" :key="field.variable">
          <label v-if="field.type !== 'checkbox'">{{ field.name }}</label>
          <input v-if="field.type === 'text'" type="text" v-model.trim="field.value">
          <input v-else-if="field.type === 'checkbox'" type="checkbox" v-model.trim="field.value">
          <template v-if="field.type === 'checkbox'">{{ capitalize(field.name, 'caps') }}</template>
        </p>
      </div>

      <div class="card-action">
        <input class="flat" type="submit" :value="$t('buttons.update')">
      </div>
    </form>

    <form class="card" @submit.prevent="saveCSS">
      <div class="card-title">
        <h2>{{ $t('settings.customStylesheet') }}</h2>
      </div>

      <div class="card-content">
        <textarea v-model="css"></textarea>
      </div>

      <div class="card-action">
        <input class="flat" type="submit" :value="$t('buttons.update')">
      </div>
    </form>

    <form class="card" @submit.prevent="saveCommands">
      <div class="card-title">
        <h2>{{ $t('settings.commands') }}</h2>
      </div>

      <div class="card-content">
        <p class="small">{{ $t('settings.commandsHelp') }}</p>

        <div v-for="command in commands" :key="command.name" class="collapsible">
          <input :id="command.name" type="checkbox">
          <label :for="command.name">
            <p>{{ capitalize(command.name) }}</p>
            <i class="material-icons">arrow_drop_down</i>
          </label>
          <div class="collapse">
            <textarea v-model.trim="command.value"></textarea>
          </div>
        </div>
      </div>

      <div class="card-action">
        <input class="flat" type="submit" :value="$t('buttons.update')">
      </div>
    </form>

  </div>
</template>

<script>
import { mapState } from 'vuex'
import { getSettings, updateSettings } from '@/utils/api'

export default {
  name: 'settings',
  data: function () {
    return {
      commands: [],
      staticGen: [],
      css: ''
    }
  },
  computed: {
    ...mapState([ 'user' ])
  },
  created () {
    getSettings()
      .then(settings => {
        if (this.$store.state.staticGen.length > 0) {
          this.parseStaticGen(settings.staticGen)
        }

        for (let key in settings.commands) {
          this.commands.push({
            name: key,
            value: settings.commands[key].join('\n')
          })
        }

        this.css = settings.css
      })
      .catch(this.$showError)
  },
  methods: {
    capitalize (name, where = '_') {
      if (where === 'caps') where = /(?=[A-Z])/
      let splitted = name.split(where)
      name = ''

      for (let i = 0; i < splitted.length; i++) {
        name += splitted[i].charAt(0).toUpperCase() + splitted[i].slice(1) + ' '
      }

      return name.slice(0, -1)
    },
    saveCommands (event) {
      let commands = {}

      for (let command of this.commands) {
        let value = command.value.split('\n')
        if (value.length === 1 && value[0] === '') {
          value = []
        }

        commands[command.name] = value
      }

      updateSettings(commands, 'commands')
        .then(() => { this.$showSuccess(this.$t('settings.commandsUpdated')) })
        .catch(this.$showError)
    },
    saveCSS (event) {
      updateSettings(this.css, 'css')
        .then(() => {
          this.$showSuccess(this.$t('settings.settingsUpdated'))
          this.$store.commit('setCSS', this.css)
          this.$emit('css')
        })
        .catch(this.$showError)
    },
    saveStaticGen (event) {
      let staticGen = {}

      for (let field of this.staticGen) {
        staticGen[field.variable] = field.value

        if (field.original === 'array') {
          let val = field.value.split(' ')
          if (val[0] === '') {
            val.shift()
          }

          staticGen[field.variable] = val
        }
      }

      updateSettings(staticGen, 'staticGen')
        .then(() => { this.$showSuccess(this.$t('settings.settingsUpdated')) })
        .catch(this.$showError)
    },
    parseStaticGen (staticgen) {
      for (let option of staticgen) {
        let value = option.value

        let field = {
          name: option.name,
          variable: option.variable,
          type: 'text',
          original: 'text',
          value: value
        }

        if (Array.isArray(value)) {
          field.original = 'array'
          field.value = value.join(' ')

          this.staticGen.push(field)
          continue
        }

        switch (typeof value) {
          case 'boolean':
            field.type = 'checkbox'
            field.original = 'boolean'
            break
        }

        this.staticGen.push(field)
      }
    }
  }
}
</script>
