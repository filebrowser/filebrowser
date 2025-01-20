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
    setCloseOnPrompt(closeFunction: () => Promise<string>, onPrompt: string) {
      const prompt = this.prompts.find((prompt) => prompt.prompt === onPrompt);
      if (prompt) {
        prompt.close = closeFunction;
      }
    },
    showHover(value: PopupProps | string) {
      if (typeof value !== "object") {
        this.prompts.push({
          prompt: value,
          confirm: null,
          action: undefined,
          props: null,
          close: null,
        });
        return;
      }

      this.prompts.push({
        prompt: value.prompt,
        confirm: value?.confirm,
        action: value?.action,
        props: value?.props,
        close: value?.close,
      });
    },
    showError() {
      this.prompts.push({
        prompt: "error",
        confirm: null,
        action: undefined,
        props: null,
        close: null,
      });
    },
    showSuccess() {
      this.prompts.push({
        prompt: "success",
        confirm: null,
        action: undefined,
        props: null,
        close: null,
      });
    },
    closeHovers() {
      this.prompts.shift()?.close?.();
    },
    // easily reset state using `$reset`
    clearLayout() {
      this.$reset();
    },
  },
});
