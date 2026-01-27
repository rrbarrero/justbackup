import { ApiClient as api } from "@/shared/infrastructure/api-client";

export interface WorkerStatsReport {
  worker_id: string;
  cpu_usage: number;
  memory_total: number;
  memory_used: number;
  memory_percent: number;
  disk_total: number;
  disk_used: number;
  disk_percent: number;
  timestamp: string;
}

export interface WorkerStatsWindow {
  worker_id: string;
  reports: WorkerStatsReport[];
}

export async function getWorkersStats(): Promise<WorkerStatsWindow[]> {
  return await api.get<WorkerStatsWindow[]>("/api/v1/workers/stats");
}
