<template>
  <div
    class="context-menu"
    ref="contextMenu"
    v-show="show"
    :style="{
      top: `${props.pos.y}px`,
      left: `${left}px`,
    }"
    @click="hideContextMenu"
    @contextmenu.prevent.stop
  >
    <slot />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed, onBeforeUnmount, nextTick } from "vue";

const emit = defineEmits(["hide"]);
const props = defineProps<{ show: boolean; pos: { x: number; y: number } }>();
const contextMenu = ref<HTMLElement | null>(null);

const left = computed(() => {
  return Math.max(
    0,
    Math.min(
      props.pos.x,
      window.innerWidth - (contextMenu.value?.clientWidth ?? 0)
    )
  );
});

const isEventInsideMenu = (event: Event) => {
  const target = event.target;
  return target instanceof Node && contextMenu.value?.contains(target);
};

const hideContextMenu = () => {
  emit("hide");
};

const handlePointerDown = (event: Event) => {
  if (!isEventInsideMenu(event)) {
    hideContextMenu();
  }
};

const handleContextMenu = (event: Event) => {
  if (!isEventInsideMenu(event)) {
    hideContextMenu();
  }
};

const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === "Escape") {
    hideContextMenu();
  }
};

const addDismissListeners = () => {
  document.addEventListener("pointerdown", handlePointerDown, true);
  document.addEventListener("contextmenu", handleContextMenu, true);
  document.addEventListener("keydown", handleKeydown, true);
  window.addEventListener("blur", hideContextMenu);
  window.addEventListener("resize", hideContextMenu);
  window.addEventListener("scroll", hideContextMenu, true);
};

const removeDismissListeners = () => {
  document.removeEventListener("pointerdown", handlePointerDown, true);
  document.removeEventListener("contextmenu", handleContextMenu, true);
  document.removeEventListener("keydown", handleKeydown, true);
  window.removeEventListener("blur", hideContextMenu);
  window.removeEventListener("resize", hideContextMenu);
  window.removeEventListener("scroll", hideContextMenu, true);
};

watch(
  () => props.show,
  async (val) => {
    removeDismissListeners();

    if (val) {
      await nextTick();
      addDismissListeners();
    }
  },
  { immediate: true }
);

onBeforeUnmount(() => {
  removeDismissListeners();
});
</script>
