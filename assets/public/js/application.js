"use strict";

var deleteHandler = function(event) {
  event.preventDefault();
  let name = this.dataset.delete;

  let element = document.getElementById(name);
  document.removeChild(element);

  //$('label[for="' + name + '"]').fadeOut().remove();
//  $('#' + name).fadeOut().remove();

  return false;
}

document.addEventListener("DOMContentLoaded", function(event) {
  //alert("hugo is working");

  return false;
});
