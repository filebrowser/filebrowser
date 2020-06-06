<template>
  <div class="card floating">
    <div class="card-title">
      <h2>{{ $t('prompts.replace') }}</h2>
    </div>

    <div class="card-content">
      <p>{{ $t('prompts.replaceMessage') }}</p>
    </div>

    <div class="card-action">
      <button
        class="button button--flat button--grey"
        :aria-label="$t('buttons.cancel')"
        :title="$t('buttons.cancel')"
        @click="$store.commit('closeHovers')"
      >{{ $t('buttons.cancel') }}</button>
      <button
        class="button button--flat button--red"
        :aria-label="$t('buttons.replace')"
        :title="$t('buttons.replace')"
        @click="showConfirm"
      >{{ $t('buttons.replace') }}</button>
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex'

export default {
  name: 'Replace',
  computed: mapState(['showConfirm'])
}
</script>
