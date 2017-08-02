import Vue from 'vue'
import Router from 'vue-router'
import Login from '@/views/Login'
import Layout from '@/views/Layout'
import Files from '@/views/Files'
import Users from '@/views/Users'
import User from '@/views/User'
import GlobalSettings from '@/views/GlobalSettings'
import ProfileSettings from '@/views/ProfileSettings'
import Error403 from '@/views/errors/403'
import Error404 from '@/views/errors/404'
import Error500 from '@/views/errors/500'
import auth from '@/utils/auth.js'
import store from '@/store'

Vue.use(Router)

const router = new Router({
  base: document.querySelector('meta[name="base"]').getAttribute('content'),
  mode: 'history',
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: Login,
      beforeEnter: function (to, from, next) {
        auth.loggedIn()
          .then(() => {
            next({ path: '/files' })
          })
          .catch(() => {
            document.title = 'Login'
            next()
          })
      }
    },
    {
      path: '/*',
      component: Layout,
      meta: {
        requiresAuth: true
      },
      children: [
        {
          path: '/files/*',
          name: 'Files',
          component: Files
        },
        {
          path: '/settings',
          name: 'Settings',
          redirect: {
            path: '/settings/profile'
          }
        },
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
          path: '/users',
          name: 'Users',
          component: Users,
          meta: {
            requiresAdmin: true
          }
        },
        {
          path: '/users/*',
          name: 'User',
          component: User,
          meta: {
            requiresAdmin: true
          }
        },
        {
          path: '/files',
          redirect: {
            path: '/files/'
          }
        },
        {
          path: '/*',
          redirect: {
            name: 'Files'
          }
        }
      ]
    }
  ]
})

router.beforeEach((to, from, next) => {
  document.title = to.name

  if (to.matched.some(record => record.meta.requiresAuth)) {
    // this route requires auth, check if logged in
    // if not, redirect to login page.
    auth.loggedIn()
      .then(() => {
        if (to.matched.some(record => record.meta.requiresAdmin)) {
          if (store.state.user.admin) {
            next()
            return
          }

          next({
            path: '/403'
          })

          return
        }

        next()
      })
      .catch(e => {
        next({
          path: '/login',
          query: { redirect: to.fullPath }
        })
      })

    return
  }

  next()
})

export default router
