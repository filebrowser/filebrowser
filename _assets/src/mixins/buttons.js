export default {
  data: function () {
    return {
      buttonState: ''
    }
  },
  methods: {
    setLoading: function () {
      let i = this.$el.querySelector('i')
      i.style.opacity = 0

      this.buttonState = i.innerHTML

      setTimeout(() => {
        i.classList.add('spin')
        i.innerHTML = 'autorenew'
        i.style.opacity = 1
      }, 200)
    },
    setDone: function (success = true) {
      let i = this.$el.querySelector('i')
      i.style.opacity = 0

      let thirdStep = () => {
        i.innerHTML = this.buttonState
        i.style.opacity = null
      }

      let secondStep = () => {
        i.style.opacity = 0
        setTimeout(thirdStep, 200)
      }

      let firstStep = () => {
        i.classList.remove('spin')
        i.innerHTML = success
          ? 'done'
          : 'close'
        i.style.opacity = 1
        setTimeout(secondStep, 1000)
      }

      setTimeout(firstStep, 200)
    }
  }
}
/* // third step ?
if (selectedItems.length === 0 && document.getElementById('listing')) {
  document.sendCostumEvent('changed-selected')
} */
