export interface HookDTO {
  id: string;
  name: string;
  phase: string;
  enabled: boolean;
  params: Record<string, string>;
}

export interface BackupDTO {
  id: string;
  host_id: string;
  host_name: string;
  host_address: string;
  path: string;
  destination: string;
  status: string;
  schedule: string;
  last_run: string;
  excludes: string[];
  incremental: boolean;
  size?: string;
  retention: number;
  encrypted: boolean;
  hooks: HookDTO[];
}

export interface CreateHookRequest {
  name: string;
  phase: string;
  enabled: boolean;
  params: Record<string, string>;
}

export interface BackupErrorResponse {
  id: string;
  job_id: string;
  backup_id: string;
  occurred_at: string;
  error_message: string;
}

// Keeping other DTOs here if they are shared or related
export interface MeasureSizeResponse {
  task_id: string;
}

export interface TaskResult {
  type: string;
  task_id: string;
  status: string;
  message: string;
  data?: any;
}
