<template>
  <video ref="videoPlayer" class="video-max video-js" controls>
    <source :src="source" />
    <track
      kind="subtitles"
      v-for="(sub, index) in subtitles"
      :key="index"
      :src="sub"
      :label="subLabel(sub)"
      :default="index === 0"
    />
    <p class="vjs-no-js">
      Sorry, your browser doesn't support embedded videos, but don't worry, you
      can <a :href="source">download it</a>
      and watch it with your favorite video player!
    </p>
  </video>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from "vue";
import videojs from "video.js";
import type Player from "video.js/dist/types/player";
import "videojs-mobile-ui";
import "videojs-hotkeys";

import "video.js/dist/video-js.min.css";
import "videojs-mobile-ui/dist/videojs-mobile-ui.css";

const videoPlayer = ref<HTMLElement | null>(null);
const player = ref<Player | null>(null);

const props = withDefaults(
  defineProps<{
    source: string;
    subtitles?: string[];
    options?: any;
  }>(),
  {
    options: {},
  }
);

onMounted(() => {
  player.value = videojs(
    videoPlayer.value!,
    {
      html5: {
        // needed for customizable subtitles
        // TODO: add to user settings
        nativeTextTracks: false,
      },
      plugins: {
        hotkeys: {
          volumeStep: 0.1,
          seekStep: 10,
          enableModifiersForNumbers: false,
        },
      },
      ...props.options,
    },
    // onReady callback
    async () => {
      // player.value!.log("onPlayerReady", this);
    }
  );
  // TODO: need to test on mobile
  // @ts-ignore
  player.value!.mobileUi();
});

onBeforeUnmount(() => {
  if (player.value) {
    player.value.dispose();
    player.value = null;
  }
});

const subLabel = (subUrl: string) => {
  let url: URL;
  try {
    url = new URL(subUrl);
  } catch (_) {
    // treat it as a relative url
    // we only need this for filename
    url = new URL(subUrl, window.location.origin);
  }

  const label = decodeURIComponent(
    url.pathname
      .split("/")
      .pop()!
      .replace(/\.[^/.]+$/, "")
  );

  return label;
};
</script>
<style scoped>
.video-max {
  width: 100%;
  height: 100%;
}
</style>
