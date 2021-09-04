import Vue from "vue";
import VueI18n from "vue-i18n";


import en from "./en.json";


Vue.use(VueI18n);

export function detectLocale() {
    return "en";
}

// eslint-disable-next-line no-unused-vars
const removeEmpty = (obj) =>
    Object.keys(obj)
        .filter((k) => obj[k] !== null && obj[k] !== undefined && obj[k] !== "") // Remove undef. and null and empty.string.
        .reduce(
            (newObj, k) =>
                typeof obj[k] === "object"
                    ? Object.assign(newObj, {[k]: removeEmpty(obj[k])}) // Recurse.
                    : Object.assign(newObj, {[k]: obj[k]}), // Copy value.
            {}
        );

const i18n = new VueI18n({
    locale: detectLocale(),
    fallbackLocale: "en",
    messages: {
        en: en,
    },
});

export default i18n;
