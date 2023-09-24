interface PopupProps {
  prompt: string;
  confirm: any;
  action?: PopupAction;
}

type PopupAction = (e: Event) => void;
