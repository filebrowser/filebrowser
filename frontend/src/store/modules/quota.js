import { users as api } from "@/api";
import { quotaExists } from "@/utils/constants";

const state = {
  inodes: null,
  space: null,
};

const mutations = {
  setQuota(state, { inodes, space }) {
    state.inodes = inodes;
    state.space = space;
  },
};

const actions = {
  fetch: async (context, delay = 0) => {
    if (!quotaExists) {
      return;
    }

    setTimeout(async () => {
      try {
        let data = await api.getQuota();
        if (
          data !== null &&
          data.inodes != undefined &&
          data.space != undefined
        ) {
          context.commit("setQuota", data);
        }
      } catch (e) {
        console.log(e);
      }
    }, delay);
  },
};

export default { state, mutations, actions, namespaced: true };
