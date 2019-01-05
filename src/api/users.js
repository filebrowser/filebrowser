import { fetchURL, fetchJSON } from './utils'

export async function getAll () {
  return fetchJSON(`/api/users`, {})
}

export async function get (id) {
  return fetchJSON(`/api/users/${id}`, {})
}

export async function create (user) {
  const res = await fetchURL(`/api/users`, {
    method: 'POST',
    body: JSON.stringify({
      what: 'user',
      which: [],
      data: user
    })
  })

  if (res.status === 201) {
    return res.headers.get('Location')
  } else {
    throw new Error(res.status)
  }

}

export async function update (user, which = ['all']) {
  const res = await fetchURL(`/api/users/${user.id}`, {
    method: 'PUT',
    body: JSON.stringify({
      what: 'user',
      which: which,
      data: user
    })
  })

  if (res.status !== 200) {
    throw new Error(res.status)
  }
}

export async function remove (id) {
  const res = await fetchURL(`/api/users/${id}`, {
    method: 'DELETE'
  })

  if (res.status !== 200) {
    throw new Error(res.status)
  }
}
