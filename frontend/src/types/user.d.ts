interface user {
  id: number;
  locale: string;
  perm: any;
}

type userKey = keyof user;
