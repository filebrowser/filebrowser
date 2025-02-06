import dayjs from "dayjs";
import { createI18n } from "vue-i18n";

import("dayjs/locale/ar");
import("dayjs/locale/en");
import("dayjs/locale/es");
import("dayjs/locale/fr");
import("dayjs/locale/id");
import("dayjs/locale/lt");
import("dayjs/locale/pt-br");
import("dayjs/locale/pt");
import("dayjs/locale/ru");
import("dayjs/locale/tr");
import("dayjs/locale/uk");
import("dayjs/locale/zh-cn");
import("dayjs/locale/zh");

// All i18n resources specified in the plugin `include` option can be loaded
// at once using the import syntax
import messages from "@intlify/unplugin-vue-i18n/messages";

export function detectLocale() {
  // locale is an RFC 5646 language tag
  // https://developer.mozilla.org/en-US/docs/Web/API/Navigator/language
  let locale = navigator.language.toLowerCase();
  switch (true) {
    case /^ar.*/i.test(locale):
      locale = "ar_AR";
      break;
    case /^en.*/i.test(locale):
      locale = "en_GB";
      break;
    case /^es-AR.*/i.test(locale):
      locale = "es_AR";
      break;
    case /^es-CO.*/i.test(locale):
      locale = "es_CO";
      break;
    case /^es-MX.*/i.test(locale):
      locale = "es_MX";
      break;
    case /^es.*/i.test(locale):
      locale = "es_ES";
      break;
    case /^fr.*/i.test(locale):
      locale = "fr_FR";
      break;
    case /^id.*/i.test(locale):
      locale = "id_ID";
      break;
    case /^lt.*/i.test(locale):
      locale = "lt_LT";
      break;
    case /^pt-BR.*/i.test(locale):
      locale = "pt_BR";
      break;
    case /^pt.*/i.test(locale):
      locale = "pt_PT";
      break;
    case /^ru.*/i.test(locale):
      locale = "ru_RU";
      break;
    case /^tr.*/i.test(locale):
      locale = "tr_TR";
      break;
    case /^uk.*/i.test(locale):
      locale = "uk_UA";
      break;
    case /^zh.*/i.test(locale):
      locale = "zh_CN";
      break;
    default:
      locale = "en_GB";
  }

  return locale;
}

// TODO: was this really necessary?
// function removeEmpty(obj: Record<string, any>): void {
//   Object.keys(obj)
//     .filter((k) => obj[k] !== null && obj[k] !== undefined && obj[k] !== "") // Remove undef. and null and empty.string.
//     .reduce(
//       (newObj, k) =>
//         typeof obj[k] === "object"
//           ? Object.assign(newObj, { [k]: removeEmpty(obj[k]) }) // Recurse.
//           : Object.assign(newObj, { [k]: obj[k] }), // Copy value.
//       {}
//     );
// }

export const rtlLanguages = ["ar_AR"];

export const i18n = createI18n({
  locale: detectLocale(),
  fallbackLocale: "en_GB",
  messages,
  // expose i18n.global for outside components
  legacy: true,
});

export const isRtl = (locale?: string) => {
  // see below
  // @ts-ignore
  return rtlLanguages.includes(locale || i18n.global.locale.value);
};

export function setLocale(locale: string) {
  let normalizedLocale = locale;
  if (locale.includes("_")) {
    normalizedLocale = locale.split("_")[0];
  }

  dayjs.locale(normalizedLocale);
  // according to doc u only need .value if legacy: false but they lied
  // https://vue-i18n.intlify.dev/guide/essentials/scope.html#local-scope-1
  //@ts-ignore
  i18n.global.locale.value = locale;
}

export function setHtmlLocale(locale: string) {
  const html = document.documentElement;
  html.lang = locale;
  if (isRtl(locale)) html.dir = "rtl";
  else html.dir = "ltr";
}

export default i18n;
