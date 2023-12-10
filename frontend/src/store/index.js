import Vue from "vue";
import Vuex from "vuex";
import mutations from "./mutations";
import getters from "./getters";
import upload from "./modules/upload";

Vue.use(Vuex);

const state = {
  user: null,
  req: {},
  oldReq: {},
  clipboard: {
    key: "",
    items: [],
  },
  jwt: "",
  progress: 0,
  loading: false,
  reload: false,
  selected: [],
  multiple: false,
  prompts: [],
  showShell: false,
  scrollPosition: 0,
  showLimit: 50,
};

export default new Vuex.Store({
  strict: true,
  state,
  getters,
  mutations,
  modules: { upload },
});
