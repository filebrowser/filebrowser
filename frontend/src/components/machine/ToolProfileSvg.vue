<template>
  <svg
    v-if="path"
    :viewBox="viewBox"
    :width="width"
    :height="height"
    preserveAspectRatio="xMidYMid meet"
    class="m-tool-profile"
    role="img"
    :aria-label="ariaLabel"
  >
    <!-- Centerline reference. -->
    <line :x1="0" :y1="0" :x2="0" :y2="totalLength" class="m-tool-profile__axis" />
    <!-- Cutting-region highlight (drawn before silhouette so the
         outline reads cleanly on top). -->
    <rect
      v-if="cuttingBox"
      :x="cuttingBox.x"
      :y="cuttingBox.y"
      :width="cuttingBox.w"
      :height="cuttingBox.h"
      class="m-tool-profile__flutes"
    />
    <!-- Silhouette path: full closed profile (holder + tool body). -->
    <path :d="path" class="m-tool-profile__body" />
    <!-- Gauge plane marker. -->
    <line
      :x1="-maxRadius"
      :y1="0"
      :x2="maxRadius"
      :y2="0"
      class="m-tool-profile__gauge"
    />
  </svg>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { FusionToolEntry, FusionHolderSegment } from "@/api/cnc";

const props = withDefaults(
  defineProps<{
    entry: FusionToolEntry | null;
    width?: number;
    height?: number;
  }>(),
  { width: 80, height: 220 }
);

const ariaLabel = computed(() => {
  if (!props.entry) return "";
  const parts: string[] = [];
  if (props.entry.type) parts.push(props.entry.type);
  if (props.entry.geometry?.DC) parts.push(`D=${props.entry.geometry.DC}`);
  return parts.join(" ");
});

// Build a series of (y, radius) points along the right silhouette.
// y grows downward toward the tool tip. Walk holder segments in
// REVERSE order because Fusion lists them tool-tip-first; we want
// to start from the spindle end (top of the rendered SVG).
interface Point { y: number; r: number }

const profile = computed<{
  points: Point[];
  cuttingStart: number;
  cuttingEnd: number;
  tipKind: "flat" | "ball" | "bull" | "chamfer";
  tipR: number;
  tipBR: number;
} | null>(() => {
  const e = props.entry;
  if (!e) return null;
  const holder = e.holder?.segments || [];
  if (holder.length === 0 && !e.geometry?.DC) return null;

  const pts: Point[] = [];
  let y = 0;
  // Walk segments tip-first → reverse so the spindle end is at y=0.
  const segs: FusionHolderSegment[] = [...holder].reverse();
  for (let i = 0; i < segs.length; i++) {
    const s = segs[i];
    // "upper-diameter" is at the spindle-end of the segment (top of
    // the segment in our visual orientation); "lower-diameter" is at
    // the tool-tip end. Reversing the array means we encounter the
    // spindle-end side first, so at y=current we have upper-diameter
    // and at y=current+height we have lower-diameter.
    if (i === 0) pts.push({ y, r: s["upper-diameter"] / 2 });
    y += s.height;
    pts.push({ y, r: s["lower-diameter"] / 2 });
  }

  // Tool stickout below the holder. LB is the full stickout, LCF is
  // the flute length at the tip. Shank above flutes = LB - LCF.
  const g = e.geometry || {};
  const LB = g.LB || 0;
  const LCF = g.LCF || 0;
  const DC = g.DC || 0;
  const SFDM = g.SFDM || DC || 0;

  if (LB > 0 && DC > 0) {
    const shankLen = Math.max(0, LB - LCF);
    if (shankLen > 0) {
      // Step from holder's last radius to shank radius.
      pts.push({ y, r: SFDM / 2 });
      y += shankLen;
      pts.push({ y, r: SFDM / 2 });
    }
    if (LCF > 0) {
      // Step from shank to flute diameter.
      if (Math.abs(SFDM - DC) > 1e-6) {
        pts.push({ y, r: DC / 2 });
      }
      y += LCF;
      pts.push({ y, r: DC / 2 });
    } else if (shankLen === 0) {
      // Edge case: LB without LCF — just draw a straight tool body
      // at DC for LB length.
      pts.push({ y, r: DC / 2 });
      y += LB;
      pts.push({ y, r: DC / 2 });
    }
  }

  // Tip shape
  let tipKind: "flat" | "ball" | "bull" | "chamfer" = "flat";
  const type = (e.type || "").toLowerCase();
  if (type.includes("ball")) tipKind = "ball";
  else if (type.includes("bull") || (g.RE && g.RE > 0)) tipKind = "bull";
  else if (type.includes("chamfer") || (g.TA && g.TA > 0)) tipKind = "chamfer";

  const cuttingStart = LCF > 0 ? y - LCF : y;
  const cuttingEnd = y;

  return {
    points: pts,
    cuttingStart,
    cuttingEnd,
    tipKind,
    tipR: DC > 0 ? DC / 2 : (pts[pts.length - 1]?.r ?? 0),
    tipBR: g.RE || 0,
  };
});

