<template>
  <div class="row">
    <div class="column">
      <form class="card" @submit="updateSettings">
        <div class="card-title">
          <h2>{{ t("settings.profileSettings") }}</h2>
        </div>

        <div class="card-content">
          <p>
            <input type="checkbox" v-model="hideDotfiles" />
            {{ t("settings.hideDotfiles") }}
          </p>
          <p>
            <input type="checkbox" v-model="singleClick" />
            {{ t("settings.singleClick") }}
          </p>
          <p>
            <input type="checkbox" v-model="dateFormat" />
            {{ t("settings.setDateFormat") }}
          </p>
          <h3>{{ t("settings.language") }}</h3>
          <languages
            class="input input--block"
            v-model:locale="locale"
          ></languages>
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
      <form
        class="card"
        v-if="!authStore.user?.lockPassword"
        @submit="updatePassword"
      >
        <div class="card-title">
          <h2>{{ t("settings.changePassword") }}</h2>
        </div>

        <div class="card-content">
          <input
            :class="passwordClass"
            type="password"
            :placeholder="t('settings.newPassword')"
            v-model="password"
            name="password"
          />
          <input
            :class="passwordClass"
            type="password"
            :placeholder="t('settings.newPasswordConfirm')"
            v-model="passwordConf"
            name="password"
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
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import { users as api } from "@/api";
import Languages from "@/components/settings/Languages.vue";
// import i18n, { rtlLanguages } from "@/i18n";
import { inject, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";

const layoutStore = useLayoutStore();
const authStore = useAuthStore();
const { t } = useI18n();

const $showSuccess = inject("$showSuccess") as IToastSuccess;
const $showError = inject("$showError") as IToastError;

const password = ref<string>("");
const passwordConf = ref<string>("");
const hideDotfiles = ref<boolean>(false);
const singleClick = ref<boolean>(false);
const dateFormat = ref<boolean>(false);
const locale = ref<string>("");

// ...mapState(useAuthStore, ["user"]),
// ...mapWritableState(useLayoutStore, ["loading"]),
//...mapActions(useAuthStore, ["updateUser"]),

const passwordClass = () => {
  const baseClass = "input input--block";

  if (password.value === "" && passwordConf.value === "") {
    return baseClass;
  }

  if (password.value === passwordConf.value) {
    return `${baseClass} input--green`;
  }

  return `${baseClass} input--red`;
};

onMounted(() => {
  layoutStore.loading = true;
  if (authStore.user === null) return false;
  locale.value = authStore.user.locale;
  hideDotfiles.value = authStore.user.hideDotfiles;
  singleClick.value = authStore.user.singleClick;
  dateFormat.value = authStore.user.dateFormat;
  layoutStore.loading = false;
});

const updatePassword = async (event: Event) => {
  event.preventDefault();

  if (
    password.value !== passwordConf.value ||
    password.value === "" ||
    authStore.user === null
  ) {
    return;
  }

  try {
    const data = {...authStore.user, id: authStore.user.id, password: password.value };
    await api.update(data, ["password"]);
    authStore.updateUser(data);
    $showSuccess(t("settings.passwordUpdated"));
  } catch (e: any) {
    $showError(e);
  }
};
const updateSettings = async (event: Event) => {
  event.preventDefault();

  try {
    if (authStore.user === null) throw "User is not set";

    const data = {
      ...authStore.user,
      id: authStore.user.id,
      locale: locale.value,
      hideDotfiles: hideDotfiles.value,
      singleClick: singleClick.value,
      dateFormat: dateFormat.value,
    };
    // TODO don't know what this is doing
    const shouldReload = false;
    // rtlLanguages.includes(data.locale) !==
    // rtlLanguages.includes(i18n.locale);
    await api.update(data, [
      "locale",
      "hideDotfiles",
      "singleClick",
      "dateFormat",
    ]);
    authStore.updateUser(data);
    if (shouldReload) {
      location.reload();
    }
    $showSuccess(t("settings.settingsUpdated"));
  } catch (e: any) {
    $showError(e);
  }
};
</script>
