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

          <!-- Machine list — one editable card per configured machine.
               Order matters: machines[0] is the default destination
               for any /api/cnc/* call without a machine_id. -->
          <div
            v-for="(machine, idx) in settings.machines"
            :key="machine.id || idx"
            class="machine-row"
          >
            <div class="machine-row__header">
              <h3>
                <span class="machine-row__index">{{ idx + 1 }}.</span>
                <input
                  class="input machine-row__name"
                  type="text"
                  v-model="machine.name"
                  :placeholder="t('settings.machineNamePlaceholder')"
                />
              </h3>
              <button
                v-if="settings.machines.length > 1"
                type="button"
                class="button button--flat machine-row__delete"
                @click="deleteMachine(idx)"
              >
                <i class="material-icons">delete</i>
                {{ t("settings.machineDelete") }}
              </button>
            </div>

            <p>
              <label class="small">{{ t("settings.machineBrand") }}</label>
              <select class="input input--block" v-model="machine.brand">
                <option value="haas">{{ t("settings.machineBrandHaas") }}</option>
                <option value="fanuc" disabled>
                  {{ t("settings.machineBrandFanuc") }}
                </option>
                <option value="mazak" disabled>
                  {{ t("settings.machineBrandMazak") }}
                </option>
                <option value="okuma" disabled>
                  {{ t("settings.machineBrandOkuma") }}
                </option>
                <option value="generic" disabled>
                  {{ t("settings.machineBrandGeneric") }}
                </option>
              </select>
              <span class="small machine-row__hint">
                {{ t("settings.machineBrandHelp") }}
              </span>
            </p>

            <p>
              <label class="small">{{ t("settings.machineHaasHost") }}</label>
              <input
                class="input input--block"
                type="text"
                placeholder="192.168.20.200"
                v-model="machine.host"
              />
            </p>

            <p>
              <label class="small">{{ t("settings.machineHaasPort") }}</label>
              <vue-number-input
                controls
                v-model.number="machine.port"
                :min="1"
                :max="65535"
              />
            </p>

            <p>
              <label class="small">{{ t("settings.machineToolSlots") }}</label>
              <vue-number-input
                controls
                v-model.number="machine.toolSlots"
                :min="0"
                :max="200"
              />
              <span class="small machine-row__hint">
                {{ t("settings.machineToolSlotsHelp") }}
              </span>
            </p>

            <p>
              <label class="small">
                <input
                  type="checkbox"
                  v-model="machine.requirePreflight"
                />
                {{ t("settings.machineRequirePreflight") }}
              </label>
              <span class="small machine-row__hint">
                {{ t("settings.machineRequirePreflightHelp") }}
              </span>
            </p>

            <p>
              <label class="small">{{ t("settings.machineCameraType") }}</label>
              <select class="input input--block" v-model="machine.cameraType">
                <option value="auto">{{ t("settings.machineCameraTypeAuto") }}</option>
                <option value="hls">{{ t("settings.machineCameraTypeHls") }}</option>
                <option value="mjpeg">{{ t("settings.machineCameraTypeMjpeg") }}</option>
                <option value="iframe">{{ t("settings.machineCameraTypeIframe") }}</option>
                <option value="none">{{ t("settings.machineCameraTypeNone") }}</option>
              </select>
              <span class="small machine-row__hint">
                {{ t("settings.machineCameraTypeHelp") }}
              </span>
            </p>

            <p>
              <label class="small">{{ t("settings.machineCameraUrl") }}</label>
              <input
                class="input input--block"
                type="text"
                :placeholder="cameraUrlPlaceholder(machine.cameraType)"
                v-model="machine.cameraUrl"
              />
            </p>

            <p>
              <button
                type="button"
                class="button button--flat"
                :disabled="probingId !== null"
                @click="runToolProbe(machine.id)"
              >
                {{ probingId === machine.id ? t("settings.toolProbing") : t("settings.toolProbeRun") }}
              </button>
            </p>
            <div
              v-if="probeResults[machine.id]"
              class="probe-report"
              :class="`probe-report--${probeResultClass(machine.id)}`"
            >
              <p>
                <strong>{{ t("settings.toolProbeVerdict") }}:</strong>
                {{ probeResults[machine.id].verdict }}
              </p>
              <p class="small">{{ probeResults[machine.id].recommendation }}</p>
              <p class="small">
                {{ t("settings.toolProbeMeta", {
                  slots: probeResults[machine.id].slots_probed,
                  ms: Math.round(probeResults[machine.id].duration_ms),
                  addr: probeResults[machine.id].bridge_address,
                }) }}
              </p>
            </div>
          </div>

          <p>
            <button
              type="button"
              class="button button--flat"
              @click="addMachine"
            >
              <i class="material-icons">add</i>
              {{ t("settings.machineAdd") }}
            </button>
          </p>

          <!-- Bearer token is install-wide, not per-machine — same
               token authenticates S2S calls against any machine_id -->
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
import type { CncMachine, CncSettings, ProbeToolsReport } from "@/api/cnc";
import { fetchURL } from "@/api/utils";
import { StatusError } from "@/api/utils";
import { useLayoutStore } from "@/stores/layout";
import Errors from "@/views/Errors.vue";
import { inject, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

const error = ref<StatusError | null>(null);
const settings = ref<CncSettings | null>(null);

// Probe is per-machine — keep results keyed by machine ID so each
// row's panel renders only its own outcome.
const probingId = ref<string | null>(null);
const probeResults = ref<Record<string, ProbeToolsReport>>({});

const probeResultClass = (id: string) => {
  const r = probeResults.value[id];
  if (!r) return "";
  switch (r.verdict) {
    case "ngc-mapping-confirmed":
      return "ok";
    case "ngc-mapping-empty":
      return "warn";
    default:
      return "err";
  }
};

const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;

const { t } = useI18n();
const layoutStore = useLayoutStore();

const newID = (): string => {
  // Client-side temp ID; server replaces with a stable one on save.
  // Random base36 string, 11 chars — collisions impossible at our scale.
  return "new-" + Math.random().toString(36).slice(2, 12);
};

const addMachine = () => {
  if (!settings.value) return;
  const idx = settings.value.machines.length + 1;
  settings.value.machines.push({
    id: newID(),
    name: `Machine ${idx}`,
    brand: "haas",
    host: "",
    port: 4196,
    toolSlots: 30,
    cameraUrl: "",
    cameraType: "auto",
    requirePreflight: false,
  });
};

const cameraUrlPlaceholder = (kind: string | undefined) => {
  switch (kind) {
    case "hls":
      return "https://… .m3u8";
    case "mjpeg":
      return "https://camera.local/snapshot or .jpg";
    case "iframe":
      return "https://protect.local/protect/livev3/<id> (UniFi Live View URL)";
    case "none":
      return "—";
    default:
      return "https://… .m3u8 (HLS) or .jpg / /snapshot (MJPEG)";
  }
};

const deleteMachine = (idx: number) => {
  if (!settings.value) return;
  if (settings.value.machines.length <= 1) return; // never let it go empty
  settings.value.machines.splice(idx, 1);
};

const save = async () => {
  if (!settings.value) return;
  if (!settings.value.machines.length) {
    $showError(new Error(t("settings.machineNoneError")) as any);
    return;
  }
  // Send only what the server expects; strip client-side temp IDs
  // — the backend mints a stable ID for any `id` it doesn't
  // recognize (or that starts with "new-").
  const cleaned = settings.value.machines.map((m) => ({
    ...m,
    id: m.id?.startsWith("new-") ? "" : m.id,
  }));
  try {
    await fetchURL(`/api/cnc/settings`, {
      method: "PUT",
      body: JSON.stringify({ machines: cleaned }),
    });
    $showSuccess(t("settings.settingsUpdated"));
    // Refetch so the user sees server-assigned IDs.
    settings.value = await api.getSettings();
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

const runToolProbe = async (machineId: string) => {
  probingId.value = machineId;
  try {
    probeResults.value = {
      ...probeResults.value,
      [machineId]: await api.probeTools(30, machineId),
    };
  } catch (e: any) {
    $showError(e);
  } finally {
    probingId.value = null;
  }
};

onMounted(async () => {
  try {
    layoutStore.loading = true;
    const fetched = await api.getSettings();
    // Defensive: backend EnsureMigrated should always produce
    // machines[0], but if a brand-new install boots with truly empty
    // settings we synthesise one so the UI has a row to edit.
    if (!fetched.machines || fetched.machines.length === 0) {
      fetched.machines = [
        {
          id: newID(),
          name: "Machine 1",
          brand: "haas",
          host: "",
          port: 4196,
          toolSlots: 30,
          cameraUrl: "",
          cameraType: "auto",
          requirePreflight: false,
        },
      ];
    }
    // Older payloads may be missing brand/cameraType/cameraUrl/toolSlots
    // — coerce to today's defaults so every input has a value.
    fetched.machines = fetched.machines.map((m: CncMachine) => ({
      brand: "haas",
      toolSlots: 30,
      cameraUrl: "",
      cameraType: "auto",
      requirePreflight: false,
      ...m,
    }));
    settings.value = fetched;
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
.machine-row {
  border: 1px solid var(--border-color, #ccc);
  border-radius: 6px;
  padding: 0.8rem 1rem;
  margin-bottom: 1rem;
  background: var(--alt-background, #fafafa);
}

.machine-row__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.6rem;
  margin-bottom: 0.6rem;
}

.machine-row__header h3 {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  margin: 0;
  font-size: 1.1rem;
  flex: 1;
}

.machine-row__index {
  color: var(--fg-muted, #888);
  font-weight: 400;
}

.machine-row__name {
  flex: 1;
  font-size: 1rem;
  padding: 0.3rem 0.6rem;
}

.machine-row__hint {
  display: block;
  margin-top: 0.2rem;
  color: var(--fg-muted, #888);
}

.machine-row__delete {
  color: #c62828;
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
}

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
</style>
