<template>
  <div
    class="context-menu"
    ref="contextMenu"
    v-show="show"
    :style="{
      top: `${top}px`,
      left: `${left}px`,
    }"
  >
    <slot />
  </div>
</template>

<script>
export default {
  name: "context-menu",
  props: ["show", "pos"],
  computed: {
    top() {
      return Math.min(
        this.pos.y,
        window.innerHeight - this.$refs.contextMenu?.clientHeight ?? 0
      );
    },
    left() {
      return Math.min(
        this.pos.x,
        window.innerWidth - this.$refs.contextMenu?.clientWidth ?? 0
      );
    },
  },
  methods: {
    hideContextMenu() {
      this.$emit("hide");
    },
  },
  watch: {
    show: function (val) {
      if (val) {
        document.addEventListener("click", this.hideContextMenu);
      } else {
        document.removeEventListener("click", this.hideContextMenu);
      }
    },
  },
};
</script>
