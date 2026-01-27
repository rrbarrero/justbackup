import { ApiClient as api } from "@/shared/infrastructure/api-client";
import { BackupStats } from "@/shared/types";

export interface DashboardStats {
  total_hosts: number;
  total_backups: number;
  active_workers: number;
  backup_stats: BackupStats;
}

export async function getDashboardStats(): Promise<DashboardStats> {
  return await api.get<DashboardStats>("/api/dashboard/stats");
}
