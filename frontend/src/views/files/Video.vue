<template>
  <div class="file-video" ref="player" style="height: 100%; width: 100%"/>
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
    };
  },
  mounted() {
    var videoObject = {
      container: ".file-video", //“#”代表容器的ID，“`.`”或“”代表容器的class
      variable: "player", // 播放函数名称，该属性必需设置，值等于下面的new ckplayer()的对象
      video: this.src, // 视频地址
      mobileCkControls: true, // 移动端h5显示控制栏
      overspread: false, //是否让视频铺满播放器
    };
    // eslint-disable-next-line no-undef
    this.instance = new ckplayer(videoObject); //初始化播放器
  },
};
</script>
