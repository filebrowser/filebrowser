export function eventPosition(event: MouseEvent) {
  let posx = 0;
  let posy = 0;

  if (event.pageX || event.pageY) {
    posx = event.pageX;
    posy = event.pageY;
  } else if (event.clientX || event.clientY) {
    posx =
      event.clientX +
      document.body.scrollLeft +
      document.documentElement.scrollLeft;
    posy =
      event.clientY +
      document.body.scrollTop +
      document.documentElement.scrollTop;
  }

  return {
    x: posx,
    y: posy,
  };
}