// Compose the SVG path: trace right side top→tip, draw tip, mirror
// up the left side, close.
const path = computed<string | null>(() => {
  const p = profile.value;
  if (!p || p.points.length < 2) return null;
  const cmds: string[] = [];
  const first = p.points[0];
  cmds.push(`M ${first.r.toFixed(4)} ${first.y.toFixed(4)}`);
  for (let i = 1; i < p.points.length; i++) {
    const pt = p.points[i];
    cmds.push(`L ${pt.r.toFixed(4)} ${pt.y.toFixed(4)}`);
  }
  // Tip
  const tipY = p.cuttingEnd;
  const tipR = p.tipR;
  switch (p.tipKind) {
    case "ball": {
      // Half-circle of radius tipR — arc from (tipR, tipY) to
      // (-tipR, tipY) sweeping through (0, tipY + tipR).
      cmds.push(`A ${tipR.toFixed(4)} ${tipR.toFixed(4)} 0 0 1 ${(-tipR).toFixed(4)} ${tipY.toFixed(4)}`);
      break;
    }
    case "bull": {
      const re = p.tipBR > 0 ? p.tipBR : Math.min(0.03, tipR * 0.2);
      // Right corner arc
      cmds.push(`L ${tipR.toFixed(4)} ${(tipY + re).toFixed(4)}`);
      // Wait — the silhouette so far ends at (tipR, tipY). For a
      // bull-nose, the corner radius sweeps from (tipR, tipY) outward
      // and around to (tipR - re, tipY + re), then across to (-tipR + re, tipY + re),
      // then back to (-tipR, tipY).
      // Reset: simpler is to issue the arcs after a small adjust.
      // Build corner arcs directly without the intermediate L.
      break;
    }
    case "chamfer": {
      // Straight line in at TA from (tipR, tipY) to (0, tipY + tipR/tan(TA))
      const ta = (props.entry?.geometry?.TA || 45) * Math.PI / 180;
      const dy = tipR / Math.tan(ta);
      cmds.push(`L 0 ${(tipY + dy).toFixed(4)}`);
      cmds.push(`L ${(-tipR).toFixed(4)} ${tipY.toFixed(4)}`);
      break;
    }
    case "flat":
    default: {
      cmds.push(`L ${(-tipR).toFixed(4)} ${tipY.toFixed(4)}`);
    }
  }
  // Bull nose has a more intricate shape; handle separately.
  if (p.tipKind === "bull") {
    const re = p.tipBR > 0 ? p.tipBR : Math.min(0.03, tipR * 0.2);
    cmds.length = 0; // rebuild cleanly
    cmds.push(`M ${first.r.toFixed(4)} ${first.y.toFixed(4)}`);
    for (let i = 1; i < p.points.length; i++) {
      const pt = p.points[i];
      cmds.push(`L ${pt.r.toFixed(4)} ${pt.y.toFixed(4)}`);
    }
    // Right corner arc — quarter-circle of radius re from (tipR, tipY)
    // sweeping down + inward.
    cmds.push(`A ${re.toFixed(4)} ${re.toFixed(4)} 0 0 1 ${(tipR - re).toFixed(4)} ${(tipY + re).toFixed(4)}`);
    // Flat across the bottom
    cmds.push(`L ${(-(tipR - re)).toFixed(4)} ${(tipY + re).toFixed(4)}`);
    // Left corner arc back up
    cmds.push(`A ${re.toFixed(4)} ${re.toFixed(4)} 0 0 1 ${(-tipR).toFixed(4)} ${tipY.toFixed(4)}`);
  }
  // Mirror up the left side. Skip the LAST point (we just placed it
  // at -tipR, tipY via the tip command) and walk in reverse from
  // second-to-last back to first.
  for (let i = p.points.length - 2; i >= 0; i--) {
    const pt = p.points[i];
    cmds.push(`L ${(-pt.r).toFixed(4)} ${pt.y.toFixed(4)}`);
  }
  cmds.push("Z");
  return cmds.join(" ");
});

