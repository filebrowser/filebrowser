import { defineStore } from "pinia";
import dayjs from "dayjs";
import i18n, { detectLocale } from "@/i18n";
import { cloneDeep } from "lodash-es";

export const useAuthStore = defineStore("auth", {
  // convert to a function
  state: (): {
    user: IUser | null;
    jwt: string;
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
    setUser(value: IUser) {
      if (value === null) {
        this.user = null;
        return;
      }

      const locale = value.locale || detectLocale();
      dayjs.locale(locale);
      // according to doc u only need .value if legacy: false
      // in createI18n but they lied
      // https://vue-i18n.intlify.dev/guide/essentials/scope.html#local-scope-1
      //@ts-ignore
      i18n.global.locale = locale;
      this.user = value;
    },
    updateUser(value: IUser) {
      if (typeof value !== "object" || !value) return;
      this.setUser(cloneDeep(value));
    },
    // easily reset state using `$reset`
    clearUser() {
      this.$reset();
    },
  },
});
