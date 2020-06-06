<template>
  <div
    ref="container"
    class="image-ex-container"
    @touchstart="touchStart"
    @touchmove="touchMove"
    @dblclick="zoomAuto"
    @mousedown="mousedownStart"
    @mousemove="mouseMove"
    @mouseup="mouseUp"
    @wheel="wheelMove"
  >
    <img ref="imgex" :src="src" class="image-ex-img" @load="setCenter">
  </div>
</template>
<script>
export default {
  props: {
    src: String,
    moveDisabledTime: {
      type: Number,
      default: () => 200
    },
    maxScale: {
      type: Number,
      default: () => 4
    },
    minScale: {
      type: Number,
      default: () => 0.25
    },
    classList: {
      type: Array,
      default: () => []
    },
    zoomStep: {
      type: Number,
      default: () => 0.25
    },
    autofill: {
      type: Boolean,
      default: () => false
    }
  },
  data() {
    return {
      scale: 1,
      lastX: null,
      lastY: null,
      inDrag: false,
      lastTouchDistance: 0,
      moveDisabled: false,
      disabledTimer: null
    }
  },
  mounted() {
    const container = this.$refs.container
    this.classList.forEach(className => container.classList.add(className))
    // set width and height if they are zero
    if (getComputedStyle(container).width === '0px') {
      container.style.width = '100%'
    }
    if (getComputedStyle(container).height === '0px') {
      container.style.height = '100%'
    }
  },
  methods: {
    setCenter() {
      const container = this.$refs.container
      const img = this.$refs.imgex

      let rate = Math.min(
        container.clientWidth / img.clientWidth,
        container.clientHeight / img.clientHeight
      )
      if (!this.autofill && rate > 1) {
        rate = 1
      }
      // height will be auto set
      img.width = Math.floor(img.clientWidth * rate)
      img.style.top = `${Math.floor((container.clientHeight - img.clientHeight) / 2)}px`
      img.style.left = `${Math.floor((container.clientWidth - img.clientWidth) / 2)}px`
      document.addEventListener('mouseup', () => { this.inDrag = false })
    },
    mousedownStart(event) {
      this.lastX = null
      this.lastY = null
      this.inDrag = true
      event.preventDefault()
    },
    mouseMove(event) {
      if (!this.inDrag) return
      this.doMove(event.movementX, event.movementY)
      event.preventDefault()
    },
    mouseUp(event) {
      this.inDrag = false
      event.preventDefault()
    },
    touchStart(event) {
      this.lastX = null
      this.lastY = null
      this.lastTouchDistance = null
      event.preventDefault()
    },
    zoomAuto(event) {
      switch (this.scale) {
        case 1:
          this.scale = 2
          break
        case 2:
          this.scale = 4
          break
        default:
        case 4:
          this.scale = 1
          break
      }
      this.setZoom()
      event.preventDefault()
    },
    touchMove(event) {
      event.preventDefault()
      if (this.lastX === null) {
        this.lastX = event.targetTouches[0].pageX
        this.lastY = event.targetTouches[0].pageY
        return
      }
      const step = this.$refs.imgex.width / 5
      if (event.targetTouches.length === 2) {
        this.moveDisabled = true
        clearTimeout(this.disabledTimer)
        this.disabledTimer = setTimeout(
          () => (this.moveDisabled = false),
          this.moveDisabledTime
        )

        const p1 = event.targetTouches[0]
        const p2 = event.targetTouches[1]
        const touchDistance = Math.sqrt(
          Math.pow(p2.pageX - p1.pageX, 2) + Math.pow(p2.pageY - p1.pageY, 2)
        )
        if (!this.lastTouchDistance) {
          this.lastTouchDistance = touchDistance
          return
        }
        this.scale += (touchDistance - this.lastTouchDistance) / step
        this.lastTouchDistance = touchDistance
        this.setZoom()
      } else if (event.targetTouches.length === 1) {
        if (this.moveDisabled) return
        const x = event.targetTouches[0].pageX - this.lastX
        const y = event.targetTouches[0].pageY - this.lastY
        if (Math.abs(x) >= step && Math.abs(y) >= step) return
        this.lastX = event.targetTouches[0].pageX
        this.lastY = event.targetTouches[0].pageY
        this.doMove(x, y)
      }
    },
    doMove(x, y) {
      const style = this.$refs.imgex.style
      style.left = `${this.pxStringToNumber(style.left) + x}px`
      style.top = `${this.pxStringToNumber(style.top) + y}px`
    },
    wheelMove(event) {
      this.scale += (event.wheelDeltaY / 100) * this.zoomStep
      this.setZoom()
    },
    setZoom() {
      this.scale = this.scale < this.minScale ? this.minScale : this.scale
      this.scale = this.scale > this.maxScale ? this.maxScale : this.scale
      this.$refs.imgex.style.transform = `scale(${this.scale})`
    },
    pxStringToNumber(style) {
      return +style.replace('px', '')
    }
  }
}
</script>
<style>
.image-ex-container {
  margin: auto;
  overflow: hidden;
  position: relative;
}

.image-ex-img {
  left: 0;
  top: 0;
  position: absolute;
  transition: transform 0.1s ease;
}
</style>
