<template>
  <errors v-if="error" :errorCode="error.status" />
  <div class="row" v-else-if="!loading">
    <div class="column">
      <form @submit="save" class="card">
        <div class="card-title">
          <h2 v-if="user.id === 0">{{ $t("settings.newUser") }}</h2>
          <h2 v-else>{{ $t("settings.user") }} {{ user.username }}</h2>
        </div>

        <div class="card-content">
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
          <input
            class="button button--flat"
            type="submit"
            :value="$t('buttons.save')"
          />
        </div>
      </form>
    </div>

    <div v-if="show === 'deleteUser'" class="card floating">
      <div class="card-content">
        <p>Are you sure you want to delete this user?</p>
      </div>

      <div class="card-action">
        <button
          class="button button--flat button--grey"
          @click="closeHovers"
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

<script>
import { mapActions, mapState, mapWritableState } from "pinia";
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import { users as api, settings } from "@/api";
import UserForm from "@/components/settings/UserForm.vue";
import Errors from "@/views/Errors.vue";
import deepClone from "lodash.clonedeep";

export default {
  name: "user",
  components: {
    UserForm,
    Errors,
  },
  data: () => {
    return {
      error: null,
      originalUser: null,
      user: {},
      createUserDir: false,
    };
  },
  created() {
    this.fetchData();
  },
  computed: {
    ...mapState(useAuthStore, ["user"]),
    ...mapState(useLayoutStore, ["show"]),
    ...mapWritableState(useLayoutStore, ["loading"]),
    isNew() {
      return this.$route.path === "/settings/users/new";
    },
  },
  watch: {
    $route: "fetchData",
    "user.perm.admin": function () {
      if (!this.user.perm.admin) return;
      this.user.lockPassword = false;
    },
  },
  methods: {
    ...mapActions(useAuthStore, ["setUser"]),
    ...mapActions(useLayoutStore, ["closeHovers", "showHover"]),
    async fetchData() {
      this.loading = true;

      try {
        if (this.isNew) {
          let { defaults, createUserDir } = await settings.get();
          this.createUserDir = createUserDir;
          this.user = {
            ...defaults,
            username: "",
            passsword: "",
            rules: [],
            lockPassword: false,
            id: 0,
          };
        } else {
          const id = this.$route.params.pathMatch;
          this.user = { ...(await api.get(id)) };
        }
      } catch (e) {
        this.error = e;
      } finally {
        this.loading = false;
      }
    },
    deletePrompt() {
      this.showHover("deleteUser");
    },
    async deleteUser(event) {
      event.preventDefault();

      try {
        await api.remove(this.user.id);
        this.$router.push({ path: "/settings/users" });
        this.$showSuccess(this.$t("settings.userDeleted"));
      } catch (e) {
        e.message === "403"
          ? this.$showError(this.$t("errors.forbidden"), false)
          : this.$showError(e);
      }
    },
    async save(event) {
      event.preventDefault();
      let user = {
        ...this.originalUser,
        ...this.user,
      };

      try {
        if (this.isNew) {
          const loc = await api.create(user);
          this.$router.push({ path: loc });
          this.$showSuccess(this.$t("settings.userCreated"));
        } else {
          await api.update(user);

          if (user.id === this.user.id) {
            this.setUser({ ...deepClone(user) });
          }

          this.$showSuccess(this.$t("settings.userUpdated"));
        }
      } catch (e) {
        this.$showError(e);
      }
    },
  },
};
</script>
@/stores/auth@/stores/file@/stores/layout
