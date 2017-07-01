<template>
  <fieldset :id="name" :data-type="type">
    <h3 v-if="title !== ''">{{ name }}</h3>
    <div class="action add">
      <i class="material-icons" title="Add">add</i>
    </div>
    <div class="action delete" :data-delete="name">
      <i class="material-icons" title="Close">close</i>
    </div>
    <!-- template blocks w/ content -->
  </fieldset>
</template>

<script>
export default {
  name: 'array-object',
  props: ['name', 'type', 'title', 'content']
}
</script>
