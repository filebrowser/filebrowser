<template>
  <button @click="action" :aria-label="label" :title="label" class="action">
    <i class="material-icons">{{ icon }}</i>
    <span>{{ label }}</span>
    <span v-if="counter > 0" class="counter">{{ counter }}</span>
  </button>
</template>

<script setup lang="ts">
import { useLayoutStore } from "@/stores/layout";

defineProps<{
  icon: any;
  label: any;
  counter: any;
  show: any;
}>();

const emit = defineEmits<{
  (e: "action"): void;
}>();

const layoutStore = useLayoutStore();

const action = () => {
  if (layoutStore.show) {
    // TODO: is not very pretty
    layoutStore.showHover(layoutStore.show as string);
  }

  emit("action");
};
</script>
