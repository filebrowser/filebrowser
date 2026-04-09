import { partial } from "filesize";

/**
 * Formats filesize as KiB/MiB/...
 */
export const filesize = partial({ base: 2 });

export const vClickOutside = {
  created(el: HTMLElement, binding: any) {
    el.clickOutsideEvent = (event: Event) => {
      const target = event.target;

      if (target instanceof Node) {
        if (!el.contains(target)) {
          binding.value(event);
        }
      }
    };

    document.addEventListener("click", el.clickOutsideEvent);
  },

  unmounted(el: HTMLElement) {
    if (el.clickOutsideEvent) {
      document.removeEventListener("click", el.clickOutsideEvent);
    }
  },
};
