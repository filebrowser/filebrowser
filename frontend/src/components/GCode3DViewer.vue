<template>
  <div class="gcode-3d-viewer">
    <div v-if="webglError" class="webgl-error">{{ webglError }}</div>
    <div class="viewer-toolbar">
      <span class="viewer-info">
        {{ pointCount.toLocaleString() }} pts
        <span v-if="lastTruncated" class="truncated-badge">truncated</span>
      </span>
      <button class="viewer-btn" @click="resetCamera" title="Reset camera">⌂</button>
    </div>
    <div ref="canvasContainer" class="gcode-canvas"></div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, watch, nextTick } from "vue";
import * as THREE from "three";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls.js";

const props = defineProps<{
  gcode: string;
  cursorLine?: number | null;
}>();

// FIX 1: emit name matches @select-line in Editor.vue
const emit = defineEmits<{
  (e: "select-line", lineIndex: number): void;
}>();

const canvasContainer = ref<HTMLDivElement | null>(null);

// three.js objects — typed as any to avoid THREE namespace TS issues
let scene: any = null;
let camera: any = null;
let renderer: any = null;
let rapidLines: any = null;   // G0 moves — gray
let feedLines: any = null;    // G1/G2/G3 moves — blue
let controls: any = null;
let highlightCross: any = null;
let animationId: number | null = null;

// bounding box center + size for camera reset
let sceneCenter: any = null;
let sceneDist = 200;

let rapidLineSrc: number[] = [];   // rapidLineSrc[i] = original file line for rapidPoints[i]
let feedLineSrc: number[] = [];    // feedLineSrc[i]  = original file line for feedPoints[i]
let resizeHandler: (() => void) | null = null;
let clickHandler: ((e: MouseEvent) => void) | null = null;

const lastTruncated = ref(false);
const pointCount = ref(0);
const isThreeReady = ref(false);
const webglError = ref<string | null>(null);

// Caps: input that exceeds these is sampled (and the "truncated" badge shown).
// Routine decimation for snappy interaction does NOT trigger the badge.
const MAX_LINES   = 1_500_000;
const MAX_POINTS  = 750_000;
const TARGET_POINTS = 250_000;
// Click radius in screen pixels — clicks farther than this don't snap to a point.
const CLICK_PICK_PIXELS = 30;

interface ParseResult {
  rapidPoints: any[];
  feedPoints: any[];
  rapidLineSrc: number[];
  feedLineSrc: number[];
  totalSourceLines: number;
  truncated: boolean;
}

