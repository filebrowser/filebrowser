<template>
  <svg
    :viewBox="`0 0 ${VBW} ${VBH}`"
    :width="width"
    :height="height"
    class="tool-geom"
    :class="{ 'tool-geom--missing': !drawable }"
    :aria-label="ariaLabel"
    role="img"
  >
    <!-- "missing data" state: dashed outline + ?  -->
    <template v-if="!drawable">
      <rect
        x="2"
        y="2"
        :width="VBW - 4"
        :height="VBH - 4"
        fill="none"
        stroke="#bbb"
        stroke-width="0.6"
        stroke-dasharray="2 2"
      />
      <text
        :x="VBW / 2"
        :y="VBH / 2 + 2"
        text-anchor="middle"
        font-size="6"
        fill="#bbb"
      >?</text>
    </template>

    <!-- drawn tool -->
    <template v-else>
      <!-- Holder (V-flange/BT shape) — fixed 8 SVG units tall.
           Top edge is narrower than the bottom (which matches body
           width). Uniform across all tools so the body proportions
           stay readable even on tiny thumbnails. -->
      <path
        :d="holderPath"
        :fill="holderFill"
        stroke="#555"
        stroke-width="0.3"
      />
      <!-- Body — diameter × length, centered horizontally on the
           holder. Cutting end is at the SVG bottom so a horizontal
           reference line groups visually align across tools. -->
      <rect
        :x="bodyX"
        :y="HOLDER_H"
        :width="bodyW"
        :height="bodyH"
        :fill="bodyFill"
        stroke="#1565c0"
        stroke-width="0.3"
      />
      <!-- Cutting plane reference — same y across all tools so the
           visual gives an at-a-glance "how far does this tool stick
           out" comparison when rendered side-by-side. -->
      <line
        x1="0"
        :y1="VBH - 0.5"
        :x2="VBW"
        :y2="VBH - 0.5"
        stroke="#c62828"
        stroke-width="0.4"
      />
    </template>
  </svg>
</template>

<script setup lang="ts">
import { computed } from "vue";

// Tool geometry as a side-on side view. Length runs vertically (top =
// holder, bottom = cutting tip), diameter runs horizontally. The
// component scales the body to (lengthRatio, diameterRatio) — caller
// is expected to compute those by dividing each tool's effective
// values by the magazine-wide max so all tools share one scale.
//
// Why ratios instead of absolute units: tools span 0.125" to 6" on
// the length axis and 0.0625" to 1.5" on the diameter axis. Drawing
// to absolute pixels would either make small tools invisible or
// large tools clip out of frame. Ratios + a magazine-wide max gives
// a comparable view at any scale.
const props = defineProps<{
  // Effective length and diameter from the tool table. Either may be
  // undefined; if so the SVG renders the dashed "missing" state.
  lengthRatio?: number;
  diameterRatio?: number;
  width: number | string;
  height: number | string;
  // Renamed from `slot` — Vue's eslint plugin treats bare `slot`
  // bindings as the deprecated named-slot attribute and rejects them.
  slotNumber?: number;
  // Optional badge color to recolor the body — caller can use this
  // to hint job state ("tool currently in spindle", etc).
  bodyFill?: string;
  holderFill?: string;
}>();

// SVG viewBox is in arbitrary units; the body fills (W, H) minus the
// holder. Keeping these as small ints means stroke widths of 0.3-0.6
// give a sharp 1 px line at typical render sizes (40-200 CSS px).
const VBW = 30;
const VBH = 100;
const HOLDER_H = 8;
const BODY_AVAILABLE_H = VBH - HOLDER_H - 1; // -1 for the red reference line

const drawable = computed(
  () =>
    typeof props.lengthRatio === "number" &&
    typeof props.diameterRatio === "number" &&
    props.lengthRatio > 0 &&
    props.diameterRatio > 0
);

const bodyW = computed(() => {
  // Min 1.5 svg-units so very-small-diameter tools (1/16" drills)
  // are still visible. Otherwise scale linearly.
  const r = props.diameterRatio ?? 0;
  return Math.max(1.5, r * VBW);
});

const bodyH = computed(() => {
  const r = props.lengthRatio ?? 0;
  return Math.max(2, r * BODY_AVAILABLE_H);
});

const bodyX = computed(() => (VBW - bodyW.value) / 2);

const holderPath = computed(() => {
  // Trapezoid: top is 60% of body width, bottom matches body width.
  // Bottoms of holder = HOLDER_H, tops = 0.
  const bw = bodyW.value;
  const bx = bodyX.value;
  const topW = bw * 0.6;
  const topX = (VBW - topW) / 2;
  return `M ${topX} 0 L ${topX + topW} 0 L ${bx + bw} ${HOLDER_H} L ${bx} ${HOLDER_H} Z`;
});

const bodyFill = computed(() => props.bodyFill || "#1976d2");
const holderFill = computed(() => props.holderFill || "#9e9e9e");

const ariaLabel = computed(() =>
  props.slotNumber !== undefined ? `Tool ${props.slotNumber}` : "Tool"
);
</script>

<style scoped>
.tool-geom {
  display: block;
  overflow: visible;
}

.tool-geom--missing {
  opacity: 0.5;
}
</style>
