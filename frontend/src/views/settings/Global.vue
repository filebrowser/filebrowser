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

          <div>
            <p class="small">{{ t("settings.userHomeBasePath") }}</p>
            <input
              class="input input--block"
              type="text"
              v-model="settings.userHomeBasePath"
            />
          </div>

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
              href="https://filebrowser.org/configuration/custom-branding"
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
              id="branding-links"
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
              <input
                class="input input--block"
                type="number"
                v-model.number="settings.tus.retryCount"
                id="tus-retryCount"
                min="0"
              />
            </p>
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
              href="https://filebrowser.org/configuration/command-runner"
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
import { useLayoutStore } from "@/stores/layout";
import { settings as api } from "@/api";
import { enableExec } from "@/utils/constants";
import UserForm from "@/components/settings/UserForm.vue";
import Rules from "@/components/settings/Rules.vue";
import Themes from "@/components/settings/Themes.vue";
import Errors from "@/views/Errors.vue";
import { computed, inject, onBeforeUnmount, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

const error = ref<any>(null);
const originalSettings = ref<ISettings | null>(null);
const settings = ref<ISettings | null>(null);
const debounceTimeout = ref<number | null>(null);

const commandObject = ref<{
  [key in keyof SettingsCommand]: string;
}>({});
const shellValue = ref<string>("");

const $showSuccess = inject<IToastSuccess>("$showSuccess") as IToastSuccess;
const $showError = inject<IToastError>("$showError") as IToastError;

const { t } = useI18n();

const layoutStore = useLayoutStore();

const formattedChunkSize = computed({
  get() {
    return settings?.value?.tus?.chunkSize
      ? formatBytes(settings?.value?.tus?.chunkSize)
      : "";
  },
  set(value: any) {
    // Use debouncing to allow the user to type freely without
    // interruption by the formatter
    // Clear the previous timeout if it exists
    if (debounceTimeout.value) {
      clearTimeout(debounceTimeout.value);
    }

    // Set a new timeout to apply the format after a short delay
    debounceTimeout.value = setTimeout(() => {
      if (settings.value) settings.value.tus.chunkSize = parseBytes(value);
    }, 1500);
  },
});

// Define funcs
const capitalize = (name: string, where: string | RegExp = "_") => {
  if (where === "caps") where = /(?=[A-Z])/;
  let splitted = name.split(where);
  name = "";

  for (let i = 0; i < splitted.length; i++) {
    name += splitted[i].charAt(0).toUpperCase() + splitted[i].slice(1) + " ";
  }

  return name.slice(0, -1);
};

const save = async () => {
  if (settings.value === null) return false;
  let newSettings: ISettings = {
    ...settings.value,
    shell:
      settings.value?.shell
        .join(" ")
        .trim()
        .split(" ")
        .filter((s: string) => s !== "") ?? [],
    commands: {},
  };

  // @ts-ignore
  for (const name of Object.keys(settings.value.commands)) {
    // @ts-ignore
    const newValue = commandObject.value[name];
    // @ts-ignore
    if (name in commandObject.value && !Array.isArray(newValue)) {
      // @ts-ignore
      newSettings.commands[name] = newValue
        .split("\n")
        .filter((cmd: string) => cmd !== "");
    } else {
      // @ts-ignore
      newSettings.commands[name] = newValue;
    }
  }
  newSettings.shell = shellValue.value.split("\n");

  try {
    await api.update(newSettings);
    $showSuccess(t("settings.settingsUpdated"));
  } catch (e: any) {
    $showError(e);
  }
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
    let newSettings: ISettings = { ...original, commands: {} };

    for (const key in original.commands) {
      // @ts-ignore
      newSettings.commands[key] = original.commands[key];
      // @ts-ignore
      commandObject.value[key] = original.commands[key].join("\n");
    }

    originalSettings.value = original;
    settings.value = newSettings;
    // @ts-ignore
    shellValue.value = newSettings.shell.join("\n");
  } catch (e) {
    error.value = e;
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
