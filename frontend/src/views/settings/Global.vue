<template>
  <errors v-if="error" :errorCode="error.status" />
  <div class="row" v-else-if="!layoutStore.loading && settings !== null">
    <div class="column">
      <form class="card" @submit.prevent="save">
        <div class="card-title">
          <h2>{{ t("settings.globalSettings") }}</h2>
        </div>

        <div class="card-content">
          <p>
            <input type="checkbox" v-model="settings.signup" />
            {{ t("settings.allowSignup") }}
          </p>

          <p>
            <input type="checkbox" v-model="settings.createUserDir" />
            {{ t("settings.createUserDir") }}
          </p>

          <p>
            <input type="checkbox" v-model="settings.hideLoginButton" />
            {{ t("settings.hideLoginButton") }}
          </p>

          <p>
            <label class="small">{{ t("settings.userHomeBasePath") }}</label>
            <input
              class="input input--block"
              type="text"
              v-model="settings.userHomeBasePath"
            />
          </p>

          <p>
            <label for="minimumPasswordLength">{{
              t("settings.minimumPasswordLength")
            }}</label>
            <vue-number-input
              controls
              v-model.number="settings.minimumPasswordLength"
              id="minimumPasswordLength"
              :min="1"
            />
          </p>

          <h3>{{ t("settings.rules") }}</h3>
          <p class="small">{{ t("settings.globalRules") }}</p>
          <rules v-model:rules="settings.rules" />

          <div v-if="enableExec">
            <h3>{{ t("settings.executeOnShell") }}</h3>
            <p class="small">{{ t("settings.executeOnShellDescription") }}</p>
            <input
              class="input input--block"
              type="text"
              placeholder="bash -c, cmd /c, ..."
              v-model="shellValue"
            />
          </div>

          <h3>{{ t("settings.branding") }}</h3>

          <i18n-t
            keypath="settings.brandingHelp"
            tag="p"
            class="small"
            scope="global"
          >
            <a
              class="link"
              target="_blank"
              href="https://filebrowser.org/customization.html#custom-branding"
              >{{ t("settings.documentation") }}</a
            >
          </i18n-t>

          <p>
            <input
              type="checkbox"
              v-model="settings.branding.disableExternal"
              id="branding-links"
            />
            {{ t("settings.disableExternalLinks") }}
          </p>

          <p>
            <input
              type="checkbox"
              v-model="settings.branding.disableUsedPercentage"
              id="branding-used-disk"
            />
            {{ t("settings.disableUsedDiskPercentage") }}
          </p>

          <p>
            <label for="theme">{{ t("settings.themes.title") }}</label>
            <themes
              class="input input--block"
              v-model:theme="settings.branding.theme"
              id="theme"
            ></themes>
          </p>

          <p>
            <label for="branding-name">{{ t("settings.instanceName") }}</label>
            <input
              class="input input--block"
              type="text"
              v-model="settings.branding.name"
              id="branding-name"
            />
          </p>

          <p>
            <label for="branding-files">{{
              t("settings.brandingDirectoryPath")
            }}</label>
            <input
              class="input input--block"
              type="text"
              v-model="settings.branding.files"
              id="branding-files"
            />
          </p>

          <h3>{{ t("settings.tusUploads") }}</h3>

          <p class="small">{{ t("settings.tusUploadsHelp") }}</p>

          <div class="tusConditionalSettings">
            <p>
              <label for="tus-chunkSize">{{
                t("settings.tusUploadsChunkSize")
              }}</label>
              <input
                class="input input--block"
                type="text"
                v-model="formattedChunkSize"
                id="tus-chunkSize"
              />
            </p>

            <p>
              <label for="tus-retryCount">{{
                t("settings.tusUploadsRetryCount")
              }}</label>
              <vue-number-input
                controls
                v-model.number="settings.tus.retryCount"
                id="tus-retryCount"
                :min="0"
              />
            </p>
          </div>

          <h3>{{ t("settings.collabora") }}</h3>

          <p class="small">{{ t("settings.collaboraHelp") }}</p>

          <p>
            <input
              type="checkbox"
              v-model="settings.collabora.enabled"
              id="collabora-enabled"
            />
            {{ t("settings.collaboraEnabled") }}
          </p>

          <p>
            <label for="collabora-url">{{ t("settings.collaboraURL") }}</label>
            <input
              class="input input--block"
              type="url"
              placeholder="https://collabora.koszhome.ro"
              v-model.trim="settings.collabora.url"
              id="collabora-url"
            />
          </p>

          <p>
            <label for="collabora-public-url">{{
              t("settings.collaboraPublicURL")
            }}</label>
            <input
              class="input input--block"
              type="url"
              placeholder="https://filebrowser.koszhome.ro"
              v-model.trim="settings.collabora.publicURL"
              id="collabora-public-url"
            />
            <span class="small">{{ t("settings.collaboraPublicURLHelp") }}</span>
          </p>

          <p>
            <label for="collabora-internal-url">{{
              t("settings.collaboraInternalURL")
            }}</label>
            <input
              class="input input--block"
              type="url"
              placeholder="http://192.168.68.204:8080"
              v-model.trim="settings.collabora.internalURL"
              id="collabora-internal-url"
            />
            <span class="small">{{ t("settings.collaboraInternalURLHelp") }}</span>
          </p>

          <p>
            <label for="collabora-wopi-secret">{{
              t("settings.collaboraWOPISecret")
            }}</label>
            <input
              class="input input--block"
              type="password"
              autocomplete="off"
              placeholder="long random secret"
              v-model.trim="settings.collabora.wopiSecret"
              id="collabora-wopi-secret"
            />
          </p>

          <p>
            <label for="collabora-token-ttl">{{
              t("settings.collaboraTokenTTL")
            }}</label>
            <input
              class="input input--block"
              type="text"
              placeholder="2h"
              v-model.trim="settings.collabora.tokenTTL"
              id="collabora-token-ttl"
            />
          </p>

          <p>
            <button
              class="button button--flat"
              type="button"
              :disabled="collaboraTesting"
              @click="testCollabora"
            >
              {{
                collaboraTesting
                  ? t("settings.collaboraTesting")
                  : t("settings.collaboraTest")
              }}
            </button>
          </p>

          <div
            v-if="collaboraTestResult"
            class="collabora-test-results"
            :class="{ 'collabora-test-results--ok': collaboraTestResult.ok }"
          >
            <p class="small">
              {{
                collaboraTestResult.ok
                  ? t("settings.collaboraTestPassed")
                  : t("settings.collaboraTestNeedsAttention")
              }}
            </p>
            <ul>
              <li
                v-for="check in collaboraTestResult.checks"
                :key="check.name"
                :class="`collabora-test-results__item collabora-test-results__item--${check.status}`"
              >
                <strong>{{ check.status.toUpperCase() }}</strong>
                <span>{{ check.message }}</span>
              </li>
            </ul>
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

    <div class="column">
      <form class="card" @submit.prevent="save">
        <div class="card-title">
          <h2>{{ t("settings.userDefaults") }}</h2>
        </div>

        <div class="card-content">
          <p class="small">{{ t("settings.defaultUserDescription") }}</p>

          <user-form
            :isNew="false"
            :isDefault="true"
            v-model:user="settings.defaults"
          />
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

    <div class="column">
      <form v-if="enableExec" class="card" @submit.prevent="save">
        <div class="card-title">
          <h2>{{ t("settings.commandRunner") }}</h2>
        </div>

        <div class="card-content">
          <i18n-t
            keypath="settings.commandRunnerHelp"
            tag="p"
            class="small"
            scope="global"
          >
            <code>FILE</code>
            <code>SCOPE</code>
            <a
              class="link"
              target="_blank"
              href="https://filebrowser.org/command-execution.html#hook-runner"
              >{{ t("settings.documentation") }}</a
            >
          </i18n-t>

          <div
            v-for="(command, key) in settings.commands"
            :key="key"
            class="collapsible"
          >
            <input :id="key" type="checkbox" />
            <label :for="key">
              <p>{{ capitalize(key) }}</p>
              <i class="material-icons">arrow_drop_down</i>
            </label>
            <div class="collapse">
              <textarea
                class="input input--block input--textarea"
                v-model.trim="commandObject[key]"
              ></textarea>
            </div>
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
import { settings as api } from "@/api";
import type { CollaboraTestResponse } from "@/api/settings";
import { StatusError } from "@/api/utils";
import Rules from "@/components/settings/Rules.vue";
import Themes from "@/components/settings/Themes.vue";
import UserForm from "@/components/settings/UserForm.vue";
import { useLayoutStore } from "@/stores/layout";
import { enableExec } from "@/utils/constants";
import { getTheme, setTheme } from "@/utils/theme";
import Errors from "@/views/Errors.vue";
import { computed, inject, onBeforeUnmount, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

const error = ref<StatusError | null>(null);
const originalSettings = ref<ISettings | null>(null);
const settings = ref<ISettings | null>(null);
const debounceTimeout = ref<number | null>(null);
const collaboraTesting = ref(false);
const collaboraTestResult = ref<CollaboraTestResponse | null>(null);

const commandObject = ref<{
  [key: string]: string[] | string;
}>({});
const shellValue = ref<string>("");

const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;

const { t } = useI18n();

const layoutStore = useLayoutStore();

const formattedChunkSize = computed({
  get() {
    return settings?.value?.tus?.chunkSize
      ? formatBytes(settings?.value?.tus?.chunkSize)
      : "";
  },
  set(value: string) {
    // Use debouncing to allow the user to type freely without
    // interruption by the formatter
    // Clear the previous timeout if it exists
    if (debounceTimeout.value) {
      clearTimeout(debounceTimeout.value);
    }

    // Set a new timeout to apply the format after a short delay
    debounceTimeout.value = window.setTimeout(() => {
      if (settings.value) settings.value.tus.chunkSize = parseBytes(value);
    }, 1500);
  },
});

// Define funcs
const capitalize = (name: string, where: string | RegExp = "_") => {
  if (where === "caps") where = /(?=[A-Z])/;
  const split = name.split(where);
  name = "";

  for (let i = 0; i < split.length; i++) {
    name += split[i].charAt(0).toUpperCase() + split[i].slice(1) + " ";
  }

  return name.slice(0, -1);
};

const normalizedCollaboraSettings = (): SettingsCollabora => ({
  configured: true,
  enabled: Boolean(settings.value?.collabora?.enabled),
  url: settings.value?.collabora?.url?.trim() ?? "",
  publicURL: settings.value?.collabora?.publicURL?.trim() ?? "",
  internalURL: settings.value?.collabora?.internalURL?.trim() ?? "",
  wopiSecret: settings.value?.collabora?.wopiSecret?.trim() ?? "",
  tokenTTL: settings.value?.collabora?.tokenTTL?.trim() || "2h",
});

const testCollabora = async () => {
  if (settings.value === null || collaboraTesting.value) return;

  collaboraTesting.value = true;
  collaboraTestResult.value = null;

  try {
    collaboraTestResult.value = await api.testCollabora(
      normalizedCollaboraSettings()
    );
    if (collaboraTestResult.value.ok) {
      $showSuccess(t("settings.collaboraTestPassed"));
    } else {
      $showError(t("settings.collaboraTestNeedsAttention"));
    }
  } catch (e: any) {
    $showError(e);
  } finally {
    collaboraTesting.value = false;
  }
};

const save = async () => {
  if (settings.value === null) return false;
  const newSettings: ISettings = {
    ...settings.value,
    shell:
      settings.value?.shell
        .join(" ")
        .trim()
        .split(" ")
        .filter((s: string) => s !== "") ?? [],
    commands: {},
  };

  newSettings.collabora = normalizedCollaboraSettings();

  const keys = Object.keys(settings.value.commands) as Array<
    keyof SettingsCommand
  >;
  for (const key of keys) {
    // not sure if we can safely assume non-null
    const newValue = commandObject.value[key];
    if (!newValue) continue;

    if (Array.isArray(newValue)) {
      newSettings.commands[key] = newValue;
    } else if (key in commandObject.value) {
      newSettings.commands[key] = newValue
        .split("\n")
        .filter((cmd: string) => cmd !== "");
    }
  }
  newSettings.shell = shellValue.value
    .trim()
    .split(" ")
    .filter((s) => s !== "");

  if (newSettings.branding.theme !== getTheme()) {
    setTheme(newSettings.branding.theme);
  }

  try {
    await api.update(newSettings);
    $showSuccess(t("settings.settingsUpdated"));
  } catch (e: any) {
    $showError(e);
  }

  return true;
};
// Parse the user-friendly input (e.g., "20M" or "1T") to bytes
const parseBytes = (input: string) => {
  const regex = /^(\d+)(\.\d+)?(B|K|KB|M|MB|G|GB|T|TB)?$/i;
  const matches = input.match(regex);
  if (matches) {
    const size = parseFloat(matches[1].concat(matches[2] || ""));
    let unit: keyof SettingsUnit =
      matches[3].toUpperCase() as keyof SettingsUnit;
    if (!unit.endsWith("B")) {
      unit += "B";
    }
    const units: SettingsUnit = {
      KB: 1024,
      MB: 1024 ** 2,
      GB: 1024 ** 3,
      TB: 1024 ** 4,
    };
    return size * (units[unit as keyof SettingsUnit] || 1);
  } else {
    return 1024 ** 2;
  }
};
// Format the chunk size in bytes to user-friendly format
const formatBytes = (bytes: number) => {
  const units = ["B", "KB", "MB", "GB", "TB"];
  let size = bytes;
  let unitIndex = 0;
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex++;
  }
  return `${size}${units[unitIndex]}`;
};

