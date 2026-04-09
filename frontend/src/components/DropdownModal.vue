<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount } from "vue";
import { vClickOutside } from "@/utils/index";

const props = withDefaults(
  defineProps<{
    position?: "top-left" | "bottom-left" | "bottom-right" | "top-right";
    closeOnClick?: boolean;
  }>(),
  {
    position: "bottom-right",
    closeOnClick: true,
  }
);

const isOpen = defineModel<boolean>();

const triggerRef = ref<HTMLElement | null>(null);
const listRef = ref<HTMLElement | null>(null);
const dropdownStyle = ref<Record<string, string>>({});

const updatePosition = () => {
  if (!isOpen.value || !triggerRef.value) return;

  const rect = triggerRef.value.getBoundingClientRect();

  dropdownStyle.value = {
    position: "fixed",
    top: props.position.includes("bottom")
      ? `${rect.bottom + 2}px`
      : `${rect.top}px`,
    left: props.position.includes("left") ? `${rect.left}px` : "auto",
    right: props.position.includes("right")
      ? `${window.innerWidth - rect.right}px`
      : "auto",
    zIndex: "11000",
  };
};

watch(isOpen, (open) => {
  if (open) {
    updatePosition();
  }
});

const onWindowChange = () => {
  updatePosition();
};

const closeDropdown = (e: Event) => {
  if (
    e.target instanceof HTMLElement &&
    listRef.value?.contains(e.target) &&
    !props.closeOnClick
  ) {
    return;
  }
  isOpen.value = false;
};

onMounted(() => {
  window.addEventListener("resize", onWindowChange);
  window.addEventListener("scroll", onWindowChange, true);
});

onBeforeUnmount(() => {
  window.removeEventListener("resize", onWindowChange);
  window.removeEventListener("scroll", onWindowChange, true);
});
</script>

<script lang="ts">
export default {
  directives: {
    clickOutside: vClickOutside,
  },
};
</script>

<template>
  <div
    class="dropdown-modal-container"
    v-click-outside="closeDropdown"
    ref="triggerRef"
  >
    <button @click="isOpen = !isOpen" class="dropdown-modal-trigger">
      <slot></slot>
      <i class="material-icons">chevron_right</i>
    </button>

    <teleport to="body">
      <div
        ref="listRef"
        class="dropdown-modal-list"
        :class="{ 'dropdown-modal-open': isOpen }"
        :style="dropdownStyle"
      >
        <div>
          <slot name="list"></slot>
        </div>
      </div>
    </teleport>
  </div>
</template>

<style scoped>
.dropdown-modal-trigger {
  background: var(--surfacePrimary);
  color: var(--textSecondary);
  border: 1px solid var(--borderPrimary);
  border-radius: 0.1em;
  padding: 0.5em 1em;
  transition: 0.2s ease all;
  margin: 0;
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: center;
}

.dropdown-modal-trigger > i {
  transform: rotate(90deg);
}

.dropdown-modal-list {
  padding: 0.25rem;
  background-color: var(--surfacePrimary);
  color: var(--textSecondary);
  display: none;
  border: 1px solid var(--borderPrimary);
  border-radius: 0.1em;
}

.dropdown-modal-list > div {
  max-height: 450px;
  padding: 0.25rem;
}

.dropdown-modal-open {
  display: block;
}
</style>