// ── Parser ───────────────────────────────────────────────────────────────────
function parseGcode(raw: string): ParseResult | null {
  let truncated = false;

  // Split all lines first, then stride-sample evenly across the whole file.
  // This ensures a 13 MB file gets geometry from start to end instead of
  // being cut off after the first MAX_CHARS bytes.
  const allLines = raw.split(/\r?\n/);
  const srcTotal = allLines.length;

  let lines: string[];
  let lineOrigins: number[]; // lineOrigins[i] = original file line index

  if (allLines.length > MAX_LINES) {
    const step = Math.ceil(allLines.length / MAX_LINES);
    lines = [];
    lineOrigins = [];
    for (let i = 0; i < allLines.length; i += step) {
      lines.push(allLines[i]);
      lineOrigins.push(i);
    }
    truncated = true;
  } else {
    lines = allLines;
    lineOrigins = allLines.map((_, i) => i);
  }

  // separate point arrays so we can color rapids vs feeds differently
  const rapidPoints: any[] = [];
  const feedPoints: any[] = [];
  const rSrc: number[] = []; // source line per rapid point
  const fSrc: number[] = []; // source line per feed point

  let x = 0, y = 0, z = 0;
  let currentMode = 1; // default to G1 if no G word seen yet
  let totalPoints = 0;

  for (let idx = 0; idx < lines.length; idx++) {
    const rawLine = lines[idx];
    const line = rawLine.trim().toUpperCase();
    if (!line || line.startsWith("(") || line.startsWith(";")) continue;

    const origLine = lineOrigins[idx];

    // update modal G state
    const gMatches = [...line.matchAll(/G(\d+(?:\.\d+)?)/g)];
    if (gMatches.length) {
      const gNum = parseFloat(gMatches[gMatches.length - 1][1]);
      if ([0, 1, 2, 3].includes(Math.floor(gNum))) {
        currentMode = Math.floor(gNum);
      }
    }

    const startX = x;
    const startY = y;
    const startZ = z;

    const xMatch = line.match(/X(-?\d+(?:\.\d+)?)/);
    const yMatch = line.match(/Y(-?\d+(?:\.\d+)?)/);
    const zMatch = line.match(/Z(-?\d+(?:\.\d+)?)/);
    const iMatch = line.match(/I(-?\d+(?:\.\d+)?)/);
    const jMatch = line.match(/J(-?\d+(?:\.\d+)?)/);

    if (xMatch) x = parseFloat(xMatch[1]);
    if (yMatch) y = parseFloat(yMatch[1]);
    if (zMatch) z = parseFloat(zMatch[1]);

    const hasMotion = xMatch || yMatch || zMatch;
    if (!hasMotion) continue;

    const isRapid = currentMode === 0;
    const target    = isRapid ? rapidPoints : feedPoints;
    const targetSrc = isRapid ? rSrc        : fSrc;

    // seed first point of this array with the start position
    if (target.length === 0) {
      target.push(new THREE.Vector3(startX, startY, startZ));
      targetSrc.push(origLine);
      totalPoints++;
    }

    if ((currentMode === 2 || currentMode === 3) && (iMatch || jMatch)) {
      // arc interpolation in XY
      const cx = startX + (iMatch ? parseFloat(iMatch[1]) : 0);
      const cy = startY + (jMatch ? parseFloat(jMatch[1]) : 0);

      const startVecX = startX - cx;
      const startVecY = startY - cy;
      const endVecX = x - cx;
      const endVecY = y - cy;

      const r = Math.hypot(startVecX, startVecY) || 0.0001;
      const startAngle = Math.atan2(startVecY, startVecX);
      const endAngle = Math.atan2(endVecY, endVecX);

      let delta = endAngle - startAngle;
      if (currentMode === 2 && delta > 0) delta -= Math.PI * 2; // CW
      else if (currentMode === 3 && delta < 0) delta += Math.PI * 2; // CCW

      const arcLen = Math.abs(delta * r);
      const segments = Math.max(8, Math.min(64, Math.ceil(arcLen / 0.5)));

      for (let s = 1; s <= segments; s++) {
        const t = s / segments;
        const ang = startAngle + delta * t;
        feedPoints.push(
          new THREE.Vector3(
            cx + Math.cos(ang) * r,
            cy + Math.sin(ang) * r,
            startZ + (z - startZ) * t
          )
        );
        fSrc.push(origLine);
        totalPoints++;
        if (totalPoints >= MAX_POINTS) { truncated = true; break; }
      }
    } else {
      target.push(new THREE.Vector3(x, y, z));
      targetSrc.push(origLine);
      totalPoints++;
    }

    if (totalPoints >= MAX_POINTS) { truncated = true; break; }
  }

  if (rapidPoints.length + feedPoints.length < 2) {
    console.warn("[GCode3DViewer] not enough motion to build geometry");
    return null;
  }

  // Decimate down to TARGET_POINTS for snappy rendering. This is NOT
  // truncation — coverage stays end-to-end, we just thin the points.
  function decimate(pts: any[], src: number[]): { pts: any[]; src: number[] } {
    if (pts.length <= TARGET_POINTS) return { pts, src };
    const step = Math.ceil(pts.length / TARGET_POINTS);
    const outPts: any[] = [];
    const outSrc: number[] = [];
    for (let i = 0; i < pts.length; i += step) {
      outPts.push(pts[i]);
      outSrc.push(src[i]);
    }
    return { pts: outPts, src: outSrc };
  }

  const { pts: finalRapid, src: finalRapidSrc } = decimate(rapidPoints, rSrc);
  const { pts: finalFeed,  src: finalFeedSrc  } = decimate(feedPoints,  fSrc);

  return {
    rapidPoints: finalRapid,
    feedPoints:  finalFeed,
    rapidLineSrc: finalRapidSrc,
    feedLineSrc:  finalFeedSrc,
    totalSourceLines: srcTotal,
    truncated,
  };
}

