import { defineStore } from "pinia";
import { detectLocale, setLocale } from "@/i18n";
import { cloneDeep } from "lodash-es";
import { useStorage } from "@vueuse/core";
import { computed, ref } from "vue";

export const useAuthStore = defineStore("auth", () => {
  const registeredUser = ref<IUser | null>(null);
  const jwt = ref("");

  const guestJwt = ref("");
  const guestUser = useStorage("guest", {
    locale: "zh-cn",
    viewMode: "list",
    singleClick: false,
    perm: { create: false },
  });

  const shareConfig = useStorage("share-config", {
    sortBy: "name",
    asc: false,
  });
  const isLoggedIn = computed(() => registeredUser.value !== null);
  const user = computed({
    get: () =>
      isLoggedIn.value ? registeredUser.value : (guestUser.value as IUser),
    set: (val) => {
      if (isLoggedIn.value) {
        registeredUser.value = val;
      } else {
        guestUser.value = val;
      }
    },
  });

  function setUser(_user: IUser) {
    if (_user === null) {
      registeredUser.value = null;
      return;
    }

    setLocale(_user.locale || detectLocale());
    registeredUser.value = _user;
  }

  function updateUser(_user: Partial<IUser>) {
    if (_user.locale) {
      setLocale(_user.locale);
    }

    user.value = {
      ...user.value,
      ...cloneDeep(_user),
    } as IUser;
  }
  // easily reset state using `$reset`
  function clearUser() {
    registeredUser.value = null;
  }

  return {
    jwt,
    guestJwt,
    shareConfig,

    isLoggedIn,
    user,

    setUser,
    updateUser,
    clearUser,
  };
});
