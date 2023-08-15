import "whatwg-fetch";
import cssVars from "css-vars-ponyfill";
import { createApp, configureCompat } from "vue";
import store from "@/store";
import router from "@/router";
import i18n from "@/i18n";
import { recaptcha, loginPage } from "@/utils/constants";
import { login, validateLogin } from "@/utils/auth";
import App from "@/App.vue";

cssVars();

configureCompat({
  MODE: 2,
});

const app = createApp(App);

app.use(store);
app.use(router);
app.use(i18n);

async function start() {
  try {
    if (loginPage) {
      await validateLogin();
    } else {
      await login("", "", "");
    }
  } catch (e) {
    console.log(e);
  }

  if (recaptcha) {
    await new Promise((resolve) => {
      const check = () => {
        if (typeof window.grecaptcha === "undefined") {
          setTimeout(check, 100);
        } else {
          resolve();
        }
      };

      check();
    });
  }

  router.isReady().then(() => app.mount("#app"));
}

start();
