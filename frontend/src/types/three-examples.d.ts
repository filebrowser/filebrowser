// frontend/src/types/three-examples.d.ts
//
// Shims for the loose-typed `three/examples/jsm/*` entry points we use.
// The official three types ship these too in newer versions, but the
// declarations are spotty across versions and our viewers don't need
// the type detail.

declare module "three/examples/jsm/controls/OrbitControls.js" {
  export const OrbitControls: any;
  const _default: any;
  export default _default;
}

declare module "three/examples/jsm/loaders/STLLoader.js" {
  export const STLLoader: any;
  const _default: any;
  export default _default;
}
