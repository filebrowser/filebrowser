import "whatwg-fetch";
import cssVars from "css-vars-ponyfill";
import { createApp } from "vue";
import VueLazyload from "vue-lazyload";
import createPinia from "@/stores";
import router from "@/router";
import i18n from "@/i18n";
import App from "@/App.vue";

cssVars();

const pinia = createPinia(router);

const app = createApp(App);

app.use(VueLazyload);
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

router.isReady().then(() => app.mount("#app"));
