import { defineStore } from "pinia";
import { quota as api } from "@/api";
import { quotaExists } from "@/utils/constants";

export const useQuotaStore = defineStore("quota", {
  state: (): {
    quota: IQuota | null;
  } => ({
    quota: null,
  }),
  actions: {
    async fetchQuota(delay: number = 0) {
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
            this.quota = data;
          }
        } catch (e) {
          console.log(e);
        }
      }, delay);
    }
  },
});
