import { disableExternal } from "@/utils/constants";
import { createApp } from "vue";
import VueLazyload from "vue-lazyload";
import Toast, { useToast } from "vue-toastification";
import createPinia from "@/stores";
import router from "@/router";
import i18n, { rtlLanguages } from "@/i18n";
import App from "@/App.vue";
import CustomToast from "@/components/CustomToast.vue";

import dayjs from "dayjs";
import localizedFormat from "dayjs/plugin/localizedFormat";
import relativeTime from "dayjs/plugin/relativeTime";
import duration from "dayjs/plugin/duration";
import "./css/styles.css";

// register dayjs plugins globally
dayjs.extend(localizedFormat);
dayjs.extend(relativeTime);
dayjs.extend(duration);

const pinia = createPinia(router);

const app = createApp(App);

app.use(VueLazyload);
app.use(Toast, {
  transition: "Vue-Toastification__bounce",
  maxToasts: 10,
  newestOnTop: true,
});

app.use(i18n);
app.use(pinia);
app.use(router);

app.mixin({
  mounted() {
    // expose vue instance to components
    this.$el.__vue__ = this;
  },
});

// provide v-focus for components
app.directive("focus", {
  mounted: (el) => {
    // initiate focus for the element
    el.focus();
  },
});

const toastConfig = {
  position: "bottom-center",
  timeout: 4000,
  closeOnClick: true,
  pauseOnFocusLoss: true,
  pauseOnHover: true,
  draggable: true,
  draggablePercent: 0.6,
  showCloseButtonOnHover: false,
  hideProgressBar: false,
  closeButton: "button",
  icon: true,
};

app.provide("$showSuccess", (message: string) => {
  const $toast = useToast();
  $toast.success(
    {
      component: CustomToast,
      props: {
        message: message,
      },
    },
    // their type defs are messed up
    //@ts-ignore
    { ...toastConfig, rtl: rtlLanguages.includes(i18n.global.locale) }
  );
});

app.provide("$showError", (error: Error | string, displayReport = true) => {
  const $toast = useToast();
  $toast.error(
    {
      component: CustomToast,
      props: {
        message: (error as Error).message || error,
        isReport: !disableExternal && displayReport,
        // TODO: i couldnt use $t inside the component
        //@ts-ignore
        reportText: i18n.global.t("buttons.reportIssue"),
      },
    },
    // their type defs are messed up
    //@ts-ignore
    {
      ...toastConfig,
      timeout: 0,
      rtl: rtlLanguages.includes(i18n.global.locale),
    }
  );
});

router.isReady().then(() => app.mount("#app"));
