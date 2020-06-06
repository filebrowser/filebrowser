<template>
  <div>
    <h3>{{ $t('settings.permissions') }}</h3>
    <p class="small">{{ $t('settings.permissionsHelp') }}</p>

    <p><input v-model="admin" type="checkbox"> {{ $t('settings.administrator') }}</p>

    <p><input v-model="perm.create" type="checkbox" :disabled="admin"> {{ $t('settings.perm.create') }}</p>
    <p><input v-model="perm.delete" type="checkbox" :disabled="admin"> {{ $t('settings.perm.delete') }}</p>
    <p><input v-model="perm.download" type="checkbox" :disabled="admin"> {{ $t('settings.perm.download') }}</p>
    <p><input v-model="perm.modify" type="checkbox" :disabled="admin"> {{ $t('settings.perm.modify') }}</p>
    <p><input v-model="perm.execute" type="checkbox" :disabled="admin"> {{ $t('settings.perm.execute') }}</p>
    <p><input v-model="perm.rename" type="checkbox" :disabled="admin"> {{ $t('settings.perm.rename') }}</p>
    <p><input v-model="perm.share" type="checkbox" :disabled="admin"> {{ $t('settings.perm.share') }}</p>
  </div>
</template>

<script>
export default {
  name: 'Permissions',
  props: ['perm'],
  computed: {
    admin: {
      get() {
        return this.perm.admin
      },
      set(value) {
        if (value) {
          for (const key in this.perm) {
            this.perm[key] = true
          }
        }

        this.perm.admin = value
      }
    }
  }
}
</script>
