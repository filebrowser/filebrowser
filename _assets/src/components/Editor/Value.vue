<template>
  <textarea v-if="htmlType === 'textarea'"
    class="scroll"
    :name="name"
    :id="name"
    :data-parent-type="parentType">
    {{ content.other }}
  </textarea>
  <input v-else-if="htmlType ==='datatime'"
    :name="name"
    :id="name"
    :value="content.other"
    type="datetime-local"
    :data-parent-type="parentType"></input>
  <input v-else
    :name="name"
    :id="name"
    :value="content.other"
    :type="htmlType"
    :data-parent-type="parentType"></input>
</template>

<script>
export default {
  name: 'value',
  props: ['htmlType', 'name', 'parentType', 'content']
}
</script>
