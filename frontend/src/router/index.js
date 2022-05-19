import Vue from "vue";
import Router from "vue-router";
import Login from "@/views/Login";
import Layout from "@/views/Layout";
import Files from "@/views/Files";
import Share from "@/views/Share";
import Users from "@/views/settings/Users";
import User from "@/views/settings/User";
import Settings from "@/views/Settings";
import GlobalSettings from "@/views/settings/Global";
import ProfileSettings from "@/views/settings/Profile";
import Shares from "@/views/settings/Shares";
import Errors from "@/views/Errors";
import store from "@/store";
import { baseURL, name } from "@/utils/constants";
import i18n from "@/i18n";

Vue.use(Router);

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

const router = new Router({
  base: baseURL,
  mode: "history",
  routes: [
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
      path: "/*",
      component: Layout,
      children: [
        {
          path: "/share/*",
          name: "Share",
          component: Share,
        },
        {
          path: "/files/*",
          name: "Files",
          component: Files,
          meta: {
            requiresAuth: true,
          },
        },
        {
          path: "/settings",
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
              path: "/settings/profile",
              name: "ProfileSettings",
              component: ProfileSettings,
            },
            {
              path: "/settings/shares",
              name: "Shares",
              component: Shares,
            },
            {
              path: "/settings/global",
              name: "GlobalSettings",
              component: GlobalSettings,
              meta: {
                requiresAdmin: true,
              },
            },
            {
              path: "/settings/users",
              name: "Users",
              component: Users,
              meta: {
                requiresAdmin: true,
              },
            },
            {
              path: "/settings/users/*",
              name: "User",
              component: User,
              meta: {
                requiresAdmin: true,
              },
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
        {
          path: "/files",
          redirect: {
            path: "/files/",
          },
        },
        {
          path: "/*",
          redirect: (to) => `/files${to.path}`,
        },
      ],
    },
  ],
});

router.beforeEach((to, from, next) => {
  const title = i18n.t(titles[to.name]);
  document.title = title + " - " + name;

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
