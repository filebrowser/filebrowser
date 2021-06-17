<template>
  <select v-on:change="change" :value="locale">
    <option v-for="(language, value) in locales" :key="value" :value="value">
      {{ $t("languages." + language) }}
    </option>
  </select>
</template>

<script>
export default {
  name: "languages",
  props: ["locale"],
  data() {
    let dataObj = {
      locales: {
        ar_AR: "arAR",
        en_GB: "enGB",
        es_AR: "esAR",
        es_CO: "esCO",
        es_ES: "esES",
        es_MX: "esMX",
        fr_FR: "frFR",
        id_ID: "idID",
        lt_LT: "ltLT",
        pt_BR: "ptBR",
        pt_PT: "ptPT",
        ru_RU: "ruRU",
        tr_TR: "trTR",
        uk_UA: "ukUA",
        zh_CN: "zhCN",
      },
    };

    Object.defineProperty(dataObj, "locales", {
      configurable: false,
      writable: false,
    });

    return dataObj;
  },
  methods: {
    change(event) {
      this.$emit("update:locale", event.target.value);
    },
  },
};
</script>
