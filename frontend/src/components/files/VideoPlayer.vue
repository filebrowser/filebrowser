<template>
  <video
    ref="videoPlayer"
    class="video-js"
    controls
    style="width: 100%; height: 100%"
  >
    <source :src="source" />
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
import Player from "video.js/dist/types/player";
import "videojs-mobile-ui";
import "videojs-hotkeys";
import { loadSubtitle } from "@/utils/subtitle";

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
      addSubtitles(props.subtitles);
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

const addSubtitles = async (subtitles: string[] | undefined) => {
  if (!subtitles) return;
  // add subtitles dynamically (srt is converted on-the-fly)
  const subs = await Promise.all(
    subtitles.map(async (s) => await loadSubtitle(s))
  );
  // TODO: player.value wouldnt work here, no idea why
  const _player = videojs.getPlayer(videoPlayer.value!);
  for (const [idx, sub] of subs.filter((s) => !!s.src).entries()) {
    _player.addRemoteTextTrack({
      src: sub.src,
      label: sub.label,
      kind: "subtitles",
      default: idx === 0,
    });
  }
};
</script>
