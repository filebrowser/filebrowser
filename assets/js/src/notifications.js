$.noty.themes.admin = {
  name: 'admin',
  helpers: {},
  modal: {
    css: {
      position: 'fixed',
      width: '100%',
      height: '100%',
      backgroundColor: '#000',
      zIndex: 10000,
      opacity: 0.6,
      display: 'none',
      left: 0,
      top: 0
    }
  }
};

$.noty.defaults = {
  layout: 'topRight',
  theme: 'admin',
  dismissQueue: true,
  animation: {
    open: 'animated bounceInRight',
    close: 'animated fadeOut',
    easing: 'swing',
    speed: 500 // opening & closing animation speed
  },
  timeout: false, // delay for closing event. Set false for sticky notifications
  force: false, // adds notification to the beginning of queue when set to true
  modal: false,
  maxVisible: 5, // you can set max visible notification for dismissQueue true option,
  killer: false, // for close all notifications before show
  closeWith: ['click'], // ['click', 'button', 'hover', 'backdrop'] // backdrop click will close all notifications
  callback: {
    onShow: function() {},
    afterShow: function() {},
    onClose: function() {},
    afterClose: function() {},
    onCloseClick: function() {},
  },
  buttons: false // an array of buttons
};

notification = function(options) {
  var icon;

  switch (options.type) {
    case "success":
      icon = '<i class="fa fa-check"></i>';
      break;
    case "error":
      icon = '<i class="fa fa-times"></i>';
      break;
    case "warning":
      icon = '<i class="fa fa-exclamation"></i>';
      break;
    case "information":
      icon = '<i class="fa fa-info"></i>';
      break;
    default:
      icon = '<i class="fa fa-bell"></i>';
  }

  var defaults = {
    template: '<div class="noty_message"><span class="noty_icon">' + icon + '</span><span class="noty_text"></span></div>'
  }

  options = $.extend({}, defaults, options);
  noty(options);
}