const totalLength = computed<number>(() => {
  const p = profile.value;
  if (!p) return 0;
  const last = p.points[p.points.length - 1];
  // Add tip length for ball/chamfer profiles so the viewBox doesn't
  // crop them.
  const tipAdd = p.tipKind === "ball"
    ? p.tipR
    : p.tipKind === "chamfer"
      ? p.tipR / Math.tan(((props.entry?.geometry?.TA || 45) * Math.PI) / 180)
      : p.tipKind === "bull"
        ? (p.tipBR > 0 ? p.tipBR : Math.min(0.03, p.tipR * 0.2))
        : 0;
  return (last?.y || 0) + tipAdd;
});

const maxRadius = computed<number>(() => {
  const p = profile.value;
  if (!p) return 0;
  let m = 0;
  for (const pt of p.points) {
    if (pt.r > m) m = pt.r;
  }
  return m;
});

const cuttingBox = computed(() => {
  const p = profile.value;
  if (!p || p.cuttingStart === p.cuttingEnd) return null;
  const r = p.tipR;
  return {
    x: -r,
    y: p.cuttingStart,
    w: r * 2,
    h: p.cuttingEnd - p.cuttingStart,
  };
});

const viewBox = computed<string>(() => {
  if (!profile.value) return "0 0 1 1";
  const r = maxRadius.value;
  const h = totalLength.value;
  const pad = Math.max(r * 0.1, 0.05);
  // viewBox: x = -r - pad, y = -pad, width = 2r + 2pad, height = h + 2pad
  return `${(-r - pad).toFixed(4)} ${(-pad).toFixed(4)} ${(r * 2 + pad * 2).toFixed(4)} ${(h + pad * 2).toFixed(4)}`;
});
</script>

<style scoped>
.m-tool-profile {
  display: block;
}
/* vector-effect: non-scaling-stroke keeps strokes 1px regardless of
   the viewBox scale — without it a 3" tall tool drawn into a 160px
   SVG would render strokes at 0.005px = invisible. */
.m-tool-profile__body {
  fill: #cfd2d6;
  stroke: #2c2c2a;
  stroke-width: 1;
  stroke-linejoin: round;
  vector-effect: non-scaling-stroke;
}
.m-tool-profile__flutes {
  fill: rgba(24, 95, 165, 0.25);
}
.m-tool-profile__axis {
  stroke: #888780;
  stroke-width: 1;
  stroke-dasharray: 3 3;
  vector-effect: non-scaling-stroke;
}
.m-tool-profile__gauge {
  stroke: #b44;
  stroke-width: 1;
  stroke-dasharray: 4 4;
  vector-effect: non-scaling-stroke;
}
</style>
