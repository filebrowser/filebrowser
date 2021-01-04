<template>
  <div
    class="image-ex-container"
    ref="container"
    @touchstart="touchStart"
    @touchmove="touchMove"
    @touchend="touchEnd"
    @dblclick="zoomAuto"
    @mousedown="mousedownStart"
    @mousemove="mouseMove"
    @mouseup="mouseUp"
    @wheel="wheelMove"
  >
    <img :src="src" class="image-ex-img image-ex-img-center" ref="imgex" @load="onLoad">
  </div>
</template>
<script>
import throttle from 'lodash.throttle'

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
      default: () => 1
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
      touches: 0,
      navOffset: 50,
      lastTouchDistance: 0,
      moveDisabled: false,
      disabledTimer: null,
      imageLoaded: false,
      position: {
        center: { x: 0, y: 0 },
        relative: { x: 0, y: 0 }
      }
    }
  },
  mounted() {
    let container = this.$refs.container
    this.classList.forEach(className => container.classList.add(className))
    // set width and height if they are zero
    if (getComputedStyle(container).width === "0px") {
      container.style.width = "100%"
    }
    if (getComputedStyle(container).height === "0px") {
      container.style.height = "100%"
    }

    window.addEventListener('resize', this.onResize)
  },
  beforeDestroy () {
    window.removeEventListener('resize', this.onResize)
    document.removeEventListener('mouseup', this.onMouseUp)
  },
  watch: {
    src: function () {
      this.setCenter()
    }
  },
  methods: {
    fit() {
      let img = this.$refs.imgex

      const wScale = window.innerWidth / img.clientWidth
      const hScale = window.innerHeight / img.clientHeight

      this.scale = wScale < hScale? wScale: hScale
      this.minScale = this.scale
      this.setZoom()
    },
    refit() {
      const target = this.fitScreenTarget()
      this.doMove(target[0], target[1])
    },
    fitScreenTarget() {
      if (this.scale <= this.minScale) {
        let style = this.$refs.imgex.style
        let posX = this.pxStringToNumber(style.left)
        let posY = this.pxStringToNumber(style.top)
        return [this.position.center.x - posX, this.position.center.y - posY]
      }
      else {
        let img = this.$refs.imgex

        const rect = img.getBoundingClientRect()
        const width = window.innerWidth
        const height = window.innerHeight

        let x = 0,y = 0

        // left out of viewport
        if (rect.left < 0 && rect.right < width) x = width - rect.right

        // right out of viewport
        else if (rect.left > 0 && rect.right > width) x = -rect.left

        // top out of viewport
        if (rect.top < 0 && rect.bottom < height) y = height - rect.bottom

        // bottom out of viewport
        else if (rect.top > 0 && rect.bottom > height) y = -rect.top

        return [x,y]
      }
    },
    checkNav(x) {
      if (this.scale <= this.minScale) {
        if (x > this.navOffset) this.$root.$emit('gallery-nav', 0)
        else if (x < -this.navOffset) this.$root.$emit('gallery-nav', 1)
      } else {
        let img = this.$refs.imgex

        const rect = img.getBoundingClientRect()
        const width = window.innerWidth

        if (rect.left > this.navOffset && rect.right > width + this.navOffset) this.$root.$emit('gallery-nav', 0)
        else if (rect.left < - this.navOffset && rect.right < width - this.navOffset) this.$root.$emit('gallery-nav', 1)
      }
    },
    onLoad() {
      let img = this.$refs.imgex

      this.imageLoaded = true

      if (img === undefined) {
        return
      }

      img.classList.remove('image-ex-img-center')
      this.setCenter()
      img.classList.add('image-ex-img-ready')

      document.addEventListener('mouseup', this.onMouseUp)
    },
    onMouseUp() {
      this.inDrag = false
      this.refit()
    },
    onResize: throttle(function() {
      if (this.imageLoaded) {
        this.setCenter()
        this.doMove(this.position.relative.x, this.position.relative.y)
      }
    }, 100),
    setCenter() {
      let container = this.$refs.container
      let img = this.$refs.imgex

      this.position.center.x = Math.floor((window.innerWidth - img.clientWidth) / 2 - container.offsetLeft)
      this.position.center.y = Math.floor((window.innerHeight - img.clientHeight) / 2 - container.offsetTop)

      img.style.left = this.position.center.x + 'px'
      img.style.top = this.position.center.y + 'px'

      this.fit()
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
      this.checkNav(event.movementX)
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

      setTimeout(() => {
        this.touches = 0
      }, 300)

      this.touches++
      if (this.touches > 1) {
        this.zoomAuto(event)
      }
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
      let step = this.$refs.imgex.width / 5
      if (event.targetTouches.length === 2) {
        this.moveDisabled = true
        clearTimeout(this.disabledTimer)
        this.disabledTimer = setTimeout(
          () => (this.moveDisabled = false),
          this.moveDisabledTime
        )

        let p1 = event.targetTouches[0]
        let p2 = event.targetTouches[1]
        let touchDistance = Math.sqrt(
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
        let x = event.targetTouches[0].pageX - this.lastX
        let y = event.targetTouches[0].pageY - this.lastY
        if (Math.abs(x) >= step && Math.abs(y) >= step) return
        this.lastX = event.targetTouches[0].pageX
        this.lastY = event.targetTouches[0].pageY
        this.doMove(x, y)
        this.checkNav(x)
      }
    },
    touchEnd(event) {
      if (event.targetTouches.length === 0) {
        this.refit()
      }
    },
    doMove(x, y) {
      let style = this.$refs.imgex.style
      let posX = this.pxStringToNumber(style.left) + x
      let posY = this.pxStringToNumber(style.top) + y

      style.left = posX + 'px'
      style.top = posY + 'px'

      this.position.relative.x =  Math.abs(this.position.center.x - posX)
      this.position.relative.y =  Math.abs(this.position.center.y - posY)

      if (posX < this.position.center.x) {
        this.position.relative.x = this.position.relative.x * -1
      }

      if (posY < this.position.center.y) {
        this.position.relative.y = this.position.relative.y * -1
      }
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
      return +style.replace("px", "")
    }
  }
}
</script>
<style>
.image-ex-container {
  margin: auto;
  position: relative;
}

.image-ex-img {
  position: absolute;
}

.image-ex-img-center {
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  position: absolute;
  transition: none;
}

.image-ex-img-ready {
  left: 0;
  top: 0;
  transition: transform 0.1s ease;
}
</style>
