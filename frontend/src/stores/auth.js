import { defineStore } from "pinia";
// import { useAuthPreferencesStore } from "./auth-preferences";
// import { useAuthEmailStore } from "./auth-email";

export const useAuthStore = defineStore("auth", {
  // convert to a function
  state: () => ({
    user: null,
    jwt: "",
  }),
  getters: {
    // user and jwt getter removed, no longer needed
    isLoggedIn: (state) => state.user !== null,
  },
  actions: {
    // no context as first argument, use `this` instead
    setUser(value) {
      if (value === null) {
        this.user = null;
        return;
      }

      // const locale = value.locale || i18n.detectLocale();
      // moment.locale(locale);
      // i18n.default.locale = locale;
      this.user = value;
    },
    updateUser(value) {
      if (typeof value !== "object") return;

      for (let field in value) {
        // if (field === "locale") {
        //   moment.locale(value[field]);
        //   i18n.default.locale = value[field];
        // }

        this.user[field] = structuredClone(value[field]);
      }
    },
    // easily reset state using `$reset`
    clearUser() {
      this.$reset();
    },
  },
});
