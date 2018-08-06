<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t('prompts.schedule') }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t('prompts.scheduleMessage') }}</p>
      <input v-focus type="datetime-local" v-model="date">
    </div>

    <div class="card-action">
      <button class="cancel flat"
        @click="close"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')">{{ $t('buttons.cancel') }}</button>
      <button class="falt"
        @click="submit"
        :aria-label="$t('buttons.schedule')"
        :title="$t('buttons.schedule')">{{ $t('buttons.schedule') }}</button>
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

