import { ApiClient } from "../../shared/infrastructure/api-client";
import { Backup } from "../domain/models";
import { toDomain } from "./backup-mapper";
import {
  BackupDTO,
  BackupErrorResponse,
  CreateHookRequest,
  MeasureSizeResponse,
  TaskResult,
  HookDTO,
} from "./backup-dtos";

// Re-export types that consumers might need (or they should import from models/dtos directly,
// but for now keeping compatibility where possible, except BackupResponse is now Backup)
export type BackupResponse = Backup;
export type {
  BackupErrorResponse,
  MeasureSizeResponse,
  TaskResult,
  CreateHookRequest,
  HookDTO,
};

export const getBackups = async (
  hostId?: string,
  status?: string,
): Promise<Backup[]> => {
  const params = new URLSearchParams();
  if (hostId) params.append("host_id", hostId);
  if (status) params.append("status", status);

  const query = params.toString() ? `?${params.toString()}` : "";
  const dtos = await ApiClient.get<BackupDTO[]>(`/api/backups${query}`);
  return (dtos || []).map(toDomain);
};

export interface CreateBackupRequest {
  host_id: string;
  path: string;
  destination: string;
  schedule: string;
  excludes: string[];
  incremental: boolean;
  retention: number;
  encrypted: boolean;
  hooks: CreateHookRequest[];
}

export const createBackup = async (
  data: CreateBackupRequest,
): Promise<Backup> => {
  const dto = await ApiClient.post<BackupDTO>("/api/backups", data);
  return toDomain(dto);
};

export interface UpdateBackupRequest {
  id: string;
  path: string;
  destination: string;
  schedule: string;
  excludes: string[];
  incremental: boolean;
  retention: number;
  encrypted: boolean;
  hooks: CreateHookRequest[];
}

export async function updateBackup(
  backup: UpdateBackupRequest,
): Promise<Backup> {
  const dto = await ApiClient.put<BackupDTO>(
    `/api/backups/${backup.id}`,
    backup,
  );
  return toDomain(dto);
}

export async function deleteBackup(id: string): Promise<void> {
  return ApiClient.delete<void>(`/api/backups/${id}`);
}

export async function runBackup(id: string): Promise<void> {
  return ApiClient.post<void>(`/api/backups/${id}/run`, {});
}

export const getBackupById = async (id: string): Promise<Backup> => {
  const dto = await ApiClient.get<BackupDTO>(`/api/backups/${id}`);
  return toDomain(dto);
};

export const getBackupErrors = async (
  id: string,
): Promise<BackupErrorResponse[]> => {
  return ApiClient.get<BackupErrorResponse[]>(`/api/backups/${id}/errors`);
};

export async function deleteBackupErrors(id: string): Promise<void> {
  return ApiClient.delete<void>(`/api/backups/${id}/errors`);
}

export const measureSize = async (
  hostId: string,
  path: string,
): Promise<MeasureSizeResponse> => {
  return ApiClient.post<MeasureSizeResponse>(`/api/hosts/${hostId}/measure`, {
    path,
  });
};

export const getTaskResult = async (taskId: string): Promise<TaskResult> => {
  return ApiClient.get<TaskResult>(`/api/tasks/${taskId}`);
};

export interface RestoreRequest {
  backup_id: string;
  path: string;
  restore_type: "local" | "remote";
  restore_addr?: string;
  restore_token?: string;
  target_host_id?: string;
  target_path?: string;
}

export const restoreBackup = async (
  req: RestoreRequest,
): Promise<{ task_id: string }> => {
  return ApiClient.post<{ task_id: string }>(
    `/api/backups/${req.backup_id}/restore`,
    req,
  );
};
