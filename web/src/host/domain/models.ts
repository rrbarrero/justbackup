export interface Host {
  id: string;
  name: string;
  hostname: string;
  user: string;
  port: number;
  path: string;
  isWorkstation: boolean;
  failedBackupsCount?: number;
}
