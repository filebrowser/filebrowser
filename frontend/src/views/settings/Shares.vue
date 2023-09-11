<template>
  <errors v-if="error" :errorCode="error.status" />
  <div class="row" v-else-if="!layoutStore.loading">
    <div class="column">
      <div class="card">
        <div class="card-title">
          <h2>{{ t("settings.shareManagement") }}</h2>
        </div>

        <div class="card-content full" v-if="links.length > 0">
          <table>
            <tr>
              <th>{{ t("settings.path") }}</th>
              <th>{{ t("settings.shareDuration") }}</th>
              <th v-if="authStore.user?.perm.admin">
                {{ t("settings.username") }}
              </th>
              <th></th>
              <th></th>
            </tr>

            <tr v-for="link in links" :key="link.hash">
              <td>
                <a :href="buildLink(link)" target="_blank">{{ link.path }}</a>
              </td>
              <td>
                <template v-if="link.expire !== 0">{{
                  humanTime(link.expire)
                }}</template>
                <template v-else>{{ t("permanent") }}</template>
              </td>
              <td v-if="authStore.user?.perm.admin">{{ link.username }}</td>
              <td class="small">
                <button
                  class="action"
                  @click="deleteLink($event, link)"
                  :aria-label="t('buttons.delete')"
                  :title="t('buttons.delete')"
                >
                  <i class="material-icons">delete</i>
                </button>
              </td>
              <td class="small">
                <button
                  class="action copy-clipboard"
                  :data-clipboard-text="buildLink(link)"
                  :aria-label="t('buttons.copyToClipboard')"
                  :title="t('buttons.copyToClipboard')"
                >
                  <i class="material-icons">content_paste</i>
                </button>
              </td>
            </tr>
          </table>
        </div>
        <h2 class="message" v-else>
          <i class="material-icons">sentiment_dissatisfied</i>
          <span>{{ t("files.lonely") }}</span>
        </h2>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import { share as api, users } from "@/api";
import dayjs from "dayjs";
import Clipboard from "clipboard";
import Errors from "@/views/Errors.vue";
import { inject, onBeforeUnmount, ref, onMounted } from "vue";
import { useI18n } from "vue-i18n";

const $showError = inject("$showError") as IToastSuccess;
const $showSuccess = inject("$showSuccess") as IToastError;
const { t } = useI18n();

const layoutStore = useLayoutStore();
const authStore = useAuthStore();

// ...mapState(useAuthStore, ["user"]),
// ...mapWritableState(useLayoutStore, ["loading"]),
const error = ref<any>(null);
const links = ref<any[]>([]);
const clip = ref<Clipboard | null>(null);

onMounted(async () => {
  layoutStore.loading = true;

  try {
    let newLinks = await api.list();
    if (authStore.user?.perm.admin) {
      let userMap = new Map();
      for (let user of await users.getAll())
        userMap.set(user.id, user.username);
      for (let link of newLinks)
        link.username = userMap.has(link.userID)
          ? userMap.get(link.userID)
          : "";
    }
    links.value = newLinks;
  } catch (e: any) {
    error.value = e;
  } finally {
    layoutStore.loading = false;
  }
  clip.value = new Clipboard(".copy-clipboard");
  clip.value.on("success", () => {
    $showSuccess(t("success.linkCopied"));
  });
});

onBeforeUnmount(() => clip.value?.destroy());

const deleteLink = async (event: Event, link: any) => {
  event.preventDefault();

  layoutStore.showHover({
    prompt: "share-delete",
    confirm: () => {
      layoutStore.closeHovers();

      try {
        api.remove(link.hash);
        links.value = links.value.filter((item) => item.hash !== link.hash);
        $showSuccess(t("settings.shareDeleted"));
      } catch (e: any) {
        $showError(e);
      }
    },
  });
};
const humanTime = (time: number) => {
  return dayjs(time * 1000).fromNow();
};

const buildLink = (share: Share) => {
  return api.getShareURL(share);
};
</script>
