export interface Hook {
  id: string;
  name: string;
  phase: string;
  enabled: boolean;
  params: Record<string, string>;
}

export interface Backup {
  id: string;
  hostId: string;
  hostName: string;
  hostAddress: string;
  path: string;
  destination: string;
  status: string;
  schedule: string;
  lastRun: string;
  excludes: string[];
  incremental: boolean;
  size?: number;

  retention: number;
  encrypted: boolean;
  hooks: Hook[];
}
