<template>
  <div class="prompt">
    <h3>{{ $t('prompts.replace') }}</h3>
    <p>{{ $t('prompts.replaceMessage') }}</p>

    <div>
      <button class="ok"
        @click="showConfirm"
        :aria-label="$t('buttons.replace')"
        :title="$t('buttons.replace')">{{ $t('buttons.replace') }}</button>
      <button class="cancel"
        @click="$store.commit('closeHovers')"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')">{{ $t('buttons.cancel') }}</button>
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
