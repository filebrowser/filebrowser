import { createPinia as _createPinia } from "pinia";
import { markRaw } from "vue";

export default function createPinia(router) {
  const pinia = _createPinia();
  pinia.use(({ store }) => {
    store.router = markRaw(router);
  });

  return pinia;
}
