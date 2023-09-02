import { useAuthStore } from "@/stores/auth";
import router from "@/router";
import jwt_decode from "jwt-decode";
import { fetchURL } from "@/api/utils";
import { baseURL } from "@/utils/constants";

export function parseToken(token) {
  // falsy or malformed jwt will throw InvalidTokenError
  const data = jwt_decode(token);

  document.cookie = `auth=${token}; Path=/; SameSite=Strict;`;

  localStorage.setItem("jwt", token);

  const authStore = useAuthStore();
  authStore.jwt = token;
  authStore.setUser(data.user);
}

export async function validateLogin() {
  try {
    if (localStorage.getItem("jwt")) {
      await renew(localStorage.getItem("jwt"));
    }
  } catch (_) {
    console.warn("Invalid JWT token in storage"); // eslint-disable-line
  }
}

export async function login(username, password, recaptcha) {
  const data = { username, password, recaptcha };

  const res = await fetchURL(
    `/api/login`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
    },
    false
  );

  const body = await res.text();

  if (res.status === 200) {
    parseToken(body);
  } else {
    throw new Error(body);
  }
}

export async function renew(jwt) {
  const res = await fetchURL(`/api/renew`, {
    method: "POST",
    headers: {
      "X-Auth": jwt,
    },
  });

  const body = await res.text();

  if (res.status === 200) {
    parseToken(body);
  } else {
    throw new Error(body);
  }
}

export async function signup(username, password) {
  const data = { username, password };

  const res = await fetchURL(
    `/api/signup`,
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(data),
    },
    false
  );

  if (res.status !== 200) {
    throw new Error(res.status);
  }
}

export function logout() {
  document.cookie = "auth=; Max-Age=0; Path=/; SameSite=Strict;";

  const authStore = useAuthStore();
  authStore.clearUser();

  localStorage.setItem("jwt", null);
  router.push({ path: "/login" });
}
