<template>
  <div id="quota">
    <div>
      <div class="quota-label">{{ t("sidebar.quota.space") }}</div>

      <br />

      <progress-bar
        :val="spaceProgress"
        size="small"
        :text="spaceProgress + '%'"
      ></progress-bar>

      <div v-if="loaded" class="quota-metric">{{ spaceUsageTitle }}</div>
    </div>

    <br />

    <div>
      <div class="quota-label">{{ t("sidebar.quota.inodes") }}</div>

      <br />

      <progress-bar
        :val="inodeProgress"
        size="small"
        :text="inodeProgress + '%'"
      ></progress-bar>

      <div v-if="loaded" class="quota-metric">{{ inodeUsageTitle }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useQuotaStore } from "@/stores/quota";
import { filesize } from "@/utils";
import { computed, onMounted } from "vue";
import { useI18n } from "vue-i18n";
import ProgressBar from "@/components/ProgressBar.vue";

const quotaStore = useQuotaStore();

const { t } = useI18n();

const loaded = computed(() =>
  quotaStore.quota
    ? quotaStore.quota.inodes !== null && quotaStore.quota.space !== null
    : false
);

const spaceProgress = computed(() =>
  quotaStore.quota && quotaStore.quota.space !== null
    ? progress(quotaStore.quota.space)
    : 0
);

const inodeProgress = computed(() =>
  quotaStore.quota && quotaStore.quota.inodes !== null
    ? progress(quotaStore.quota.inodes)
    : 0
);

const spaceUsageTitle = computed(() =>
  !quotaStore.quota
    ? "- / -"
    : filesize(quotaStore.quota.space.usage) +
      " / " +
      filesize(quotaStore.quota.space.quota)
);

const inodeUsageTitle = computed(() =>
  !quotaStore.quota
    ? "- / -"
    : quotaStore.quota.inodes.usage +
      " / " +
      quotaStore.quota.inodes.quota
);

const progress = (info: QuotaInfo) => {
  let prc = (info.usage / info.quota) * 100;
  prc = Math.round((prc + Number.EPSILON) * 100) / 100;
  return Math.min(prc, 100);
};

onMounted(() => {
  quotaStore.fetchQuota();
});
</script>
