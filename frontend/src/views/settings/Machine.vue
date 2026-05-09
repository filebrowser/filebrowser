<template>
  <errors v-if="error" :errorCode="error.status" />
  <div class="row" v-else-if="!layoutStore.loading && settings !== null">
    <div class="column">
      <form class="card" @submit.prevent="save">
        <div class="card-title">
          <h2>{{ t("settings.machineSettings") }}</h2>
        </div>

        <div class="card-content">
          <p class="small">{{ t("settings.machineSettingsHelp") }}</p>

          <h3>{{ t("settings.machineHaasBridge") }}</h3>
          <p>
            <label class="small">{{ t("settings.machineHaasHost") }}</label>
            <input
              class="input input--block"
              type="text"
              placeholder="192.168.20.200"
              v-model="settings.haasHost"
            />
          </p>

          <p>
            <label class="small">{{ t("settings.machineHaasPort") }}</label>
            <vue-number-input
              controls
              v-model.number="settings.haasPort"
              :min="1"
              :max="65535"
            />
          </p>

          <h3>{{ t("settings.machineCamera") }}</h3>
          <p>
            <label class="small">{{ t("settings.machineCameraUrl") }}</label>
            <input
              class="input input--block"
              type="text"
              placeholder="https://… .m3u8 (HLS) or .jpg / /snapshot (MJPEG)"
              v-model="settings.cameraUrl"
            />
          </p>

          <h3>{{ t("settings.machineToken") }}</h3>
          <p class="small">{{ t("settings.machineTokenHelp") }}</p>
          <p>
            <input
              class="input input--block"
              type="text"
              readonly
              :value="settings.machineToken || t('settings.machineTokenEmpty')"
            />
          </p>
          <p>
            <button
              type="button"
              class="button button--flat"
              @click="regenerateToken"
            >
              {{ t("settings.machineTokenRegenerate") }}
            </button>
          </p>
        </div>

        <div class="card-action">
          <input
            class="button button--flat"
            type="submit"
            :value="t('buttons.update')"
          />
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { cnc as api } from "@/api";
import type { CncSettings } from "@/api/cnc";
import { StatusError } from "@/api/utils";
import { useLayoutStore } from "@/stores/layout";
import Errors from "@/views/Errors.vue";
import { inject, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

const error = ref<StatusError | null>(null);
const settings = ref<CncSettings | null>(null);

const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;

const { t } = useI18n();
const layoutStore = useLayoutStore();

const save = async () => {
  if (!settings.value) return;
  try {
    await api.updateSettings(settings.value);
    $showSuccess(t("settings.settingsUpdated"));
  } catch (e: any) {
    $showError(e);
  }
};

const regenerateToken = async () => {
  try {
    const r = await api.regenerateToken();
    if (settings.value) settings.value.machineToken = r.machineToken;
    $showSuccess(t("settings.machineTokenRegenerated"));
  } catch (e: any) {
    $showError(e);
  }
};

onMounted(async () => {
  try {
    layoutStore.loading = true;
    settings.value = await api.getSettings();
  } catch (err) {
    if (err instanceof Error) {
      error.value = err as StatusError;
    }
  } finally {
    layoutStore.loading = false;
  }
});
</script>
