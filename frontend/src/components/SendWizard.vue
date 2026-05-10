<template>
  <section class="machine-card send-wizard">
    <div class="card-header">
      <i class="material-icons">send</i>
      {{ t("sendWizard.title") }}
      <span class="card-header__spacer" />
      <button class="link-btn" @click="$emit('cancel')">
        {{ t("sendWizard.cancel") }}
      </button>
    </div>
    <div class="card-body wizard-body">
      <!-- 1. File summary ────────────────────────────────────── -->
      <section class="wizard-section">
        <h3 class="wizard-section__title">{{ t("sendWizard.file") }}</h3>
        <div class="wizard-file">
          <i class="material-icons">description</i>
          <code>{{ filePath }}</code>
        </div>
      </section>

      <!-- 1.5 Destination machine — only when more than one is configured.
           Defaults to the currently-selected machine in the global store
           but the operator can route this specific send to a different
           controller without having to change /machine first. -->
      <section v-if="cncStore.machines.length > 1" class="wizard-section">
        <h3 class="wizard-section__title">
          {{ t("sendWizard.destination") }}
        </h3>
        <select
          class="wizard-destination"
          v-model="destinationId"
        >
          <option v-for="m in cncStore.machines" :key="m.id" :value="m.id">
            {{ m.name || m.id }} <template v-if="m.host">· {{ m.host }}:{{ m.port }}</template>
          </option>
        </select>
      </section>

      <!-- 2. Bridge connection check ──────────────────────────── -->
      <section class="wizard-section">
        <h3 class="wizard-section__title">{{ t("sendWizard.connection") }}</h3>
        <div class="wizard-connection">
          <button
            class="check-btn"
            :disabled="checking"
            @click="runCheck"
          >
            <i class="material-icons">network_check</i>
            {{ checking ? t("machine.checking") : t("machine.checkConnection") }}
          </button>
          <span
            v-if="checkResult"
            class="wizard-pill"
            :class="connectionOK ? 'wizard-pill--ok' : 'wizard-pill--err'"
          >
            <i class="material-icons">
              {{ connectionOK ? "check_circle" : "error" }}
            </i>
            {{ connectionOK ? t("sendWizard.connectionOk") : t("sendWizard.connectionFail") }}
          </span>
          <span v-else class="wizard-pill wizard-pill--unknown">
            {{ t("sendWizard.connectionUnknown") }}
          </span>
        </div>
        <div v-if="checkResult" class="wizard-connection__detail">
          <div :class="checkResult.bridge.ok ? 'ok' : 'err'">
            <strong>{{ t("machine.checkBridge") }}:</strong>
            <template v-if="checkResult.bridge.ok">
              {{ checkResult.bridge.address }} · {{ Math.round(checkResult.bridge.latency_ms || 0) }} ms
            </template>
            <template v-else>{{ checkResult.bridge.error || "?" }}</template>
          </div>
          <div :class="checkResult.controller.ok ? 'ok' : 'err'">
            <strong>{{ t("machine.checkController") }}:</strong>
            <template v-if="checkResult.controller.ok">
              Q104 → {{ checkResult.controller.mode }} · {{ Math.round(checkResult.controller.latency_ms || 0) }} ms
            </template>
            <template v-else>{{ checkResult.controller.error || "?" }}</template>
          </div>
        </div>
      </section>

      <!-- 2b. Tool check — NC ↔ tool-table ───────────────────── -->
      <section class="wizard-section">
        <h3 class="wizard-section__title">{{ t("sendWizard.toolCheck") }}</h3>
        <div v-if="preflightLoading" class="wizard-tools__loading">
          {{ t("sendWizard.toolCheckLoading") }}
        </div>
        <div v-else-if="preflightError" class="wizard-tools__err">
          {{ preflightError }}
        </div>
        <div v-else-if="preflight && preflight.table_missing" class="wizard-tools__warning">
          <i class="material-icons">warning</i>
          {{ t("sendWizard.toolCheckNoTable") }}
        </div>
        <template v-else-if="preflight">
          <div
            v-if="preflight.starting_tool !== undefined"
            class="wizard-starting-tool"
            :class="{ 'wizard-starting-tool--swap': preflight.spindle_swap }"
          >
            <i class="material-icons">play_circle</i>
            <span>
              {{ t("sendWizard.startingTool", { n: preflight.starting_tool }) }}
              <template v-if="preflight.current_spindle_tool !== undefined">
                ·
                <template v-if="preflight.spindle_swap">
                  <strong>{{ t("sendWizard.spindleSwap", { from: preflight.current_spindle_tool, to: preflight.starting_tool }) }}</strong>
                </template>
                <template v-else>
                  {{ t("sendWizard.spindleMatch", { n: preflight.current_spindle_tool }) }}
                </template>
              </template>
              <template v-else>
                · <span class="wizard-starting-tool__unknown">{{ t("sendWizard.spindleUnknown") }}</span>
              </template>
            </span>
          </div>
          <div class="wizard-tools__summary">
            <span class="wizard-pill wizard-pill--ok" v-if="preflight.summary.ok > 0">
              {{ preflight.summary.ok }} OK
            </span>
            <span class="wizard-pill wizard-pill--warn" v-if="preflight.summary.warn > 0">
              {{ preflight.summary.warn }} {{ t("sendWizard.warn") }}
            </span>
            <span class="wizard-pill wizard-pill--err" v-if="preflight.summary.empty > 0">
              {{ preflight.summary.empty }} {{ t("sendWizard.toolEmpty") }}
            </span>
            <span class="wizard-pill wizard-pill--err" v-if="preflight.summary.offline > 0">
              {{ preflight.summary.offline }} {{ t("sendWizard.toolOffline") }}
            </span>
            <span class="wizard-pill wizard-pill--err" v-if="preflight.summary.missing > 0">
              {{ preflight.summary.missing }} {{ t("sendWizard.toolMissing") }}
            </span>
            <span v-if="preflight.table_read_at" class="wizard-tools__age">
              {{ t("sendWizard.tableReadAt", { ts: fmtTs(preflight.table_read_at) }) }}
            </span>
          </div>
          <table class="wizard-tools" v-if="preflight.tools.length > 0">
            <thead>
              <tr>
                <th>{{ t("sendWizard.tool") }}</th>
                <th>{{ t("sendWizard.expected") }}</th>
                <th>{{ t("sendWizard.actual") }}</th>
                <th>{{ t("sendWizard.refs") }}</th>
                <th>{{ t("sendWizard.statusCol") }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="tu in preflight.tools" :key="tu.tool" :class="`tool-row--${tu.status}`">
                <td><strong>T{{ tu.tool }}</strong></td>
                <td class="tool-expected">
                  <div v-if="tu.expected_diameter !== undefined">⌀ {{ tu.expected_diameter.toFixed(4) }}</div>
                  <div v-if="tu.comment" class="tool-comment">{{ tu.comment }}</div>
                  <div v-else-if="tu.expected_diameter === undefined" class="tool-comment">
                    {{ t("sendWizard.noComment") }}
                  </div>
                </td>
                <td class="tool-actual">
                  <span v-if="tu.actual_diameter !== undefined">⌀ {{ tu.actual_diameter.toFixed(4) }}</span>
                  <span v-else>—</span>
                  <span v-if="tu.diameter_delta !== undefined" class="tool-delta">
                    Δ {{ tu.diameter_delta >= 0 ? "+" : "" }}{{ tu.diameter_delta.toFixed(4) }}
                  </span>
                </td>
                <td class="tool-refs">{{ tu.reference_count }}×</td>
                <td>
                  <span class="badge" :class="`badge--${tu.status}`" :title="tu.status_reason">
                    {{ tu.status }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
          <div v-else class="wizard-tools__loading">
            {{ t("sendWizard.toolCheckNoTools") }}
          </div>
        </template>
      </section>

      <!-- 3. Send method ─────────────────────────────────────── -->
      <section class="wizard-section">
        <h3 class="wizard-section__title">{{ t("sendWizard.method") }}</h3>
        <label
          class="wizard-method"
          :class="{ 'is-selected': method === 'mem' }"
        >
          <input type="radio" value="mem" v-model="method" />
          <div class="wizard-method__body">
            <strong>{{ t("sendWizard.methodMem") }}</strong>
            <span class="badge badge--ok">{{ t("sendWizard.methodRecommended") }}</span>
            <p>{{ t("sendWizard.methodMemBody") }}</p>
            <div class="wizard-method__steps">
              <em>{{ t("sendWizard.controllerSteps") }}:</em>
              <ol>
                <li>{{ t("sendWizard.memStep1") }}</li>
                <li>{{ t("sendWizard.memStep2") }}</li>
                <li>{{ t("sendWizard.memStep3") }}</li>
              </ol>
            </div>
          </div>
        </label>
        <label
          class="wizard-method"
          :class="{ 'is-selected': method === 'dnc' }"
        >
          <input type="radio" value="dnc" v-model="method" />
          <div class="wizard-method__body">
            <strong>{{ t("sendWizard.methodDnc") }}</strong>
            <p>{{ t("sendWizard.methodDncBody") }}</p>
            <div class="wizard-method__steps">
              <em>{{ t("sendWizard.controllerSteps") }}:</em>
              <ol>
                <li>{{ t("sendWizard.dncStep1") }}</li>
                <li>{{ t("sendWizard.dncStep2") }}</li>
                <li>{{ t("sendWizard.dncStep3") }}</li>
              </ol>
            </div>
          </div>
        </label>
      </section>

      <!-- 4. Send ────────────────────────────────────────────── -->
      <section class="wizard-section wizard-section--actions">
        <div v-if="sendError" class="wizard-error">{{ sendError }}</div>
        <p v-if="destinationRequiresPreflight" class="wizard-preflight-note">
          <i class="material-icons">verified</i>
          {{ t("sendWizard.requirePreflightNote") }}
        </p>
        <p class="wizard-prereq" v-if="!sendable">
          <i class="material-icons">info</i>
          {{ t("sendWizard.prereqHint") }}
        </p>
        <button
          class="button button--primary wizard-send"
          :disabled="!sendable || sending"
          @click="doSend"
        >
          <i class="material-icons">send</i>
          {{ sending ? t("sendWizard.sending") : t("sendWizard.sendButton", { method: method.toUpperCase() }) }}
        </button>
      </section>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { cnc as cncApi } from "@/api";
import type { CncCheckResult, Preflight, SendMethod } from "@/api/cnc";
import { useCncStore } from "@/stores/cnc";

const props = defineProps<{
  filePath: string;
}>();

const emit = defineEmits<{
  (e: "cancel"): void;
  (e: "started"): void;
}>();

const { t } = useI18n();
const cncStore = useCncStore();

// LocalStorage-persist the operator's last-used method so the next
// send defaults to whatever they picked last. Default is "mem" since
// that's the safe path for any program that fits in NC memory.
const METHOD_KEY = "cncSendMethod";
const method = ref<SendMethod>(
  ((): SendMethod => {
    const v = localStorage.getItem(METHOD_KEY);
    return v === "dnc" ? "dnc" : "mem";
  })()
);

// Destination machine — defaults to the globally-selected one. The
// preflight + connection check both follow this id, so routing a
// send to a different controller updates the wizard's state in place
// without needing to leave the page.
const destinationId = ref<string>(cncStore.currentMachineId || "");
watch(
  () => cncStore.currentMachineId,
  (id) => {
    if (id && !destinationId.value) destinationId.value = id;
  }
);

const checkResult = ref<CncCheckResult | null>(null);
const checking = ref(false);
const sending = ref(false);
const sendError = ref<string>("");

const preflight = ref<Preflight | null>(null);
const preflightLoading = ref(false);
const preflightError = ref<string>("");

const fmtTs = (iso: string) => {
  if (!iso) return "—";
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
};

const connectionOK = computed(
  () =>
    !!checkResult.value &&
    checkResult.value.bridge.ok &&
    checkResult.value.controller.ok
);

// Send is enabled only after a successful connection check. The
// operator-side controller prep is on them — this gate just stops the
// most common foot-gun: trying to stream when the bridge is offline.
const sendable = computed(() => connectionOK.value);

// Surface a small "this destination requires preflight" note when
// the chosen machine has the server-side gate on. Heads off the 409
// surprise — operator sees the requirement before they hit Send.
const destinationRequiresPreflight = computed(() => {
  const id = destinationId.value || cncStore.currentMachineId;
  if (!id) return false;
  const m = cncStore.machines.find((x) => x.id === id);
  return !!m?.requirePreflight;
});

const runCheck = async () => {
  checking.value = true;
  try {
    checkResult.value = await cncApi.checkConnection(
      destinationId.value || undefined
    );
  } catch (e: any) {
    checkResult.value = {
      bridge: { ok: false, error: e?.message || String(e) },
      controller: { ok: false, error: "skipped — bridge unreachable" },
    };
  } finally {
    checking.value = false;
  }
};

const doSend = async () => {
  if (!sendable.value || sending.value) return;
  sending.value = true;
  sendError.value = "";
  try {
    localStorage.setItem(METHOD_KEY, method.value);
    // If the operator picked a non-default destination in the wizard
    // dropdown, switch the global store first so /machine reflects the
    // job we're about to start. setCurrentMachine is a no-op if the id
    // already matches.
    if (destinationId.value && destinationId.value !== cncStore.currentMachineId) {
      await cncStore.setCurrentMachine(destinationId.value);
    }
    await cncApi.start(
      props.filePath,
      method.value,
      destinationId.value || undefined
    );
    emit("started");
  } catch (e: any) {
    sendError.value = e?.message || String(e);
  } finally {
    sending.value = false;
  }
};

const runPreflight = async () => {
  preflightLoading.value = true;
  preflightError.value = "";
  try {
    preflight.value = await cncApi.getPreflight(
      props.filePath,
      destinationId.value || undefined
    );
  } catch (e: any) {
    preflightError.value = e?.message || String(e);
    preflight.value = null;
  } finally {
    preflightLoading.value = false;
  }
};

// Auto-run a connection check + preflight on open so the operator
// doesn't have to click twice. Both are independent — preflight
// reads the latest persisted tool table from disk, so a flaky
// bridge doesn't block the tool comparison.
onMounted(() => {
  runCheck();
  runPreflight();
});

// Re-run check + preflight when the operator changes destination so
// the displayed numbers always match the controller they're routing
// to. Skip on first paint (onMounted already covered it).
watch(destinationId, (id, prev) => {
  if (id && prev !== undefined && id !== prev) {
    checkResult.value = null;
    preflight.value = null;
    runCheck();
    runPreflight();
  }
});
</script>

<style scoped>
.send-wizard {
  /* Span both columns of the /machine grid since the wizard's content
     reads better at full width. */
  grid-column: 1 / -1;
}

.wizard-body {
  flex-direction: column;
  align-items: stretch;
  padding: 1rem 1.2rem 1.2rem;
  gap: 1rem;
  overflow: auto;
}

.wizard-section {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.wizard-section__title {
  margin: 0;
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--fg-muted, #888);
  font-weight: 600;
}

.wizard-file {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.7rem;
  border-radius: 4px;
  background: var(--alt-background, #fafafa);
  border: 1px solid var(--border-color, #eee);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 0.85rem;
  word-break: break-all;
}

.wizard-connection {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  flex-wrap: wrap;
}

.wizard-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.2rem 0.55rem;
  border-radius: 999px;
  font-size: 0.78rem;
  font-weight: 500;
}

.wizard-pill .material-icons {
  font-size: 0.95rem;
}

.wizard-pill--ok {
  background: rgba(46, 125, 50, 0.12);
  color: #2e7d32;
}

.wizard-pill--err {
  background: rgba(198, 40, 40, 0.12);
  color: #c62828;
}

.wizard-pill--unknown {
  background: var(--alt-background, #f5f5f5);
  color: var(--fg-muted, #888);
}

.wizard-pill--warn {
  background: rgba(245, 124, 0, 0.14);
  color: #ef6c00;
}

/* Tool-check table — dense per-tool comparison between expected
   (parsed from comments) and actual (from latest tool-table read). */
.wizard-tools__loading,
.wizard-tools__warning {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.5rem 0.7rem;
  border-radius: 4px;
  background: var(--alt-background, #fafafa);
  color: var(--fg-muted, #666);
  font-size: 0.85rem;
}

.wizard-tools__warning {
  background: rgba(245, 124, 0, 0.08);
  color: #ef6c00;
}

.wizard-tools__warning .material-icons {
  font-size: 1rem;
}

.wizard-tools__err {
  padding: 0.5rem 0.7rem;
  border-radius: 4px;
  background: rgba(198, 40, 40, 0.08);
  color: #c62828;
  font-size: 0.85rem;
}

.wizard-starting-tool {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.4rem 0.6rem;
  margin-bottom: 0.5rem;
  border: 1px solid var(--border-color, #ddd);
  border-radius: 4px;
  background: var(--alt-background, #fafafa);
  font-size: 0.85rem;
}

.wizard-starting-tool .material-icons {
  font-size: 1rem;
  color: var(--primaryColor, #2196f3);
}

.wizard-starting-tool--swap {
  /* Spindle currently has a different tool than the program starts
     with. Highlighted because the operator should confirm the
     incoming tool is in the carousel before pressing Cycle Start. */
  border-color: #ed6c02;
  background: rgba(237, 108, 2, 0.08);
  color: #b85100;
}

.wizard-starting-tool--swap .material-icons {
  color: #ed6c02;
}

.wizard-starting-tool__unknown {
  color: var(--fg-muted, #888);
  font-size: 0.78rem;
}

.wizard-tools__summary {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  align-items: center;
}

.wizard-tools__age {
  margin-left: auto;
  font-size: 0.75rem;
  color: var(--fg-muted, #888);
}

.wizard-tools {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.83rem;
  margin-top: 0.4rem;
}

.wizard-tools th,
.wizard-tools td {
  padding: 0.35rem 0.5rem;
  text-align: left;
  border-bottom: 1px solid var(--border-color, #eee);
  vertical-align: top;
}

.wizard-tools thead th {
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--fg-muted, #888);
  font-weight: 500;
}

.wizard-tools .tool-comment {
  color: var(--fg-muted, #888);
  font-size: 0.78rem;
}

.wizard-tools .tool-actual,
.wizard-tools .tool-expected {
  font-variant-numeric: tabular-nums;
}

.wizard-tools .tool-delta {
  display: block;
  font-size: 0.72rem;
  color: var(--fg-muted, #888);
}

.wizard-tools .tool-refs {
  text-align: right;
  color: var(--fg-muted, #888);
  font-variant-numeric: tabular-nums;
}

.wizard-tools tr.tool-row--warn td,
.wizard-tools tr.tool-row--empty td,
.wizard-tools tr.tool-row--offline td,
.wizard-tools tr.tool-row--missing td {
  background: rgba(198, 40, 40, 0.04);
}

.wizard-tools tr.tool-row--warn td {
  background: rgba(245, 124, 0, 0.06);
}

.badge--warn {
  background: rgba(245, 124, 0, 0.14);
  color: #ef6c00;
}

.badge--empty,
.badge--offline,
.badge--missing {
  background: rgba(198, 40, 40, 0.12);
  color: #c62828;
}

.wizard-connection__detail {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 0.4rem;
  font-size: 0.78rem;
  color: var(--fg-muted, #666);
}

.wizard-connection__detail .ok {
  color: #2e7d32;
}

.wizard-connection__detail .err {
  color: #c62828;
}

/* Destination dropdown — only renders when >1 machine. Same style
   as a section input field; full-width so machine names + addresses
   don't truncate. */
.wizard-destination {
  width: 100%;
  padding: 0.5rem 0.7rem;
  font-size: 0.9rem;
  border: 1px solid var(--border-color, #ddd);
  border-radius: 4px;
  background: var(--surface, #fff);
  color: inherit;
  cursor: pointer;
}

.wizard-destination:hover,
.wizard-destination:focus {
  border-color: var(--primaryColor, #2196f3);
  outline: none;
}

/* Method selector — radios styled as wide cards with explanatory body */
.wizard-method {
  display: grid;
  grid-template-columns: 1.2rem 1fr;
  gap: 0.6rem;
  padding: 0.7rem 0.9rem;
  border: 1px solid var(--border-color, #ddd);
  border-radius: 6px;
  cursor: pointer;
  transition: border-color 0.15s, background 0.15s;
}

.wizard-method:hover {
  border-color: var(--primaryColor, #2196f3);
}

.wizard-method.is-selected {
  border-color: var(--primaryColor, #2196f3);
  background: rgba(33, 150, 243, 0.04);
}

.wizard-method input[type="radio"] {
  margin-top: 0.3rem;
}

.wizard-method__body {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}

.wizard-method__body strong {
  font-size: 0.95rem;
}

.wizard-method__body .badge {
  display: inline-block;
  align-self: flex-start;
  padding: 0.05rem 0.4rem;
  border-radius: 999px;
  font-size: 0.65rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  background: rgba(46, 125, 50, 0.12);
  color: #2e7d32;
}

.wizard-method__body p {
  margin: 0;
  font-size: 0.83rem;
  color: var(--fg-muted, #666);
  line-height: 1.4;
}

.wizard-method__steps {
  font-size: 0.78rem;
  color: var(--fg-muted, #888);
}

.wizard-method__steps em {
  font-style: normal;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  font-weight: 600;
  font-size: 0.7rem;
}

.wizard-method__steps ol {
  margin: 0.2rem 0 0;
  padding-left: 1.2rem;
}

.wizard-method__steps li {
  padding: 0.05rem 0;
}

.wizard-section--actions {
  border-top: 1px solid var(--border-color, #eee);
  padding-top: 1rem;
  align-items: stretch;
}

.wizard-prereq {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.4rem;
  color: var(--fg-muted, #888);
  font-size: 0.82rem;
}

.wizard-prereq .material-icons {
  font-size: 1rem;
}

.wizard-preflight-note {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.4rem;
  font-size: 0.82rem;
  color: #1976d2;
  background: rgba(33, 150, 243, 0.08);
  border: 1px solid rgba(33, 150, 243, 0.3);
  border-radius: 4px;
  padding: 0.4rem 0.6rem;
}

.wizard-preflight-note .material-icons {
  font-size: 1rem;
}

.wizard-error {
  padding: 0.5rem 0.7rem;
  border-radius: 4px;
  background: rgba(198, 40, 40, 0.1);
  color: #c62828;
  font-size: 0.85rem;
}

.wizard-send {
  padding: 0.7rem 1rem;
  font-size: 0.95rem;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  background: var(--primaryColor, #2196f3);
  color: #fff;
  border: 0;
  border-radius: 4px;
  cursor: pointer;
}

.wizard-send:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.wizard-send:not(:disabled):hover {
  filter: brightness(1.05);
}

.link-btn {
  background: none;
  border: 0;
  color: var(--primaryColor, #2196f3);
  cursor: pointer;
  font-size: 0.85rem;
  padding: 0.2rem 0.4rem;
}

.link-btn:hover {
  text-decoration: underline;
}
</style>
