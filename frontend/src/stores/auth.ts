import { defineStore } from "pinia";
import dayjs from "dayjs";
import i18n, { detectLocale } from "@/i18n";
import { cloneDeep } from "lodash-es";

export const useAuthStore = defineStore("auth", {
  // convert to a function
  state: (): {
    user: user | null,
    jwt: string
  } => ({
    user: null,
    jwt: "",
  }),
  getters: {
    // user and jwt getter removed, no longer needed
    isLoggedIn: (state) => state.user !== null,
  },
  actions: {
    // no context as first argument, use `this` instead
    setUser(value: user) {
      if (value === null) {
        this.user = null;
        return;
      }

      const locale = value.locale || detectLocale();
      dayjs.locale(locale);
      // @ts-ignore Don't know how to fix this yet
      i18n.global.locale.value = locale;
      this.user = value;
    },
    updateUser(value: user) {
      if (typeof value !== "object") return;

      let field: userKey
      for (field in value) {
        if (field === "locale") {
          const locale = value[field];
          dayjs.locale(locale);
          // @ts-ignore Don't know how to fix this yet
          i18n.global.locale.value = locale;
        }

        // @ts-ignore to fix
        this.user[field] = cloneDeep(value[field]);
      }
    },
    // easily reset state using `$reset`
    clearUser() {
      this.$reset();
    },
  },
});
