<template>
<div class="vue-bpmn-diagram-container">
    <br />Click and hold the mouse to move the diagram. Use CTRL+Scroll to zoom.<br />
    <Bpmn />
</div>
</template>

<script>
  import Bpmn from '@/components/files/Bpmn';

  export default {
    components: {
      Bpmn
    },
    methods: {

    }
  };
</script>


