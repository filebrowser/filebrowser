import { fetchURL, fetchJSON } from "./utils";

export async function getAll() {
  return fetchJSON(`/api/users`, {});
}

export async function get(id: number) {
  return fetchJSON(`/api/users/${id}`, {});
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
}

export async function update(user: IUser, which = ["all"]) {
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
