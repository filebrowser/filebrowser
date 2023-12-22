import * as i18n from "@/i18n";
import moment from "moment";

const mutations = {
  closeHovers: (state) => {
    state.prompts.pop();
  },
  toggleShell: (state) => {
    state.show = null;
    state.showShell = !state.showShell;
  },
  showHover: (state, value) => {
    if (typeof value !== "object") {
      state.prompts.push({
        prompt: value,
        confirm: null,
        action: null,
        props: null,
      });
      return;
    }

    state.prompts.push({
      prompt: value.prompt, // Should not be null
      confirm: value?.confirm,
      action: value?.action,
      props: value?.props,
    });
  },
  showError: (state) => {
    state.prompts.push("error");
  },
  showSuccess: (state) => {
    state.prompts.push("success");
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
    const selectedItems = state.selected.map((i) => state.req.items[i]);
    state.oldReq = state.req;
    state.req = value;
    state.selected = [];

    if (!state.req?.items) return;
    state.selected = state.req.items
      .filter((item) => selectedItems.some((rItem) => rItem.url === item.url))
      .map((item) => item.index);
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
  setUploadSpeed: (state, value) => {
    state.upload.speedMbyte = value;
  },
  setETA(state, value) {
    state.upload.eta = value;
  },
  resetUpload(state) {
    state.upload.uploads = {};
    state.upload.queue = [];
    state.upload.progress = [];
    state.upload.sizes = [];
    state.upload.id = 0;
    state.upload.speedMbyte = 0;
    state.upload.eta = 0;
  },
};

export default mutations;
