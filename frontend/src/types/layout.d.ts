interface PopupProps {
  prompt: string;
  confirm?: any;
  action?: PopupAction;
  props?: any;
  close?: any;
}

type PopupAction = (e: Event) => void;
