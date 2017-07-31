import Vue from 'vue'
import Router from 'vue-router'
import Login from '@/components/Login'
import Main from '@/components/Main'
import Files from '@/components/Files'
import Users from '@/components/Users'
import User from '@/components/User'
import GlobalSettings from '@/components/GlobalSettings'
import ProfileSettings from '@/components/ProfileSettings'
import error403 from '@/components/errors/403'
import error404 from '@/components/errors/404'
import error500 from '@/components/errors/500'
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
      path: '/',
      redirect: {
        path: '/files/'
      }
    },
    {
      path: '/*',
      component: Main,
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
          component: error403
        },
        {
          path: '/404',
          name: 'Not Found',
          component: error404
        },
        {
          path: '/500',
          name: 'Internal Server Error',
          component: error500
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
          path: '/users/',
          redirect: {
            path: '/users'
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
