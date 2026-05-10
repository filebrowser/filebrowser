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

          <h3>{{ t("settings.toolProbe") }}</h3>
          <p class="small">{{ t("settings.toolProbeHelp") }}</p>
          <p>
            <button
              type="button"
              class="button button--flat"
              :disabled="probing"
              @click="runToolProbe"
            >
              {{ probing ? t("settings.toolProbing") : t("settings.toolProbeRun") }}
            </button>
          </p>
          <div v-if="probeResult" class="probe-report" :class="`probe-report--${probeResultClass}`">
            <p>
              <strong>{{ t("settings.toolProbeVerdict") }}:</strong>
              {{ probeResult.verdict }}
            </p>
            <p class="small">{{ probeResult.recommendation }}</p>
            <p class="small">
              {{ t("settings.toolProbeMeta", {
                slots: probeResult.slots_probed,
                ms: Math.round(probeResult.duration_ms),
                addr: probeResult.bridge_address,
              }) }}
            </p>
            <table class="probe-report__table">
              <thead>
                <tr>
                  <th>Base</th>
                  <th>Label</th>
                  <th>OK</th>
                  <th>Empty</th>
                  <th>Err</th>
                  <th>First sample</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="b in probeResult.bases" :key="b.base">
                  <td><code>{{ b.base }}</code></td>
                  <td>{{ b.label }}</td>
                  <td>{{ b.ok }}</td>
                  <td>{{ b.empty }}</td>
                  <td>{{ b.errors }}</td>
                  <td>
                    <code v-if="b.samples[0]">
                      {{ b.samples[0].value || b.samples[0].error || "—" }}
                    </code>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
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
import type { CncSettings, ProbeToolsReport } from "@/api/cnc";
import { StatusError } from "@/api/utils";
import { useLayoutStore } from "@/stores/layout";
import Errors from "@/views/Errors.vue";
import { computed, inject, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

const error = ref<StatusError | null>(null);
const settings = ref<CncSettings | null>(null);
const probing = ref(false);
const probeResult = ref<ProbeToolsReport | null>(null);

const probeResultClass = computed(() => {
  if (!probeResult.value) return "";
  switch (probeResult.value.verdict) {
    case "ngc-mapping-confirmed":
      return "ok";
    case "ngc-mapping-empty":
      return "warn";
    default:
      return "err";
  }
});

const runToolProbe = async () => {
  probing.value = true;
  try {
    probeResult.value = await api.probeTools(30);
  } catch (e: any) {
    $showError(e);
  } finally {
    probing.value = false;
  }
};

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

<style scoped>
.probe-report {
  margin-top: 0.6rem;
  padding: 0.6rem 0.8rem;
  border-radius: 6px;
  border: 1px solid #ccc;
}

.probe-report--ok {
  border-color: #2e7d32;
  background: rgba(46, 125, 50, 0.08);
}

.probe-report--warn {
  border-color: #ed6c02;
  background: rgba(237, 108, 2, 0.08);
}

.probe-report--err {
  border-color: #c62828;
  background: rgba(198, 40, 40, 0.08);
}

.probe-report__table {
  width: 100%;
  margin-top: 0.5rem;
  border-collapse: collapse;
  font-size: 0.85rem;
  font-variant-numeric: tabular-nums;
}

.probe-report__table th,
.probe-report__table td {
  padding: 0.25rem 0.5rem;
  text-align: left;
  border-bottom: 1px solid rgba(0, 0, 0, 0.08);
}

.probe-report__table code {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 0.8rem;
}
</style>
