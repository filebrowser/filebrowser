<template>
  <div :class="['login-view', { 'login-view--with-captcha': recaptcha }]">
    <form @submit.prevent="submit" class="login-form">
      <div class="logo login-form__logo">
        <img :src="logoURL" :alt="name" class="logo__image" />
      </div>

      <h1 class="login-form__title">{{ title }}</h1>

      <span class="login-form__description">
        To use all the <b>{{ name }}</b> features you need to
        {{ title.toLowerCase() }} first.
      </span>

      <div class="login-form__inputs">
        <Input
          v-model="username"
          autocapitalize="off"
          class="login-form__input"
          :placeholder="$t('login.username')"
          autofocus
        />

        <Input
          type="password"
          v-model="password"
          class="login-form__input"
          :placeholder="$t('login.password')"
        />

        <Input
          v-if="createMode"
          type="password"
          v-model="passwordConfirm"
          class="login-form__input"
          :placeholder="$t('login.passwordConfirm')"
        />
      </div>

      <Button class="login-form__submit" fullWidth>
        {{ title }}
      </Button>

      <div v-if="recaptcha" id="recaptcha"></div>

      <p v-if="!signup" class="login-form__change-mode" @click="toggleMode">
        {{
          createMode ? $t("login.loginInstead") : $t("login.createAnAccount")
        }}
      </p>

      <span v-if="error" class="login-form__error">
        {{ error }}
      </span>
    </form>
  </div>
</template>

<script>
import * as auth from "@/utils/auth";
import {
  name,
  logoURL,
  recaptcha,
  recaptchaKey,
  signup,
} from "@/utils/constants";

import Input from "@/components/Input.vue";
import Button from "@/components/Button.vue";

export default {
  name: "LoginView",

  components: {
    Input,
    Button,
  },

  data() {
    return {
      name,
      signup,
      logoURL,
      recaptcha,
      createMode: false,
      error: "",
      username: "",
      password: "",
      passwordConfirm: "",
    };
  },

  computed: {
    title() {
      return this.createMode
        ? this.$t("login.signup")
        : this.$t("login.submit");
    },
  },

  watch: {
    error(value) {
      if (value) {
        setTimeout(() => {
          this.error = "";
        }, 5000);
      }
    },
  },

  mounted() {
    this.renderRecaptcha();
  },

  methods: {
    renderRecaptcha() {
      if (!this.recaptcha) return;

      window.grecaptcha.ready(() => {
        window.grecaptcha.render("recaptcha", { sitekey: recaptchaKey });
      });
    },

    toggleMode() {
      this.createMode = !this.createMode;
    },

    async submit() {
      let captcha = "";
      if (recaptcha) {
        captcha = window.grecaptcha.getResponse();

        if (captcha === "") {
          this.error = this.$t("login.wrongCredentials");
          return;
        }
      }

      if (this.createMode && this.password !== this.passwordConfirm) {
        this.error = this.$t("login.passwordsDontMatch");
        return;
      }

      try {
        if (this.createMode) {
          await auth.signup(this.username, this.password);
        }

        await auth.login(this.username, this.password, captcha);

        const redirect = this.$route.query.redirect || "/files/";
        this.$router.push({ path: redirect });
      } catch (e) {
        this.error =
          e.message == 409
            ? this.$t("login.usernameTaken")
            : this.$t("login.wrongCredentials");
      }
    },
  },
};
</script>

<style src="@/css/login.css" />
