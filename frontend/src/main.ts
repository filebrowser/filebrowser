import { disableExternal } from "@/utils/constants";
import { createApp } from "vue";
import VueNumberInput from "@chenfengyuan/vue-number-input";
import VueLazyload from "vue-lazyload";
import { createVfm } from "vue-final-modal";
import Toast, { POSITION, useToast } from "vue-toastification";
import type {
  ToastOptions,
  PluginOptions,
} from "vue-toastification/dist/types/types";
import createPinia from "@/stores";
import router from "@/router";
import i18n, { isRtl } from "@/i18n";
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
const vfm = createVfm();

const app = createApp(App);

app.component(VueNumberInput.name || "vue-number-input", VueNumberInput);
app.use(VueLazyload);
app.use(Toast, {
  transition: "Vue-Toastification__bounce",
  maxToasts: 10,
  newestOnTop: true,
} satisfies PluginOptions);

app.use(vfm);
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
  mounted: async (el) => {
    // initiate focus for the element
    el.focus();
  },
});

const toastConfig = {
  position: POSITION.BOTTOM_CENTER,
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
} satisfies ToastOptions;

app.provide("$showSuccess", (message: string) => {
  const $toast = useToast();
  $toast.success(
    {
      component: CustomToast,
      props: {
        message: message,
      },
    },
    { ...toastConfig, rtl: isRtl() }
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
        // TODO: could you add this to the component itself?
        reportText: i18n.global.t("buttons.reportIssue"),
      },
    },
    {
      ...toastConfig,
      timeout: 0,
      rtl: isRtl(),
    }
  );
});

router.isReady().then(() => app.mount("#app"));
