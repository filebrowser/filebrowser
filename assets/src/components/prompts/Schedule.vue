<template>
  <div class="prompt">
    <h3>{{ $t('prompts.schedule') }}</h3>
    <p>{{ $t('prompts.scheduleMessage') }}</p>
    <input autofocus type="datetime-local" v-model="date">
    <div>
      <button class="ok"
        @click="submit"
        :aria-label="$t('buttons.schedule')"
        :title="$t('buttons.schedule')">{{ $t('buttons.schedule') }}</button>
      <button class="cancel"
        @click="close"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')">{{ $t('buttons.cancel') }}</button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'schedule',
  data: function () {
    return {
      date: ''
    }
  },
  methods: {
    close () {
      this.$store.commit('closeHovers')
    },
    submit: function (event) {
      event.preventDefault()
      if (this.date === '') return
      this.close()
      this.$store.commit('setSchedule', this.date)
      document.getElementById('save-button').click()
    }
  }
}
</script>

