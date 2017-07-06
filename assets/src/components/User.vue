<template>
  <div class="dashboard">
    <h1>User</h1>

    <p><label for="username">Username</label><input type="text" v-model="username" name="username"></p>
    <p><label for="password">Password</label><input type="password" :disabled="passwordBlock" v-model="password" name="password"></p>
    <p><label for="scope">Scope</label><input type="text" v-model="scope" name="scope"></p>

    <hr>

    <h2>Permissions</h2>

    <p class="small">You can set the user to be an administrator or choose the permissions individually.
      If you select "Administrator", all of the other options will be automatically checked.
      The management of users remains a privilege of an administrator.</p>

    <p><input type="checkbox" v-model="admin"> Administrator</p>
    <p><input type="checkbox" :disabled="admin" v-model="allowNew"> Create new files and directories</p>
    <p><input type="checkbox" :disabled="admin" v-model="allowEdit"> Edit, rename and delete files or directories.</p>
    <p><input type="checkbox" :disabled="admin" v-model="allowCommands"> Execute commands</p>

    <h3>Commands</h3>

    <p class="small">A space separated list with the available commands for this user. Example: <i>git svn hg</i>.</p>

    <input type="text" v-model="commands">

    <hr>

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

    <textarea v-model="rules"></textarea>

    <hr>

    <h2>CSS</h2>

    <p class="small">Costum user CSS</p>

    <textarea name="css"></textarea>
  </div>
</template>

<script>
import api from '@/utils/api'

export default {
  name: 'user',
  data: () => {
    return {
      admin: false,
      allowNew: false,
      allowEdit: false,
      allowCommands: false,
      passwordBlock: true,
      password: '',
      username: '',
      scope: '',
      rules: '',
      css: '',
      commands: ''
    }
  },
  created () {
    if (this.$route.path === '/users/new') return

    api.getUser(this.$route.params[0]).then(user => {
      this.admin = user.admin
      this.allowCommands = user.allowCommands
      this.allowNew = user.allowNew
      this.allowEdit = user.allowEdit
      this.scope = user.filesystem
      this.username = user.username
      this.commands = user.commands.join(' ')
      this.css = user.css

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
    }).catch(error => {
      console.log(error)
    })
  },
  watch: {
    admin: function () {
      if (!this.admin) return
      this.allowCommands = true
      this.allowEdit = true
      this.allowNew = true
    }
  }
}
</script>

<style>
.dashboard {
  max-width: 600px;
}

.dashboard textarea,
.dashboard input[type="text"],
.dashboard input[type="password"] {
  padding: .5em 1em;
  display: block;
  border: 1px solid #e9e9e9;
  transition: .2s ease border;
  color: #333;
  width: 100%;
}

.dashboard textarea:focus,
.dashboard textarea:hover,
.dashboard input[type="text"]:focus,
.dashboard input[type="password"]:focus,
.dashboard input[type="text"]:hover,
.dashboard input[type="password"]:hover {
  border-color: #9f9f9f;
}

.dashboard textarea {
  font-family: monospace;
  min-height: 10em;
  resize: vertical;
}

.dashboard p label {
  margin-bottom: .2em;
  display: block;
  font-size: .8em
}

hr {
    border-bottom: 2px solid rgba(181, 181, 181, 0.5);
    border-top: 0;
    border-right: 0;
    border-left: 0;
    margin: 1em 0;
}

li code,
p code {
  background: rgba(0, 0, 0, 0.05);
  padding: .1em;
  border-radius: .2em;
}

.small {
  font-size: .8em;
  line-height: 1.5;
}
</style>
