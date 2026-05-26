<template>
  <div
    class="context-menu"
    ref="contextMenu"
    v-show="show"
    :style="{
      top: `${top}px`,
      left: `${left}px`,
      maxHeight: `${maxHeight}px`,
    }"
    @click="hideContextMenu"
    @contextmenu.prevent.stop
  >
    <slot />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onBeforeUnmount, nextTick } from "vue";

const emit = defineEmits(["hide"]);
const props = defineProps<{ show: boolean; pos: { x: number; y: number } }>();
const contextMenu = ref<HTMLElement | null>(null);

const viewportPadding = 8;
const left = ref(0);
const top = ref(0);
const maxHeight = ref(120);

const updateMenuPosition = () => {
  const menu = contextMenu.value;
  const menuWidth = menu?.offsetWidth ?? 0;
  const menuHeight = menu?.offsetHeight ?? 0;

  maxHeight.value = Math.max(120, window.innerHeight - viewportPadding * 2);

  const visibleHeight = Math.min(menuHeight, maxHeight.value);
  const maxLeft = Math.max(
    viewportPadding,
    window.innerWidth - menuWidth - viewportPadding
  );
  const maxTop = Math.max(
    viewportPadding,
    window.innerHeight - visibleHeight - viewportPadding
  );

  left.value = Math.max(viewportPadding, Math.min(props.pos.x, maxLeft));
  top.value = Math.max(viewportPadding, Math.min(props.pos.y, maxTop));
};

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

const handleScroll = (event: Event) => {
  // Keep the menu open when the user scrolls inside the menu itself.
  // Close it only when the page or another scrollable container moves.
  if (isEventInsideMenu(event)) {
    return;
  }

  hideContextMenu();
};

const addDismissListeners = () => {
  document.addEventListener("pointerdown", handlePointerDown, true);
  document.addEventListener("contextmenu", handleContextMenu, true);
  document.addEventListener("keydown", handleKeydown, true);
  window.addEventListener("blur", hideContextMenu);
  window.addEventListener("resize", hideContextMenu);
  window.addEventListener("scroll", handleScroll, true);
};

const removeDismissListeners = () => {
  document.removeEventListener("pointerdown", handlePointerDown, true);
  document.removeEventListener("contextmenu", handleContextMenu, true);
  document.removeEventListener("keydown", handleKeydown, true);
  window.removeEventListener("blur", hideContextMenu);
  window.removeEventListener("resize", hideContextMenu);
  window.removeEventListener("scroll", handleScroll, true);
};

watch(
  () => [props.show, props.pos.x, props.pos.y] as const,
  async ([show]) => {
    removeDismissListeners();

    if (show) {
      await nextTick();
      updateMenuPosition();
      window.requestAnimationFrame(updateMenuPosition);
      addDismissListeners();
    }
  },
  { immediate: true }
);

onBeforeUnmount(() => {
  removeDismissListeners();
});
</script>

<style scoped>
.context-menu {
  overflow-y: auto;
  overscroll-behavior: contain;
}
</style>