// Define Hooks

onMounted(async () => {
  try {
    layoutStore.loading = true;
    const original: ISettings = await api.get();
    const newSettings: ISettings = {
      ...original,
      collabora: {
        configured: original.collabora?.configured ?? true,
        enabled: original.collabora?.enabled ?? false,
        url: original.collabora?.url ?? "",
        publicURL: original.collabora?.publicURL ?? "",
        internalURL: original.collabora?.internalURL ?? "",
        wopiSecret: original.collabora?.wopiSecret ?? "",
        tokenTTL: original.collabora?.tokenTTL || "2h",
      },
      commands: {},
    };

    const keys = Object.keys(original.commands) as Array<keyof SettingsCommand>;
    for (const key of keys) {
      newSettings.commands[key] = original.commands[key];
      commandObject.value[key] = original.commands[key]!.join("\n");
    }

    originalSettings.value = original;
    settings.value = newSettings;
    shellValue.value = newSettings.shell.join(" ");
  } catch (err) {
    if (err instanceof Error) {
      error.value = err;
    }
  } finally {
    layoutStore.loading = false;
  }
});

// Clear the debounce timeout when the component is destroyed
onBeforeUnmount(() => {
  if (debounceTimeout.value) {
    clearTimeout(debounceTimeout.value);
  }
});
</script>


<style scoped>
.collabora-test-results {
  border-left: 4px solid var(--red);
  margin: 1em 0;
  padding: 0.75em 1em;
  background: rgba(0, 0, 0, 0.04);
}

.collabora-test-results--ok {
  border-left-color: var(--icon-green);
}

.collabora-test-results ul {
  margin: 0.5em 0 0;
  padding-left: 0;
  list-style: none;
}

.collabora-test-results__item {
  display: flex;
  gap: 0.75em;
  align-items: flex-start;
  margin: 0.4em 0;
}

.collabora-test-results__item--success strong {
  color: var(--icon-green);
}

.collabora-test-results__item--warning strong {
  color: var(--icon-yellow);
}

.collabora-test-results__item--error strong {
  color: var(--red);
}
</style>
