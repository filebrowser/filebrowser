import * as i18n from "@/i18n";
import moment from "moment";

const mutations = {
  closeHovers: (state) => {
    state.show = null;
    state.showConfirm = null;
  },
  toggleShell: (state) => {
    state.showShell = !state.showShell;
  },
  showHover: (state, value) => {
    if (typeof value !== "object") {
      state.show = value;
      return;
    }

    state.show = value.prompt;
    state.showConfirm = value.confirm;
  },
  showError: (state) => {
    state.show = "error";
  },
  showSuccess: (state) => {
    state.show = "success";
  },
  setLoading: (state, value) => {
    state.loading = value;
  },
  setReload: (state, value) => {
    state.reload = value;
  },
  setUser: (state, value) => {
    if (value === null) {
      state.user = null;
      return;
    }

    let locale = value.locale;

    if (locale === "") {
      locale = i18n.detectLocale();
    }

    moment.locale(locale);
    i18n.default.locale = locale;
    state.user = value;
  },
  setJWT: (state, value) => (state.jwt = value),
  multiple: (state, value) => (state.multiple = value),
  addSelected: (state, value) => state.selected.push(value),
  removeSelected: (state, value) => {
    let i = state.selected.indexOf(value);
    if (i === -1) return;
    state.selected.splice(i, 1);
  },
  resetSelected: (state) => {
    state.selected = [];
  },
  updateUser: (state, value) => {
    if (typeof value !== "object") return;

    for (let field in value) {
      if (field === "locale") {
        moment.locale(value[field]);
        i18n.default.locale = value[field];
      }

      state.user[field] = value[field];
    }
  },
  updateRequest: (state, value) => {
    state.oldReq = state.req;
    state.req = value;
  },
  updateClipboard: (state, value) => {
    state.clipboard.key = value.key;
    state.clipboard.items = value.items;
    state.clipboard.path = value.path;
  },
  resetClipboard: (state) => {
    state.clipboard.key = "";
    state.clipboard.items = [];
  },
  showContextMenu: (state, value) => {
    state.contextMenu = {
      x: value.x,
      y: value.y,
    };
  },
  hideContextMenu: (state) => {
    state.contextMenu = null;
  },
  addDiskUsage: (state, value) => {
    if (value.path[value.path.length - 1] == "/") {
      value.path = value.path.slice(0, -1);
    }

    let tmp = state.diskUsages;
    state.diskUsages = null;
    tmp[value.path] = value.usage;
    state.diskUsages = tmp;
  },
};

export default mutations;