// ── Highlight crosshair ──────────────────────────────────────────────────────
// A small 3D cross (three orthogonal segments) that lands on the picked
// vertex. More visually precise than a sphere for "I picked exactly this point".
function ensureHighlightCross(size: number) {
  if (!scene) return;
  if (highlightCross) {
    scene.remove(highlightCross);
    highlightCross.traverse((c: any) => {
      c.geometry?.dispose();
      c.material?.dispose();
    });
    highlightCross = null;
  }
  const mat = new THREE.LineBasicMaterial({
    color: 0xff3300,
    depthTest: false, // always visible on top of toolpath
    transparent: true,
  });
  const group = new THREE.Group();
  group.renderOrder = 999;
  for (const axis of [
    [size, 0, 0],
    [0, size, 0],
    [0, 0, size],
  ]) {
    const geom = new THREE.BufferGeometry().setFromPoints([
      new THREE.Vector3(-axis[0], -axis[1], -axis[2]),
      new THREE.Vector3(axis[0], axis[1], axis[2]),
    ]);
    group.add(new THREE.Line(geom, mat));
  }
  highlightCross = group;
  scene.add(highlightCross);
}

function setHighlightPosition(x: number, y: number, z: number) {
  if (!highlightCross) return;
  highlightCross.position.set(x, y, z);
  highlightCross.visible = true;
}

// ── Geometry management ──────────────────────────────────────────────────────
function clearGeometry() {
  for (const obj of [rapidLines, feedLines]) {
    if (obj && scene) {
      scene.remove(obj);
      obj.geometry.dispose();
      obj.material.dispose();
    }
  }
  rapidLines = null;
  feedLines  = null;
  rapidLineSrc = [];
  feedLineSrc  = [];
}

// ── Three.js init ────────────────────────────────────────────────────────────
function initThree() {
  const el = canvasContainer.value;
  if (!el) return;

  const width  = el.clientWidth  || 400;
  const height = el.clientHeight || 300;

  scene = new THREE.Scene();
  scene.background = new THREE.Color(0x1a1a1a);

  camera = new THREE.PerspectiveCamera(50, width / height, 0.01, 50000);
  camera.position.set(0, 0, 200);

  scene.add(new THREE.AmbientLight(0xffffff, 0.8));
  const dir = new THREE.DirectionalLight(0xffffff, 0.4);
  dir.position.set(100, 200, 100);
  scene.add(dir);

  try {
    renderer = new THREE.WebGLRenderer({ antialias: true });
  } catch (e) {
    webglError.value = "WebGL is not available in this browser.";
    return;
  }
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2)); // cap at 2x for Pi
  renderer.setSize(width, height);
  el.appendChild(renderer.domElement);

  controls = new OrbitControls(camera, renderer.domElement);
  controls.enableDamping    = true;
  controls.dampingFactor    = 0.1;
  controls.enablePan        = true;
  controls.screenSpacePanning = true;

  resizeHandler = () => {
    if (!renderer || !camera || !canvasContainer.value) return;
    const w = canvasContainer.value.clientWidth  || 400;
    const h = canvasContainer.value.clientHeight || 300;
    renderer.setSize(w, h);
    camera.aspect = w / h;
    camera.updateProjectionMatrix();
  };
  window.addEventListener("resize", resizeHandler);

  // Click → nearest vertex in screen space.
  // Raycasting Lines was unreliable: a tiny world-space threshold misses
  // most clicks, a large one snaps to whatever segment happens to lie within
  // the corridor (usually biased toward whichever direction the toolpath runs).
  // Projecting every vertex to screen and picking the nearest gives the user
  // exactly what they clicked, with no orientation bias.
  clickHandler = (event: MouseEvent) => {
    if (!renderer || !camera) return;
    if (event.button !== 0) return;

    const rect = renderer.domElement.getBoundingClientRect();
    const cx = event.clientX - rect.left;
    const cy = event.clientY - rect.top;

    const targets: any[] = [];
    if (feedLines)  targets.push(feedLines);
    if (rapidLines) targets.push(rapidLines);
    if (!targets.length) return;

    const tmp = new THREE.Vector3();
    let bestObj: any = null;
    let bestIdx = 0;
    let bestDistSq = Infinity;

    for (const obj of targets) {
      const posAttr = obj.geometry.getAttribute("position") as any;
      const count   = posAttr.count;
      for (let i = 0; i < count; i++) {
        tmp.set(posAttr.getX(i), posAttr.getY(i), posAttr.getZ(i));
        tmp.project(camera);
        // skip points behind the camera or outside the frustum
        if (tmp.z < -1 || tmp.z > 1) continue;
        const sx = (tmp.x + 1) * 0.5 * rect.width;
        const sy = (1 - tmp.y) * 0.5 * rect.height;
        const dx = sx - cx;
        const dy = sy - cy;
        const d2 = dx * dx + dy * dy;
        if (d2 < bestDistSq) {
          bestDistSq = d2;
          bestObj = obj;
          bestIdx = i;
        }
      }
    }

    if (!bestObj) return;
    if (bestDistSq > CLICK_PICK_PIXELS * CLICK_PICK_PIXELS) return;

    const posAttr = bestObj.geometry.getAttribute("position") as any;
    setHighlightPosition(
      posAttr.getX(bestIdx),
      posAttr.getY(bestIdx),
      posAttr.getZ(bestIdx)
    );

    const src = bestObj === feedLines ? feedLineSrc : rapidLineSrc;
    emit("select-line", src[bestIdx] ?? 0);
  };
  renderer.domElement.addEventListener("click", clickHandler);

  const animate = () => {
    if (!renderer || !scene || !camera) return;
    controls?.update();
    renderer.render(scene, camera);
    animationId = requestAnimationFrame(animate);
  };
  animate();

  isThreeReady.value = true;
}

