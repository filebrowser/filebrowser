interface PopupProps {
  prompt: string;
  confirm?: any;
  action?: PopupAction;
  props?: any;
  close?: (() => Promise<string>) | null;
}

type PopupAction = (e: Event) => void;
