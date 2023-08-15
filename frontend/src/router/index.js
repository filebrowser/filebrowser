import { createRouter, createWebHistory } from "vue-router";
import Login from "@/views/Login.vue";
import Layout from "@/views/Layout.vue";
import Files from "@/views/Files.vue";
import Share from "@/views/Share.vue";
import Users from "@/views/settings/Users.vue";
import User from "@/views/settings/User.vue";
import Settings from "@/views/Settings.vue";
import GlobalSettings from "@/views/settings/Global.vue";
import ProfileSettings from "@/views/settings/Profile.vue";
import Shares from "@/views/settings/Shares.vue";
import Errors from "@/views/Errors.vue";
import store from "@/store";
import { baseURL, name } from "@/utils/constants";
import i18n, { rtlLanguages } from "@/i18n";

const titles = {
  Login: "sidebar.login",
  Share: "buttons.share",
  Files: "files.files",
  Settings: "sidebar.settings",
  ProfileSettings: "settings.profileSettings",
  Shares: "settings.shareManagement",
  GlobalSettings: "settings.globalSettings",
  Users: "settings.users",
  User: "settings.user",
  Forbidden: "errors.forbidden",
  NotFound: "errors.notFound",
  InternalServerError: "errors.internal",
};

const routes = [
  {
    path: "/login",
    name: "Login",
    component: Login,
    beforeEnter: (to, from, next) => {
      if (store.getters.isLogged) {
        return next({ path: "/files" });
      }

      next();
    },
  },
  {
    path: "/share",
    component: Layout,
    children: [
      {
        path: ":pathMatch(.*)*",
        name: "Share",
        component: Share,
      },
    ],
  },
  {
    path: "/files",
    component: Layout,
    children: [
      {
        path: ":pathMatch(.*)*",
        name: "Files",
        component: Files,
        meta: {
          requiresAuth: true,
        },
      },
    ],
  },
  {
    path: "/settings",
    component: Layout,
    children: [
      {
        path: "",
        name: "Settings",
        component: Settings,
        redirect: {
          path: "/settings/profile",
        },
        meta: {
          requiresAuth: true,
        },
        children: [
          {
            path: "profile",
            name: "ProfileSettings",
            component: ProfileSettings,
          },
          {
            path: "shares",
            name: "Shares",
            component: Shares,
          },
          {
            path: "global",
            name: "GlobalSettings",
            component: GlobalSettings,
            meta: {
              requiresAdmin: true,
            },
          },
          {
            path: "users",
            name: "Users",
            component: Users,
            meta: {
              requiresAdmin: true,
            },
          },
          {
            path: "users/:id(.*)*",
            name: "User",
            component: User,
            meta: {
              requiresAdmin: true,
            },
          },
        ],
      },
    ],
  },
  {
    path: "/403",
    name: "Forbidden",
    component: Errors,
    props: {
      errorCode: 403,
      showHeader: true,
    },
  },
  {
    path: "/404",
    name: "NotFound",
    component: Errors,
    props: {
      errorCode: 404,
      showHeader: true,
    },
  },
  {
    path: "/500",
    name: "InternalServerError",
    component: Errors,
    props: {
      errorCode: 500,
      showHeader: true,
    },
  },
  // {
  //   path: "/files",
  //   redirect: {
  //     path: "/files/",
  //   },
  // },
  {
    path: "/:catchAll(.*)*",
    redirect: (to) => `/files${to.params.catchAll}`,
  },
];

const router = createRouter({
  history: createWebHistory(baseURL),
  routes,
});

router.beforeEach((to, from, next) => {
  // const title = i18n.t(titles[to.name]);
  const title = titles[to.name];
  document.title = title + " - " + name;

  console.log({ from, to });

  /*** RTL related settings per route ****/
  const rtlSet = document.querySelector("body").classList.contains("rtl");
  const shouldSetRtl = rtlLanguages.includes(i18n.locale);
  switch (true) {
    case shouldSetRtl && !rtlSet:
      document.querySelector("body").classList.add("rtl");
      break;
    case !shouldSetRtl && rtlSet:
      document.querySelector("body").classList.remove("rtl");
      break;
  }

  if (to.matched.some((record) => record.meta.requiresAuth)) {
    if (!store.getters.isLogged) {
      next({
        path: "/login",
        query: { redirect: to.fullPath },
      });

      return;
    }

    if (to.matched.some((record) => record.meta.requiresAdmin)) {
      if (!store.state.user.perm.admin) {
        next({ path: "/403" });
        return;
      }
    }
  }

  next();
});

export default router;
