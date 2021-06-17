import Vue from "vue";
import VueI18n from "vue-i18n";

import arAR from "./ar_AR.json";
import enGB from "./en_GB.json";
import esAR from "./es_AR.json";
import esCO from "./es_CO.json";
import esES from "./es_ES.json";
import esMX from "./es_MX.json";
import frFR from "./fr_FR.json";
import idID from "./id_ID.json";
import ltLT from "./lt_LT.json";
import ptBR from "./pt_BR.json";
import ptPT from "./pt_BR.json";
import ruRU from "./ru_RU.json";
import trTR from "./tr_TR.json";
import ukUA from "./uk_UA.json";
import zhCN from "./zh_CN.json";

Vue.use(VueI18n);

export function detectLocale() {
  let locale = (navigator.language || navigator.browserLanguage).toLowerCase();
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

const removeEmpty = (obj) =>
  Object.keys(obj)
    .filter((k) => obj[k] !== null && obj[k] !== undefined && obj[k] !== "") // Remove undef. and null and empty.string.
    .reduce(
      (newObj, k) =>
        typeof obj[k] === "object"
          ? Object.assign(newObj, { [k]: removeEmpty(obj[k]) }) // Recurse.
          : Object.assign(newObj, { [k]: obj[k] }), // Copy value.
      {}
    );

const i18n = new VueI18n({
  locale: detectLocale(),
  fallbackLocale: "en_GB",
  messages: {
    ar_AR: removeEmpty(arAR),
    en_GB: enGB,
    es_AR: removeEmpty(esAR),
    es_CO: removeEmpty(esCO),
    es_ES: removeEmpty(esES),
    es_MX: removeEmpty(esMX),
    fr_FR: removeEmpty(frFR),
    id_ID: removeEmpty(idID),
    lt_LT: removeEmpty(ltLT),
    pt_BR: removeEmpty(ptBR),
    pt_PT: removeEmpty(ptPT),
    ru_RU: removeEmpty(ruRU),
    tr_TR: removeEmpty(trTR),
    uk_UA: removeEmpty(ukUA),
    zh_CN: removeEmpty(zhCN),
  },
});

export default i18n;
