import { defineStore } from "pinia";

/**
 * Dummy store for exposing router to be used in other stores
 * @example
 *  // route: () => {
 *  //   const routerStore = useRouterStore();
 *  //   return routerStore.router.currentRoute;
 *  // },
 */
export const useRouterStore = defineStore("routerStore", () => {
  // can be an empty definition
  return {};
});
