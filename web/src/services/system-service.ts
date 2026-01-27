import { ApiClient as api } from "@/shared/infrastructure/api-client";

export interface DiskUsage {
  total: string;
  free: string;
  used: string;
}

export const getDiskUsage = async (): Promise<DiskUsage> => {
  return await api.get<DiskUsage>("/api/system/disk-usage");
};