// ── Geometry update ──────────────────────────────────────────────────────────
function updateGeometry(gcode: string) {
  if (!scene || !camera) return;

  clearGeometry();

  const result = parseGcode(gcode || "");
  if (!result) return;

  const { rapidPoints, feedPoints, rapidLineSrc: rSrc, feedLineSrc: fSrc, truncated } = result;
  rapidLineSrc = rSrc;
  feedLineSrc      = fSrc;
  lastTruncated.value = truncated;
  pointCount.value       = rapidPoints.length + feedPoints.length;

  // G0 rapids — dashed gray
  if (rapidPoints.length >= 2) {
    const geom = new THREE.BufferGeometry().setFromPoints(rapidPoints);
    const mat  = new THREE.LineDashedMaterial({
      color: 0x888888,
      dashSize: 2,
      gapSize: 1,
    });
    rapidLines = new THREE.Line(geom, mat);
    rapidLines.computeLineDistances(); // required for dashed material
    scene.add(rapidLines);
  }

  // G1/G2/G3 feeds — blue
  if (feedPoints.length >= 2) {
    const geom = new THREE.BufferGeometry().setFromPoints(feedPoints);
    const mat  = new THREE.LineBasicMaterial({ color: 0x4287f5 });
    feedLines = new THREE.Line(geom, mat);
    scene.add(feedLines);
  }

  // fit camera to combined bounding box
  const allPoints = [...rapidPoints, ...feedPoints];
  if (!allPoints.length) return;

  const bbox = new THREE.Box3();
  for (const p of allPoints) bbox.expandByPoint(p);

  const size   = new THREE.Vector3();
  const center = new THREE.Vector3();
  bbox.getSize(size);
  bbox.getCenter(center);

  sceneCenter = center.clone();
  const maxDim = Math.max(size.x || 1, size.y || 1, size.z || 1);
  sceneDist    = maxDim * 2.5;

  camera.position.set(
    center.x + sceneDist,
    center.y + sceneDist,
    center.z + sceneDist
  );
  camera.near = maxDim * 0.0001;
  camera.far  = maxDim * 100;
  camera.lookAt(center);
  camera.updateProjectionMatrix();

  if (controls) {
    controls.target.copy(center);
    controls.update();
  }

  // Crosshair size scales with part — visible but never engulfing.
  ensureHighlightCross(maxDim * 0.04);
}

