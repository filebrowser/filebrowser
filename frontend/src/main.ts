import { disableExternal } from "@/utils/constants";
import { createApp } from "vue";
import VueNumberInput from "@chenfengyuan/vue-number-input";
import VueLazyload from "vue-lazyload";
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

const app = createApp(App);

app.component(VueNumberInput.name || "vue-number-input", VueNumberInput);
app.use(VueLazyload);
app.use(Toast, {
  transition: "Vue-Toastification__bounce",
  maxToasts: 10,
  newestOnTop: true,
} satisfies PluginOptions);

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

const normalizeErrorToast = (error: Error | string, displayReport: boolean) => {
  let message = String((error as Error).message || error || "");

  const statusWrappedMessage = message.match(
    /^\s*\d{3}\s+[^()]*\(([\s\S]*Security scan blocked the upload[\s\S]*)\)\s*$/
  );
  if (statusWrappedMessage?.[1]) {
    message = statusWrappedMessage[1].trim();
  }

  const isSecurityScanBlock = message.includes(
    "Security scan blocked the upload"
  );

  return {
    message,
    isReport: !disableExternal && displayReport && !isSecurityScanBlock,
  };
};

app.provide("$showError", (error: Error | string, displayReport = true) => {
  const $toast = useToast();
  const normalized = normalizeErrorToast(error, displayReport);

  $toast.error(
    {
      component: CustomToast,
      props: {
        message: normalized.message,
        isReport: normalized.isReport,
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
