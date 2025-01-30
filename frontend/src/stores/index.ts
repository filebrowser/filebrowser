import { createPinia as _createPinia } from "pinia";
import { markRaw } from "vue";
import type { Router } from "vue-router";

export default function createPinia(router: Router) {
  const pinia = _createPinia();
  pinia.use(({ store }) => {
    store.router = markRaw(router);
  });

  return pinia;
}
