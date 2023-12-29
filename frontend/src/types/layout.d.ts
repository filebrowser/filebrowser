interface PopupProps {
  prompt: string;
  confirm?: any;
  action?: PopupAction;
  props?: any;
}

type PopupAction = (e: Event) => void;
