import { defineStore } from "pinia";
// import { useAuthPreferencesStore } from "./auth-preferences";
// import { useAuthEmailStore } from "./auth-email";

export const useLayoutStore = defineStore("layout", {
  // convert to a function
  state: (): {
    loading: boolean;
    prompts: PopupProps[];
    showShell: boolean | null;
  } => ({
    loading: false,
    prompts: [],
    showShell: false,
  }),
  getters: {
    currentPrompt(state) {
      return state.prompts.length > 0
        ? state.prompts[state.prompts.length - 1]
        : null;
    },
    currentPromptName(): string | null | undefined {
      return this.currentPrompt?.prompt;
    },
    // user and jwt getter removed, no longer needed
  },
  actions: {
    // no context as first argument, use `this` instead
    toggleShell() {
      this.showShell = !this.showShell;
    },
    showHover(value: PopupProps | string) {
      if (typeof value !== "object") {
        this.prompts.push({
          prompt: value,
          confirm: null,
          action: undefined,
          props: null,
        });
        return;
      }

      this.prompts.push({
        prompt: value.prompt,
        confirm: value?.confirm,
        action: value?.action,
        props: value?.props,
      });
    },
    showError() {
      this.prompts.push({
        prompt: "error",
        confirm: null,
        action: undefined,
        props: null,
      });
    },
    showSuccess() {
      this.prompts.push({
        prompt: "success",
        confirm: null,
        action: undefined,
        props: null,
      });
    },
    closeHovers() {
      this.prompts.pop();
    },
    // easily reset state using `$reset`
    clearLayout() {
      this.$reset();
    },
  },
});
