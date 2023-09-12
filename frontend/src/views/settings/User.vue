<template>
  <errors v-if="error" :errorCode="error.status" />
  <div class="row" v-else-if="!layoutStore.loading">
    <div class="column">
      <form @submit="save" class="card">
        <div class="card-title">
          <h2 v-if="user?.id === 0">{{ $t("settings.newUser") }}</h2>
          <h2 v-else>{{ $t("settings.user") }} {{ user?.username }}</h2>
        </div>

        <div class="card-content" v-if="user">
          <user-form
            v-model:user="user"
            v-model:createUserDir="createUserDir"
            :isDefault="false"
            :isNew="isNew"
          />
        </div>

        <div class="card-action">
          <button
            v-if="!isNew"
            @click.prevent="deletePrompt"
            type="button"
            class="button button--flat button--red"
            :aria-label="$t('buttons.delete')"
            :title="$t('buttons.delete')"
          >
            {{ $t("buttons.delete") }}
          </button>
          <router-link to="/settings/users">
            <button
              class="button button--flat button--grey"
              :aria-label="$t('buttons.cancel')"
              :title="$t('buttons.cancel')"
            >
              {{ $t("buttons.cancel") }}
            </button>
          </router-link>
          <input
            class="button button--flat"
            type="submit"
            :value="$t('buttons.save')"
          />
        </div>
      </form>
    </div>

    <div v-if="layoutStore.show === 'deleteUser'" class="card floating">
      <div class="card-content">
        <p>Are you sure you want to delete this user?</p>
      </div>

      <div class="card-action">
        <button
          class="button button--flat button--grey"
          @click="layoutStore.closeHovers"
          v-focus
          :aria-label="$t('buttons.cancel')"
          :title="$t('buttons.cancel')"
        >
          {{ $t("buttons.cancel") }}
        </button>
        <button class="button button--flat" @click="deleteUser">
          {{ $t("buttons.delete") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import { users as api, settings } from "@/api";
import UserForm from "@/components/settings/UserForm.vue";
import Errors from "@/views/Errors.vue";
import { computed, inject, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useI18n } from "vue-i18n";
import { StatusError } from "@/api/utils";

const error = ref<any | null>(null);
const originalUser = ref<IUser | null>(null);
const user = ref<IUser | null>(null);
const createUserDir = ref<boolean>(false);

const $showError = inject("$showError") as IToastError;
const $showSuccess = inject("$showSuccess") as IToastSuccess;

const authStore = useAuthStore();
const layoutStore = useLayoutStore();
const route = useRoute();
const router = useRouter();
const { t } = useI18n();

onMounted(() => {
  fetchData();
});

const isNew = computed(() => route.path === "/settings/users/new");

watch(route, () => fetchData());
watch(user, () => {
  if (!user.value?.perm.admin) return;
  user.value.lockPassword = false;
});

const fetchData = async () => {
  layoutStore.loading = true;

  try {
    if (isNew.value) {
      let { defaults, createUserDir: _createUserDir } = await settings.get();
      createUserDir.value = _createUserDir;
      user.value = {
        ...defaults,
        username: "",
        password: "",
        rules: [],
        lockPassword: false,
        id: 0,
      };
    } else {
      const id = Array.isArray(route.params.id)
        ? route.params.id.join("")
        : route.params.id;
      user.value = { ...(await api.get(parseInt(id))) };
    }
  } catch (e) {
    error.value = e;
  } finally {
    layoutStore.loading = false;
  }
};

const deletePrompt = () => layoutStore.showHover("deleteUser");

const deleteUser = async (e: Event) => {
  e.preventDefault();
  if (user.value === null) {
    return false;
  }
  try {
    await api.remove(user.value.id);
    router.push({ path: "/settings/users" });
    $showSuccess(t("settings.userDeleted"));
  } catch (err) {
    if (err instanceof StatusError) {
      err.status === 403 ? $showError(t("errors.forbidden")) : $showError(err);
    } else if (err instanceof Error) {
      $showError(err);
    }
  }
};
const save = async (event: Event) => {
  event.preventDefault();
  if (originalUser.value === null || user.value === null) {
    return false;
  }

  try {
    if (isNew.value) {
      const newUser: IUser = {
        ...originalUser.value,
        ...user.value,
      };

      const loc = (await api.create(newUser)) || "/settings/users";
      router.push({ path: loc });
      $showSuccess(t("settings.userCreated"));
    } else {
      await api.update(user.value);

      if (user.value.id === authStore.user?.id) {
        authStore.updateUser(user.value);
      }

      $showSuccess(t("settings.userUpdated"));
    }
  } catch (e: any) {
    $showError(e);
  }
};
</script>
