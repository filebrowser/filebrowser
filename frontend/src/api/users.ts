import { fetchURL, fetchJSON, StatusError } from "./utils";

export async function getAll() {
  return fetchJSON<IUser[]>(`/api/users`, {});
}

export async function get(id: number) {
  return fetchJSON<IUser>(`/api/users/${id}`, {});
}

export async function create(user: IUser) {
  const res = await fetchURL(`/api/users`, {
    method: "POST",
    body: JSON.stringify({
      what: "user",
      which: [],
      data: user,
    }),
  });

  if (res.status === 201) {
    return res.headers.get("Location");
  }

  throw new StatusError(await res.text(), res.status);
}

export async function update(user: Partial<IUser>, which = ["all"]) {
  await fetchURL(`/api/users/${user.id}`, {
    method: "PUT",
    body: JSON.stringify({
      what: "user",
      which: which,
      data: user,
    }),
  });
}

export async function remove(id: number) {
  await fetchURL(`/api/users/${id}`, {
    method: "DELETE",
  });
}

export async function enableOTP(id: number, password: string) {
  const res = await fetchURL(`/api/users/${id}/otp`, {
    method: "POST",
    body: JSON.stringify({
      password,
    }),
  });
  const payload: IOtpSetupKey = await res.json();

  return payload;
}

export async function checkOtp(id: number, code: string) {
  return fetchURL(`/api/users/${id}/otp/check`, {
    method: "POST",
    body: JSON.stringify({
      code,
    }),
  });
}

export async function getOtpInfo(id: number, code: string) {
  const res = await fetchURL(`/api/users/${id}/otp`, {
    method: "GET",
    headers: {
      "X-TOTP-CODE": code,
    },
  });
  const payload: IOtpSetupKey = await res.json();

  return payload;
}

export async function disableOtp(id: number, code: string) {
  return fetchURL(`/api/users/${id}/otp`, {
    method: "DELETE",
    headers: {
      "X-TOTP-CODE": code,
    },
  });
}
