<template>
  <div class="column">
    <form v-if="authStore?.user?.otpEnabled" class="card">
      <div class="card-title">
        <h2>{{ t("otp.name") }}</h2>
      </div>

      <div v-if="otpSetupKey" class="card-content">
        <div class="qrcode-container">
          <qrcode-vue :value="otpSetupKey" :size="300" level="M" />
        </div>
        <div class="setup-key-container">
          <input
            :value="otpSecretB32"
            class="input input--block"
            type="text"
            name="otpSetupKey"
            disabled
          />
          <button class="action copy-clipboard" @click="copyOtpSetupKey">
            <i class="material-icons">content_paste_go</i>
          </button>
        </div>
        <div class="setup-key-container">
          <input
            v-model="otpCode"
            :placeholder="t('settings.otpCodeCheckPlaceholder')"
            type="text"
            pattern="[0-9]*"
            inputmode="numeric"
            maxlength="6"
            class="input input--block"
          />
          <button class="action copy-clipboard" @click="checkOtpCode">
            <i class="material-icons">send</i>
          </button>
        </div>
        <button class="button button--block button--red" @click="disableOtp">
          {{ t("buttons.disable") }}
        </button>
      </div>

      <div v-if="!otpSetupKey" class="card-action">
        <button class="button button--flat" @click="showOtpInfo">
          {{ t("prompts.show") }}
        </button>
      </div>
    </form>

    <form v-else class="card" @submit="enable2FA">
      <div class="card-title">
        <h2>{{ t("otp.name") }}</h2>
      </div>

      <div class="card-content">
        <input
          v-if="!otpSetupKey"
          v-model="passwordForOTP"
          :placeholder="t('settings.password')"
          class="input input--block"
          type="password"
          name="password"
        />
        <template v-else>
          <div class="qrcode-container">
            <qrcode-vue :value="otpSetupKey" :size="300" level="M" />
          </div>
          <div class="setup-key-container">
            <input
              :value="otpSecretB32"
              class="input input--block"
              type="text"
              name="otpSetupKey"
              disabled
            />
            <button class="action copy-clipboard" @click="copyOtpSetupKey">
              <i class="material-icons">content_paste_go</i>
            </button>
          </div>
          <div class="setup-key-container">
            <input
              v-model="otpCode"
              :placeholder="t('settings.otpCodeCheckPlaceholder')"
              type="text"
              pattern="[0-9]*"
              inputmode="numeric"
              maxlength="6"
              class="input input--block"
            />
            <button class="action copy-clipboard" @click="checkOtpCode">
              <i class="material-icons">send</i>
            </button>
          </div>
        </template>
      </div>

      <div class="card-action">
        <input
          v-if="!otpSetupKey"
          :value="t('buttons.enable')"
          class="button button--flat"
          type="submit"
          name="submitEnableOTPForm"
        />
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { base32 } from "@scure/base";
import QrcodeVue from "qrcode.vue";
import { copy } from "@/utils/clipboard";
import { useLayoutStore } from "@/stores/layout";
import { useAuthStore } from "@/stores/auth";
import { useI18n } from "vue-i18n";
import { users as api } from "@/api";
import { inject, ref } from "vue";
import { computed } from "vue";

const layoutStore = useLayoutStore();
const authStore = useAuthStore();
const { t } = useI18n();

const $showSuccess = inject<IToastSuccess>("$showSuccess")!;
const $showError = inject<IToastError>("$showError")!;

const passwordForOTP = ref<string>("");
const otpSetupKey = ref<string>("");
const otpCode = ref<string>("");

const otpSecretB32 = computed(() => {
  const otpURI = new URL(otpSetupKey.value);
  const encoder = new TextEncoder();
  const secstr = String(otpURI.searchParams.get("secret"));
  const secret = encoder.encode(secstr);

  return base32.encode(secret);
});

const showOtpInfo = async (event: Event) => {
  event.preventDefault();
  layoutStore.showHover({
    prompt: "otp",
    confirm: async (code: string) => {
      if (authStore.user === null) {
        return;
      }

      try {
        const res = await api.getOtpInfo(authStore.user.id, code);
        otpSetupKey.value = res.setupKey;
      } catch (err: any) {
        $showError(err);
      }
    },
  });
};
const disableOtp = async (event: Event) => {
  event.preventDefault();

  layoutStore.showHover({
    prompt: "otp",
    confirm: async (code: string) => {
      if (authStore.user === null) {
        return;
      }

      try {
        await api.disableOtp(authStore.user.id, code);
        otpSetupKey.value = "";
        authStore.user.otpEnabled = false;
      } catch (err: any) {
        $showError(err);
      }
    },
  });
};
const enable2FA = async (event: Event) => {
  event.preventDefault();
  if (authStore.user === null || otpSetupKey.value) {
    return;
  }

  try {
    const res = await api.enableOTP(authStore.user.id, passwordForOTP.value);

    otpSetupKey.value = res.setupKey;
    authStore.user.otpEnabled = true;
    $showSuccess(t("otp.enabledSuccessfully"));
  } catch (err: any) {
    $showError(err);
  } finally {
    passwordForOTP.value = "";
  }
};
const copyToClipboard = async (text: string) => {
  try {
    await copy({ text });
    $showSuccess(t("success.linkCopied"));
  } catch {
    try {
      await copy({ text }, { permission: true });
      $showSuccess(t("success.linkCopied"));
    } catch (e: any) {
      $showError(e);
    }
  }
};
const copyOtpSetupKey = async (event: Event) => {
  event.preventDefault();
  await copyToClipboard(otpSecretB32.value);
};
const checkOtpCode = async (event: Event) => {
  event.preventDefault();
  if (authStore.user === null) {
    return;
  }

  try {
    await api.checkOtp(authStore.user.id, otpCode.value);
    $showSuccess(t("otp.verificationSucceed"));
  } catch (err: any) {
    console.log(err);
    $showError(t("otp.verificationFailed"));
  }
};
</script>

<style lang="css" scoped>
.qrcode-container,
.setup-key-container {
  display: flex;
  justify-content: center;
  align-items: center;
  margin: 1em 0;
}

.setup-key-container {
  justify-content: space-between;
}

.setup-key-container > * {
  margin: 0 0.5em;
}
</style>
