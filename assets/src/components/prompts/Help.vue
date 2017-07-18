<template>
  <div class="prompt help">
    <h3>Help</h3>

    <ul>
      <li><strong>F1</strong> - this information</li>
      <li><strong>F2</strong> - rename file</li>
      <li><strong>DEL</strong> - delete selected items</li>
      <li><strong>ESC</strong> - clear selection and/or close the prompt</li>
      <li><strong>CTRL + S</strong> - save a file or download the directory where you are</li>
      <li><strong>CTRL + Click</strong> - select multiple files or directories</li>
      <li><strong>Double click</strong> - open a file or directory</li>
      <li><strong>Click</strong> - select file or directory</li>
    </ul>

    <p>Not available yet</p>

    <ul>
      <li><strong>Alt + Click</strong> - select a group of files</li>
    </ul>

    <div>
      <button type="submit" @click="$store.commit('closeHovers')" class="ok">OK</button>
    </div>
  </div>
</template>

<script>
export default {name: 'help'}
</script>

