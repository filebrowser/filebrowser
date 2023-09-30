<template>
  <video
    ref="videoPlayer"
    class="video-js"
    controls
    style="width: 100%; height: 100%"
  >
    <source :src="source" />
    <track
      kind="subtitles"
      v-for="(sub, index) in subtitles"
      :key="index"
      :src="sub"
      :label="'Subtitle ' + index"
      :default="index === 0"
    />
    <p class="vjs-no-js">
      Sorry, your browser doesn't support embedded videos, but don't worry, you
      can <a :href="source">download it</a>
      and watch it with your favorite video player!
    </p>
  </video>
</template>

<script>
import videojs from "video.js";
import "videojs-mobile-ui";
import "videojs-hotkeys";

import "video.js/dist/video-js.min.css";
import "videojs-mobile-ui/dist/videojs-mobile-ui.css";

export default {
  name: "VideoPlayer",
  props: {
    source: {
      type: String,
      default() {
        return "";
      },
    },
    options: {
      type: Object,
      default() {
        return {};
      },
    },
    subtitles: {
      type: Array,
      default() {
        return [];
      },
    },
  },
  data() {
    return {
      player: null,
    };
  },
  mounted() {
    this.player = videojs(
      this.$refs.videoPlayer,
      {
        ...this.options,
        plugins: {
          hotkeys: {
            volumeStep: 0.1,
            seekStep: 10,
            enableModifiersForNumbers: false,
          },
        },
      },
      () => {
        // this.player.log("onPlayerReady", this);
      }
    );
    this.player.mobileUi();
  },
  beforeUnmount() {
    if (this.player) {
      this.player.dispose();
    }
  },
};
</script>
