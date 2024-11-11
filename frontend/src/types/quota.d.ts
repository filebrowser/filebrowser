interface IQuota {
  inodes: QuotaInfo;
  space: QuotaInfo;
}

interface QuotaInfo {
  quota: number;
  usage: number;
}
