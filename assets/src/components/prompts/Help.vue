<template>
  <div class="prompt help">
    <h3>{{ $t('help.help') }}</h3>

    <ul>
      <li><strong>F1</strong> - {{ $t('help.f1') }}</li>
      <li><strong>F2</strong> - {{ $t('help.f2') }}</li>
      <li><strong>DEL</strong> - {{ $t('help.del') }}</li>
      <li><strong>ESC</strong> - {{ $t('help.esc') }}</li>
      <li><strong>CTRL + S</strong> - {{ $t('help.ctrl.s') }}</li>
      <li><strong>CTRL + Click</strong> - {{ $t('help.ctrl.click') }}</li>
      <li><strong>Click</strong> - {{ $t('help.click') }}</li>
      <li><strong>Double click</strong> - {{ $t('help.doubleClick') }}</li>
    </ul>

    <div>
      <button type="submit" @click="$store.commit('closeHovers')" class="ok">{{ $t('buttons.ok') }}</button>
    </div>
  </div>
</template>

<script>
export default {name: 'help'}
</script>

