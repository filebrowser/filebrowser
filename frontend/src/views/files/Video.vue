<template>
  <div id="file-video" ref="player" style="height: 100%; width: 100%" />
</template>
<script>
export default {
  props: {
    src: {
      type: String,
      default: "",
    },
  },
  data() {
    return {
      instance: null,
      videoConfig: {
        container: "#file-video", //“#”代表容器的ID，“`.`”或“”代表容器的class
        variable: "player", // 播放函数名称，该属性必需设置，值等于下面的new ckplayer()的对象
        video: this.src, // 视频地址
        mobileCkControls: true, // 移动端h5显示控制栏
        overspread: false, //是否让视频铺满播放器
        seek: 0, // 默认需要跳转的秒数
      },
    };
  },
  computed: {
    localKey() {
      return this.src.replace(/t=d+/, ""); // 避免时间戳的干扰
    },
  },
  mounted() {
    this.loadProcess();
    // eslint-disable-next-line no-undef
    this.instance = new ckplayer(this.videoConfig); //初始化播放器
    window.player = this.instance;
    this.$nextTick(() => this.loadHandler());
  },
  methods: {
    loadProcess() {
      this.videoConfig.seek = localStorage.getItem(this.localKey) || 0;
    },
    loadHandler() {
      this.instance.addListener("time", this.timeHandler); //监听播放时间
      this.instance.addListener("ended", this.VideoPlayEndedHandler); //监听播放结束
    },
    timeHandler(time) {
      localStorage.setItem(this.localKey, time);
    },
    VideoPlayEndedHandler() {},
  },
};
</script>
