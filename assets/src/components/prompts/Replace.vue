<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t('prompts.replace') }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t('prompts.replaceMessage') }}</p>
    </div>

    <div class="card-action">
      <button class="flat cancel"
        @click="$store.commit('closeHovers')"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')">{{ $t('buttons.cancel') }}</button>
      <button class="flat"
        @click="showConfirm"
        :aria-label="$t('buttons.replace')"
        :title="$t('buttons.replace')">{{ $t('buttons.replace') }}</button>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex'

export default {
  name: 'replace',
  computed: mapState(['showConfirm'])
}
</script>
