<template>
  <div class="upload-files">
    <div class="card floating">
      <div class="card-title">
        <h2>{{ $t("prompts.uploadFiles") }}</h2>
      </div>

      <div class="card-content">
        <div class="file" v-for="file in filesInUpload" :key="file.id">
          <div class="file-name">{{ file.name }}</div>
          <div class="file-progress">
            <div v-bind:style="{ width: file.progress + '%' }"></div>
          </div>
        </div>
      </div>

      <div class="card-action">
        <button type="submit" @click="$store.commit('closeHovers')" class="button button--flat" :aria-label="$t('buttons.close')" :title="$t('buttons.close')">
          {{ $t("buttons.close") }}
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import { mapGetters } from "vuex";

export default {
  name: "uploadFiles",
  computed: {
    ...mapGetters(["filesInUpload"]),
  },
  methods: {},
};
</script>
