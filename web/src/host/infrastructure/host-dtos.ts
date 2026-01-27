export interface HostDTO {
  id: string;
  name: string;
  hostname: string;
  user: string;
  port: number;
  path: string;
  is_workstation: boolean;
  failed_backups_count?: number;
}

export interface CreateHostRequest {
  name: string;
  hostname: string;
  user: string;
  port: number;
  path: string;
  is_workstation: boolean;
}
