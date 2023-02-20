import store from "@/store";
import router from "@/router";
import { Base64 } from "js-base64";
import { baseURL, authMethod } from "@/utils/constants";
import cookie from "@/utils/cookie";

export function parseToken(token) {
  const parts = token.split(".");

  if (parts.length !== 3) {
    throw new Error("token malformed");
  }

  const data = JSON.parse(Base64.decode(parts[1]));

  document.cookie = `auth=${token}; path=/`;

  localStorage.setItem("jwt", token);
  store.commit("setJWT", token);
  store.commit("setUser", data.user);
}

export async function validateLogin() {
  let jwt = localStorage.getItem("jwt");

  if (authMethod === "oidc" && (!jwt || jwt === "null")) {
    jwt = cookie("auth");
  }

  try {
    if (jwt) {
      await renew(jwt);
    }
  } catch (_) {
    console.warn('Invalid JWT token in storage') // eslint-disable-line
  }
}

export async function login(username, password, recaptcha) {
  const data = { username, password, recaptcha };

  const res = await fetch(`${baseURL}/api/login`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  });

  const body = await res.text();
  if (res.status === 200) {
    parseToken(body);
  } else {
    throw new Error(body);
  }
}

export async function renew(jwt) {
  const res = await fetch(`${baseURL}/api/renew`, {
    method: "POST",
    headers: {
      "X-Auth": jwt,
    },
  });

  const body = await res.text();

  if (res.status === 200) {
    parseToken(body);
  } else {
    if (authMethod === "oidc") {
      clearLoginState();
      document.location.replace(document.location.pathname);
    }
    throw new Error(body);
  }
}

export async function signup(username, password) {
  const data = { username, password };

  const res = await fetch(`${baseURL}/api/signup`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  });

  if (res.status !== 200) {
    throw new Error(res.status);
  }
}

export function logout() {
  clearLoginState();
  router.push({ path: "/login" });
}

function clearLoginState() {
  document.cookie = "auth=; expires=Thu, 01 Jan 1970 00:00:01 GMT; path=/";
  store.commit("setJWT", "");
  store.commit("setUser", null);
  localStorage.setItem("jwt", null);
}
