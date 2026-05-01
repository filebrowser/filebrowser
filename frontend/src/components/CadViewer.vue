<template>
  <div class="cad-viewer">
    <div v-if="error" class="cad-error">{{ error }}</div>
    <div v-else-if="loading" class="cad-loading">
      Loading {{ filename }}…
    </div>
    <div ref="canvasContainer" class="cad-canvas"></div>
  </div>
</template>

<!--
  CAD viewer — scaffolding for STEP / IGES / BREP / STL preview.
  Lazy-loads `occt-import-js` (OpenCascade WASM) so the main bundle
  stays small for users who never open CAD.

  Status: scaffold only. The STEP load path is sketched but disabled —
  uncommenting requires `occt-import-js` to be in package.json and the
  WASM blob copied into /public/. See docs/CAD_VIEWER_TODO.md for the
  staged plan.
-->

<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref } from "vue";
import * as THREE from "three";
import { OrbitControls } from "three/examples/jsm/controls/OrbitControls.js";

const props = defineProps<{
  /** Raw bytes of the CAD file. */
  data: ArrayBuffer | null;
  /** Filename for extension detection + display. */
  filename: string;
}>();

const canvasContainer = ref<HTMLDivElement | null>(null);
const loading = ref(false);
const error = ref<string | null>(null);

let scene: any = null;
let camera: any = null;
let renderer: any = null;
let controls: any = null;
let meshGroup: any = null;
let animationId: number | null = null;
let resizeHandler: (() => void) | null = null;

function initThree() {
  const el = canvasContainer.value;
  if (!el) return;

  const w = el.clientWidth || 400;
  const h = el.clientHeight || 300;

  scene = new THREE.Scene();
  scene.background = new THREE.Color(0x1a1a1a);

  camera = new THREE.PerspectiveCamera(50, w / h, 0.01, 50000);
  camera.position.set(100, 100, 100);

  scene.add(new THREE.AmbientLight(0xffffff, 0.6));
  const dir = new THREE.DirectionalLight(0xffffff, 0.6);
  dir.position.set(200, 300, 200);
  scene.add(dir);

  renderer = new THREE.WebGLRenderer({ antialias: true });
  renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
  renderer.setSize(w, h);
  el.appendChild(renderer.domElement);

  controls = new OrbitControls(camera, renderer.domElement);
  controls.enableDamping = true;
  controls.dampingFactor = 0.1;

  resizeHandler = () => {
    if (!renderer || !camera || !canvasContainer.value) return;
    const ww = canvasContainer.value.clientWidth || 400;
    const hh = canvasContainer.value.clientHeight || 300;
    renderer.setSize(ww, hh);
    camera.aspect = ww / hh;
    camera.updateProjectionMatrix();
  };
  window.addEventListener("resize", resizeHandler);

  const animate = () => {
    if (!renderer || !scene || !camera) return;
    controls?.update();
    renderer.render(scene, camera);
    animationId = requestAnimationFrame(animate);
  };
  animate();
}

function ext(name: string): string {
  const dot = name.lastIndexOf(".");
  return dot < 0 ? "" : name.slice(dot + 1).toLowerCase();
}

async function loadFile() {
  if (!props.data) return;
  const e = ext(props.filename);
  loading.value = true;
  error.value = null;

  try {
    if (e === "step" || e === "stp" || e === "iges" || e === "igs" || e === "brep") {
      // TODO(occt-import-js): once `occt-import-js` is added to package.json
      // and the WASM blob is reachable, replace this stub with:
      //
      //   const occtFactory = (await import("occt-import-js")).default;
      //   const occt = await occtFactory({ locateFile: (f) => `/${f}` });
      //   const result = occt.ReadStepFile(new Uint8Array(props.data!));
      //   for (const mesh of result.meshes) addMesh(mesh);
      //
      // See docs/CAD_VIEWER_TODO.md for the full plan.
      throw new Error("STEP/IGES viewer not wired in yet — see docs/CAD_VIEWER_TODO.md");
    }
    if (e === "stl") {
      const { STLLoader } = await import("three/examples/jsm/loaders/STLLoader.js");
      const geom = new STLLoader().parse(props.data);
      addRawMesh(geom);
      return;
    }
    throw new Error(`Unsupported CAD format: .${e}`);
  } catch (err: any) {
    error.value = err?.message || String(err);
  } finally {
    loading.value = false;
  }
}

function addRawMesh(geom: any) {
  if (!scene) return;
  if (meshGroup) {
    scene.remove(meshGroup);
    meshGroup.traverse((c: any) => {
      c.geometry?.dispose();
      c.material?.dispose();
    });
  }
  meshGroup = new THREE.Group();
  const mat = new THREE.MeshStandardMaterial({ color: 0x9aa5b1, metalness: 0.1, roughness: 0.7 });
  meshGroup.add(new THREE.Mesh(geom, mat));
  scene.add(meshGroup);
  fitCameraToObject(meshGroup);
}

function fitCameraToObject(obj: any) {
  if (!camera || !controls) return;
  const box = new THREE.Box3().setFromObject(obj);
  const center = new THREE.Vector3();
  const size = new THREE.Vector3();
  box.getCenter(center);
  box.getSize(size);
  const maxDim = Math.max(size.x || 1, size.y || 1, size.z || 1);
  const dist = maxDim * 2.5;
  camera.position.set(center.x + dist, center.y + dist, center.z + dist);
  camera.near = maxDim * 0.0001;
  camera.far = maxDim * 100;
  camera.lookAt(center);
  camera.updateProjectionMatrix();
  controls.target.copy(center);
  controls.update();
}

onMounted(async () => {
  initThree();
  await loadFile();
});

onBeforeUnmount(() => {
  if (animationId != null) cancelAnimationFrame(animationId);
  if (resizeHandler) {
    window.removeEventListener("resize", resizeHandler);
    resizeHandler = null;
  }
  if (meshGroup && scene) {
    scene.remove(meshGroup);
    meshGroup.traverse((c: any) => {
      c.geometry?.dispose();
      c.material?.dispose();
    });
    meshGroup = null;
  }
  controls?.dispose();
  if (renderer) {
    renderer.dispose();
    if (renderer.domElement && canvasContainer.value?.contains(renderer.domElement)) {
      canvasContainer.value.removeChild(renderer.domElement);
    }
  }
  scene = null;
  camera = null;
  renderer = null;
  controls = null;
});
</script>

<style scoped>
.cad-viewer {
  position: relative;
  width: 100%;
  height: 100%;
  background: #1a1a1a;
  display: flex;
  flex-direction: column;
}
.cad-canvas {
  flex: 1;
  min-height: 0;
}
.cad-error,
.cad-loading {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #aaa;
  font-family: monospace;
  font-size: 13px;
  background: #1a1a1a;
  z-index: 10;
}
.cad-error {
  color: #f88;
}
</style>
