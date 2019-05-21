import Vue from 'vue'
import Router from 'vue-router'
import Login from '@/views/Login'
import Layout from '@/views/Layout'
import Files from '@/views/Files'
import Share from '@/views/Share'
import Users from '@/views/settings/Users'
import User from '@/views/settings/User'
import Settings from '@/views/Settings'
import GlobalSettings from '@/views/settings/Global'
import ProfileSettings from '@/views/settings/Profile'
import Error403 from '@/views/errors/403'
import Error404 from '@/views/errors/404'
import Error500 from '@/views/errors/500'
import store from '@/store'
import { baseURL } from '@/utils/constants'

Vue.use(Router)

const router = new Router({
  base: baseURL,
  mode: 'history',
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: Login,
      beforeEnter: (to, from, next) => {
        if (store.getters.isLogged) {
          return next({ path: '/files' })
        }

        document.title = 'Login'
        next()
      }
    },
    {
      path: '/*',
      component: Layout,
      children: [
        {
          path: '/share/*',
          name: 'Share',
          component: Share
        },
        {
          path: '/files/*',
          name: 'Files',
          component: Files,
          meta: {
            requiresAuth: true
          }
        },
        {
          path: '/settings',
          name: 'Settings',
          component: Settings,
          redirect: {
            path: '/settings/profile'
          },
          meta: {
            requiresAuth: true
          },
          children: [
            {
              path: '/settings/profile',
              name: 'Profile Settings',
              component: ProfileSettings
            },
            {
              path: '/settings/global',
              name: 'Global Settings',
              component: GlobalSettings,
              meta: {
                requiresAdmin: true
              }
            },
            {
              path: '/settings/users',
              name: 'Users',
              component: Users,
              meta: {
                requiresAdmin: true
              }
            },
            {
              path: '/settings/users/*',
              name: 'User',
              component: User,
              meta: {
                requiresAdmin: true
              }
            }
          ]
        },
        {
          path: '/403',
          name: 'Forbidden',
          component: Error403
        },
        {
          path: '/404',
          name: 'Not Found',
          component: Error404
        },
        {
          path: '/500',
          name: 'Internal Server Error',
          component: Error500
        },
        {
          path: '/files',
          redirect: {
            path: '/files/'
          }
        },
        {
          path: '/*',
          redirect: to => `/files${to.path}`
        }
      ]
    }
  ]
})

router.beforeEach((to, from, next) => {
  document.title = to.name

  if (to.matched.some(record => record.meta.requiresAuth)) {
    if (!store.getters.isLogged) {
      next({
        path: '/login',
        query: { redirect: to.fullPath }
      })

      return
    }

    if (to.matched.some(record => record.meta.requiresAdmin)) {
      if (!store.state.user.perm.admin) {
        next({ path: '/403' })
        return
      }
    }
  }

  next()
})

export default router
