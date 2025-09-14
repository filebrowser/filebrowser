import { useAuthStore } from "@/stores/auth";
import router from "@/router";
import type { JwtPayload } from "jwt-decode";
import { jwtDecode } from "jwt-decode";
import { baseURL, noAuth } from "./constants";
import { StatusError } from "@/api/utils";

export function parseToken(body: SessionToken) {
  // falsy or malformed jwt will throw InvalidTokenError
  const data = jwtDecode<JwtPayload & { user: IUser }>(body.token);

  document.cookie = `auth=${body.token}; Path=/; SameSite=Strict;`;

  localStorage.setItem("jwt", body.token);

  const authStore = useAuthStore();
  authStore.jwt = body.token;
  authStore.setUser(data.user);

  const expiresAt = new Date(body.expiresAt);

  if (authStore.logoutTimer) {
    clearTimeout(authStore.logoutTimer);
  }

  authStore.setLogoutTimer(
    window.setTimeout(() => {
      logout();
    }, expiresAt.getTime() - Date.now())
  );
}

export async function validateLogin() {
  try {
    if (localStorage.getItem("jwt")) {
      await renew(<string>localStorage.getItem("jwt"));
    }
  } catch (error) {
    console.warn("Invalid JWT token in storage");
    throw error;
  }
}

export async function login(
  username: string,
  password: string,
  recaptcha: string
) {
  const data = { username, password, recaptcha };

  const res = await fetch(`${baseURL}/api/login`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  });


  if (res.status === 200) {
    const body = await res.json();
    parseToken(body);
  } else {
    const body = await res.text();
    throw new StatusError(
      body || `${res.status} ${res.statusText}`,
      res.status
    );
  }
}

export async function renew(jwt: string) {
  const res = await fetch(`${baseURL}/api/renew`, {
    method: "POST",
    headers: {
      "X-Auth": jwt,
    },
  });


  if (res.status === 200) {
    const body = await res.json();
    parseToken(body);
  } else {
    const body = await res.text();
    throw new StatusError(
      body || `${res.status} ${res.statusText}`,
      res.status
    );
  }
}

export async function signup(username: string, password: string) {
  const data = { username, password };

  const res = await fetch(`${baseURL}/api/signup`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  });

  if (res.status !== 200) {
    throw new StatusError(`${res.status} ${res.statusText}`, res.status);
  }
}

export function logout() {
  document.cookie = "auth=; Max-Age=0; Path=/; SameSite=Strict;";

  const authStore = useAuthStore();
  authStore.clearUser();

  localStorage.setItem("jwt", "");
  if (noAuth) {
    window.location.reload();
  } else {
    router.push({ path: "/login" });
  }
}
