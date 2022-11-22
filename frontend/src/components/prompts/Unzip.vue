<template>
    <div class="card floating">
      <div class="card-title">
        <h2>{{ $t("prompts.unzip") }}</h2>
      </div>
  
      <div class="card-content">
        <file-list @update:selected="(val) => (dest = val)"></file-list>
      </div>
  
      <div class="card-action">
        <button
          class="button button--flat button--grey"
          @click="$store.commit('closeHovers')"
          :aria-label="$t('buttons.cancel')"
          :title="$t('buttons.cancel')"
        >
          {{ $t("buttons.cancel") }}
        </button>
        <button
          class="button button--flat"
          @click="unzip"
          :disabled="$route.path === dest"
          :aria-label="$t('buttons.unzip')"
          :title="$t('buttons.unzip')"
        >
          {{ $t("buttons.unzip") }}
        </button>
      </div>
    </div>
  </template>
  
  <script>
  import { mapState } from "vuex";
  import FileList from "./FileList";
  import { files as api } from "@/api";
  import buttons from "@/utils/buttons";
  
  export default {
    name: "unzip",
    components: { FileList },
    data: function () {
      return {
        current: window.location.pathname,
        dest: null,
      };
    },
    computed: mapState(["req", "selected"]),
    methods: {
      unzip: async function (event) {
        event.preventDefault();
        let items = [];
  
        for (let item of this.selected) {
          items.push({
            from: this.req.items[item].url,
            to: this.dest,
            name: this.req.items[item].name,
          });
        }
  
        let action = async () => {
          buttons.loading("unzip");
  
          await api
            .unzip(items)
            .then(() => {
              buttons.success("unzip");
              this.$router.push({ path: this.dest });
            })
            .catch((e) => {
              buttons.done("unzip");
              this.$showError(e);
            });
        };
        action();
      },
    },
  };
  </script>