// ── Cursor → 3D highlight ────────────────────────────────────────────────────
// Binary-searches the source-line map for the geometry vertex closest to the
// editor cursor's source line. Accurate even when most source lines produce no
// geometry (comments, M-codes, etc.) — linear interpolation drifts badly here.
function highlightLine(line: number | null | undefined) {
  if (!scene || !highlightCross || line == null || line < 0) return;

  const obj = feedLines || rapidLines;
  if (!obj) return;

  const posAttr = obj.geometry.getAttribute("position") as any;
  const count   = posAttr.count || 0;
  if (!count) return;

  const lineSrc = feedLines ? feedLineSrc : rapidLineSrc;
  let lo = 0, hi = lineSrc.length - 1;
  while (lo < hi) {
    const mid = (lo + hi) >> 1;
    if (lineSrc[mid] < line) lo = mid + 1;
    else hi = mid;
  }
  const idx = Math.min(lo, count - 1);

  setHighlightPosition(posAttr.getX(idx), posAttr.getY(idx), posAttr.getZ(idx));
}

// ── Camera reset ─────────────────────────────────────────────────────────────
function resetCamera() {
  if (!camera || !controls || !sceneCenter) return;
  camera.position.set(
    sceneCenter.x + sceneDist,
    sceneCenter.y + sceneDist,
    sceneCenter.z + sceneDist
  );
  camera.lookAt(sceneCenter);
  controls.target.copy(sceneCenter);
  controls.update();
}

// ── Lifecycle ────────────────────────────────────────────────────────────────
onMounted(() => {
  nextTick(() => {
    initThree();
    if (isThreeReady.value) {
      updateGeometry(props.gcode || "");
    }
  });
});

watch(
  () => props.gcode,
  (val) => {
    if (!isThreeReady.value) return;
    updateGeometry(val || "");
  },
  { immediate: false }
);

watch(
  () => props.cursorLine,
  (line) => highlightLine(line ?? null)
);

onBeforeUnmount(() => {
  if (animationId != null) cancelAnimationFrame(animationId);

  if (resizeHandler) {
    window.removeEventListener("resize", resizeHandler);
    resizeHandler = null;
  }

  if (renderer && clickHandler) {
    renderer.domElement.removeEventListener("click", clickHandler);
    clickHandler = null;
  }

  clearGeometry();

  if (highlightCross && scene) {
    scene.remove(highlightCross);
    highlightCross.traverse((c: any) => {
      c.geometry?.dispose();
      c.material?.dispose();
    });
    highlightCross = null;
  }

  controls?.dispose();

  if (renderer) {
    renderer.dispose();
    if (renderer.domElement && canvasContainer.value?.contains(renderer.domElement)) {
      canvasContainer.value.removeChild(renderer.domElement);
    }
  }

  scene = null; camera = null; renderer = null;
  controls = null;
});
</script>

<style scoped>
.gcode-3d-viewer {
  position: relative;
  width: 100%;
  height: 100%;
  background: #1a1a1a;
  display: flex;
  flex-direction: column;
}

.viewer-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 10px;
  background: #111;
  border-bottom: 1px solid #333;
  flex-shrink: 0;
}

.viewer-info {
  font-size: 11px;
  color: #888;
  font-family: monospace;
}

.truncated-badge {
  margin-left: 6px;
  background: #c57d00;
  color: #fff;
  font-size: 10px;
  padding: 1px 5px;
  border-radius: 3px;
}

.viewer-btn {
  background: #222;
  border: 1px solid #444;
  color: #aaa;
  font-size: 13px;
  width: 26px;
  height: 26px;
  border-radius: 3px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.1s;
}

.viewer-btn:hover {
  background: #333;
  color: #fff;
}

.gcode-canvas {
  flex: 1;
  min-height: 0;
}

.webgl-error {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #f88;
  font-size: 13px;
  font-family: monospace;
  background: #1a1a1a;
  z-index: 10;
}
</style>