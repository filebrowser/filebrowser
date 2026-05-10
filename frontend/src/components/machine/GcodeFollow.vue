<template>
  <div class="m-gcode" ref="root">
    <div class="m-gcode__overlay">
      <span class="m-gcode__seg">
        {{ following ? t("machine.gcodeFollowing", { n: machineLine }) : t("machine.gcodeDetached") }}
      </span>
      <button class="m-gcode__btn" @click="snapBack">⏎ {{ t("machine.gcodeLive") }}</button>
    </div>
    <div class="m-gcode__scroll" ref="scrollEl" @wheel.passive="onWheel" @touchstart.passive="onTouch">
      <div v-for="(line, i) in lines" :key="i" class="m-gcode__line" :class="{ 'm-gcode__line--active': i + 1 === machineLine }">
        <span class="m-gcode__ln">{{ i + 1 }}</span>
        <span class="m-gcode__code" :class="{ 'm-gcode__code--comment': isComment(line) }">{{ line }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, watch } from "vue";
import { useI18n } from "vue-i18n";

const { t } = useI18n();

const props = defineProps<{
  gcode: string;
  machineLine: number;
}>();

// Operator can scroll freely without affecting the live tracking.
// We auto-follow only while `following` is true; any wheel / touch
// scroll flips it off until they hit the "⏎ live" button.
const following = ref(true);
const scrollEl = ref<HTMLDivElement | null>(null);
const root = ref<HTMLDivElement | null>(null);

const lines = computed(() => (props.gcode || "").split(/\r?\n/));

const isComment = (l: string) => {
  const t = l.trim();
  return t.startsWith("(") || t.startsWith(";");
};

const onWheel = () => {
  following.value = false;
};
const onTouch = () => {
  following.value = false;
};

const snapBack = () => {
  following.value = true;
  scrollToMachineLine();
};

const scrollToMachineLine = () => {
  if (!following.value) return;
  const sc = scrollEl.value;
  if (!sc) return;
  // Each line is rendered at the same height; query the nth child for
  // a robust offset rather than computing px math.
  const idx = Math.max(0, props.machineLine - 1);
  const target = sc.children[idx] as HTMLElement | undefined;
  if (!target) return;
  // Center the active line in the viewport.
  const wantTop = target.offsetTop - sc.clientHeight / 2 + target.clientHeight / 2;
  sc.scrollTo({ top: Math.max(0, wantTop), behavior: "smooth" });
};

watch(() => props.machineLine, () => {
  if (following.value) nextTick(scrollToMachineLine);
});

watch(() => props.gcode, () => {
  if (following.value) nextTick(scrollToMachineLine);
});

onBeforeUnmount(() => {
  following.value = false;
});
</script>

<style scoped>
.m-gcode {
  position: relative;
  display: flex;
  flex-direction: column;
  background: #2C2C2A;
  color: #B4B2A9;
  border-radius: 6px;
  min-height: 0;
  height: 100%;
  overflow: hidden;
}
.m-gcode__overlay {
  position: absolute;
  top: 6px;
  right: 6px;
  display: flex;
  gap: 4px;
  align-items: center;
  z-index: 2;
}
.m-gcode__seg {
  background: rgba(212, 209, 199, 0.1);
  color: #B4B2A9;
  font-size: 9px;
  padding: 2px 6px;
  border-radius: 3px;
}
.m-gcode__btn {
  font-size: 9px;
  padding: 2px 6px;
  border: 1px solid #444441;
  border-radius: 3px;
  background: rgba(0,0,0,0.3);
  color: #D3D1C7;
  cursor: pointer;
}
.m-gcode__btn:hover { background: rgba(0,0,0,0.5); }

.m-gcode__scroll {
  flex: 1;
  overflow-y: auto;
  overscroll-behavior: contain;
  padding: 26px 8px 6px; /* leave room for the overlay */
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 10px;
  line-height: 1.45;
}

.m-gcode__line {
  display: flex;
  gap: 8px;
}
.m-gcode__ln {
  color: #5F5E5A;
  min-width: 28px;
  text-align: right;
  user-select: none;
}
.m-gcode__line--active {
  background: rgba(55, 138, 221, 0.18);
  margin: 0 -8px;
  padding: 0 8px;
  color: #E6F1FB;
}
.m-gcode__code--comment { color: #5F5E5A; }
</style>
