import { removePrefix } from "./utils";
import { baseURL } from "@/utils/constants";
import { useAuthStore } from "@/stores/auth";

const ssl = window.location.protocol === "https:";
const protocol = ssl ? "wss:" : "ws:";

export default function command(
  url: string,
  command: string,
  onmessage: WebSocket["onmessage"],
  onclose: WebSocket["onclose"]
) {
  const authStore = useAuthStore();

  url = removePrefix(url);
  url = `${protocol}//${window.location.host}${baseURL}/api/command${url}?auth=${authStore.jwt}`;

  const conn = new window.WebSocket(url);
  conn.onopen = () => conn.send(command);
  conn.onmessage = onmessage;
  conn.onclose = onclose;
}
