<template>
  <form @submit="save" class="dashboard">
    <h1 v-if="id === 0">New User</h1>
    <h1 v-else>User {{ username }}</h1>

    <p><label for="username">Username</label><input type="text" v-model="username" id="username"></p>
    <p><label for="password">Password</label><input type="password" :placeholder="passwordPlaceholder" v-model="password" id="password"></p>
    <p><label for="scope">Scope</label><input type="text" v-model="filesystem" id="scope"></p>

    <h2>Permissions</h2>

    <p class="small">You can set the user to be an administrator or choose the permissions individually.
      If you select "Administrator", all of the other options will be automatically checked.
      The management of users remains a privilege of an administrator.</p>

    <p><input type="checkbox" v-model="admin"> Administrator</p>
    <p><input type="checkbox" :disabled="admin" v-model="allowNew"> Create new files and directories</p>
    <p><input type="checkbox" :disabled="admin" v-model="allowEdit"> Edit, rename and delete files or directories.</p>
    <p><input type="checkbox" :disabled="admin" v-model="allowCommands"> Execute commands</p>
    <p v-for="(value, key) in permissions" :key="key">
      <input type="checkbox" :disabled="admin" v-model="permissions[key]"> {{ capitalize(key) }}
    </p>

    <h3>Commands</h3>

    <p class="small">A space separated list with the available commands for this user. Example: <i>git svn hg</i>.</p>

    <input type="text" v-model.trim="commands">

    <h2>Rules</h2>

    <p class="small">Here you can define a set of allow and disallow rules for this specific user. The blocked files won't
      show up in the listings and they won't be accessible to the user. We support regex and paths relative to
      the user's scope.</p>

    <p class="small">Each rule goes in one different line and must start with the keyword <code>allow</code> or <code>disallow</code>.
      Then you should write <code>regex</code> if you are using a regular expression and then the expression or the path.</p>

    <p class="small"><strong>Examples</strong></p>

    <ul class="small">
      <li><code>disallow regex \\/\\..+</code> - prevents the access to any dot file (such as .git, .gitignore) in every folder.</li>
      <li><code>disallow /Caddyfile</code> - blocks the access to the file named <i>Caddyfile</i> on the root of the scope</li>
    </ul>

    <textarea v-model.trim="rules"></textarea>

    <h2>Custom Stylesheet</h2>

    <textarea name="css"></textarea>

    <div v-if="$store.state.show === 'deleteUser'" class="prompt">
      <h3>Delete User</h3>
      <p>Are you sure you want to delete this user?</p>
      <div>
        <button @click="deleteUser" autofocus>Delete</button>
        <button @click="closeHovers" class="cancel">Cancel</button>
      </div>
    </div>

    <p>
      <button v-if="id !== 0" @click.prevent="deletePrompt" type="button" class="delete">Delete</button>
      <input type="submit" value="Save">
    </p>
  </form>
</template>

<script>
import { mapMutations } from 'vuex'
import api from '@/utils/api'

export default {
  name: 'user',
  data: () => {
    return {
      id: 0,
      admin: false,
      allowNew: false,
      allowEdit: false,
      allowCommands: false,
      permissions: {},
      password: '',
      username: '',
      filesystem: '',
      rules: '',
      css: '',
      commands: ''
    }
  },
  computed: {
    passwordPlaceholder () {
      if (this.$route.path === '/users/new') return ''
      return '(leave blank to avoid changes)'
    }
  },
  created () {
    this.fetchData()
  },
  watch: {
    '$route': 'fetchData',
    admin: function () {
      if (!this.admin) return
      this.allowCommands = true
      this.allowEdit = true
      this.allowNew = true
      for (let key in this.permissions) {
        this.permissions[key] = true
      }
    }
  },
  methods: {
    ...mapMutations(['closeHovers']),
    fetchData () {
      let user = this.$route.params[0]

      if (this.$route.path === '/users/new') {
        user = 'base'
      }

      api.getUser(user).then(user => {
        this.id = user.ID
        this.admin = user.admin
        this.allowCommands = user.allowCommands
        this.allowNew = user.allowNew
        this.allowEdit = user.allowEdit
        this.filesystem = user.filesystem
        this.username = user.username
        this.commands = user.commands.join(' ')
        this.css = user.css
        this.permissions = user.permissions

        for (let rule of user.rules) {
          if (rule.allow) {
            this.rules += 'allow '
          } else {
            this.rules += 'disallow '
          }

          if (rule.regex) {
            this.rules += 'regex ' + rule.regexp.raw
          } else {
            this.rules += rule.path
          }

          this.rules += '\n'
        }

        this.rules = this.rules.trim()
      }).catch(() => {
        this.$router.push({ path: '/users/new' })
      })
    },
    capitalize (name) {
      let splitted = name.split(/(?=[A-Z])/)
      name = ''

      for (let i = 0; i < splitted.length; i++) {
        name += splitted[i].charAt(0).toUpperCase() + splitted[i].slice(1) + ' '
      }

      return name.slice(0, -1)
    },
    reset () {
      this.id = 0
      this.admin = false
      this.allowNew = false
      this.allowEdit = false
      this.permissins = {}
      this.allowCommands = false
      this.password = ''
      this.username = ''
      this.filesystem = ''
      this.rules = ''
      this.css = ''
      this.commands = ''
    },
    deletePrompt (event) {
      this.$store.commit('showHover', 'deleteUser')
    },
    deleteUser (event) {
      event.preventDefault()

      api.deleteUser(this.id).then(location => {
        this.$router.push({ path: '/users' })
        this.$store.commit('showSuccess', 'User deleted!')
      }).catch(e => {
        this.$store.commit('showError', e)
      })
    },
    save (event) {
      event.preventDefault()
      let user = this.parseForm()

      if (this.$route.path === '/users/new') {
        api.newUser(user).then(location => {
          this.$router.push({ path: location })
          this.$store.commit('showSuccess', 'User created!')
        }).catch(e => {
          this.$store.commit('showError', e)
        })

        return
      }

      api.updateUser(user).then(location => {
        this.$store.commit('showSuccess', 'User updated!')
      }).catch(e => {
        this.$store.commit('showError', e)
      })
    },
    parseForm () {
      let user = {
        ID: this.id,
        username: this.username,
        password: this.password,
        filesystem: this.filesystem,
        admin: this.admin,
        allowCommands: this.allowCommands,
        allowNew: this.allowNew,
        allowEdit: this.allowEdit,
        permissions: this.permissions,
        css: this.css,
        commands: this.commands.split(' '),
        rules: []
      }

      let rules = this.rules.split('\n')

      for (let rawRule of rules) {
        let rule = {
          allow: true,
          path: '',
          regex: false,
          regexp: {
            raw: ''
          }
        }

        rawRule = rawRule.split(' ')

        // Skip a malformed rule
        if (rawRule.length < 2) {
          continue
        }

        // Skip a malformed rule
        if (rawRule[0] !== 'allow' && rawRule[0] !== 'disallow') {
          continue
        }

        rule.allow = (rawRule[0] === 'allow')
        rawRule.shift()

        if (rawRule[0] === 'regex') {
          rule.regex = true
          rawRule.shift()
          rule.regexp.raw = rawRule.join(' ')
        } else {
          rule.path = rawRule.join(' ')
        }

        user.rules.push(rule)
      }

      return user
    }
  }
}
</script>

<style>

</style>
