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
import { computed, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { cnc as cncApi } from "@/api";
import type { CncCheckResult, SendMethod } from "@/api/cnc";

const props = defineProps<{
  filePath: string;
}>();

const emit = defineEmits<{
  (e: "cancel"): void;
  (e: "started"): void;
}>();

const { t } = useI18n();

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

const checkResult = ref<CncCheckResult | null>(null);
const checking = ref(false);
const sending = ref(false);
const sendError = ref<string>("");

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

const runCheck = async () => {
  checking.value = true;
  try {
    checkResult.value = await cncApi.checkConnection();
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
    await cncApi.start(props.filePath, method.value);
    emit("started");
  } catch (e: any) {
    sendError.value = e?.message || String(e);
  } finally {
    sending.value = false;
  }
};

// Auto-run a connection check on open so the operator doesn't have to
// click twice. If it fails they can still re-run manually.
onMounted(() => {
  runCheck();
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
