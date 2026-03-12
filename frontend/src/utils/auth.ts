import { useAuthStore } from "@/stores/auth";
import router from "@/router";
import { authMethod, baseURL, noAuth, logoutPage } from "./constants";
import { StatusError } from "@/api/utils";
import { setSafeTimeout } from "@/api/utils";

export async function saveToken(token: string) {
  const authStore = useAuthStore();

  localStorage.setItem("token", token);
  authStore.jwt = token;

  const res = await fetch(`${baseURL}/api/me`, {
    headers: {
      "X-Auth": token,
    },
  });

  if (res.status !== 200) {
    throw new StatusError(
      `${res.status} ${res.statusText}`,
      res.status
    );
  }

  const user = await res.json();
  authStore.setUser(user);

  // proxy auth with custom logout subject to unknown external timeout
  if (logoutPage !== "/login" && authMethod === "proxy") {
    console.warn("idle timeout disabled with proxy auth and custom logout");
    return;
  }

  if (authStore.logoutTimer) {
    clearTimeout(authStore.logoutTimer);
  }

  // Default session timeout: 2 hours (matches server default)
  const timeout = 2 * 60 * 60 * 1000;
  authStore.setLogoutTimer(
    setSafeTimeout(() => {
      logout("inactivity");
    }, timeout)
  );
}

export async function validateLogin() {
  try {
    if (localStorage.getItem("token")) {
      await renew(<string>localStorage.getItem("token"));
    }
  } catch (error) {
    console.warn("Invalid token in storage");
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

  const body = await res.text();

  if (res.status === 200) {
    await saveToken(body);
  } else {
    throw new StatusError(
      body || `${res.status} ${res.statusText}`,
      res.status
    );
  }
}

export async function renew(token: string) {
  const res = await fetch(`${baseURL}/api/renew`, {
    method: "POST",
    headers: {
      "X-Auth": token,
    },
  });

  const body = await res.text();

  if (res.status === 200) {
    await saveToken(body);
  } else {
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
    const body = await res.text();
    throw new StatusError(
      body || `${res.status} ${res.statusText}`,
      res.status
    );
  }
}

export async function logout(reason?: string) {
  const authStore = useAuthStore();

  if (authStore.jwt) {
    try {
      await fetch(`${baseURL}/api/logout`, {
        method: "POST",
        headers: {
          "X-Auth": authStore.jwt,
        },
      });
    } catch {}
  }

  authStore.clearUser();
  localStorage.setItem("token", "");

  if (noAuth) {
    window.location.reload();
  } else if (logoutPage !== "/login") {
    document.location.href = `${logoutPage}`;
  } else {
    if (typeof reason === "string" && reason.trim() !== "") {
      router.push({
        path: "/login",
        query: { "logout-reason": reason },
      });
    } else {
      router.push({
        path: "/login",
      });
    }
  }
}
