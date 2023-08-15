import { createStore } from "vuex";
import mutations from "./mutations";
import getters from "./getters";
import upload from "./modules/upload";
import router from "@/router";

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
  show: null,
  showShell: false,
  showConfirm: null,
  showAction: null,
  get route() {
    return router.currentRoute.value;
  },
};

const store = createStore({
  strict: true,
  state,
  getters,
  mutations,
  modules: { upload },
});

export default store;
