import { ApiClient } from "../../shared/infrastructure/api-client";
import { Host } from "../domain/models";
import { toDomain } from "./host-mapper";
import { HostDTO, CreateHostRequest } from "./host-dtos";

export type HostResponse = Host;
export type { CreateHostRequest };

export const createHost = async (data: CreateHostRequest): Promise<Host> => {
  const dto = await ApiClient.post<HostDTO>("/api/hosts", data);
  return toDomain(dto);
};

export const updateHost = async (
  id: string,
  data: CreateHostRequest,
): Promise<Host> => {
  const dto = await ApiClient.put<HostDTO>(`/api/hosts/${id}`, data);
  return toDomain(dto);
};

export const getHosts = async (): Promise<Host[]> => {
  const dtos = await ApiClient.get<HostDTO[]>("/api/hosts");
  return (dtos || []).map(toDomain);
};

export const getHost = async (id: string): Promise<Host> => {
  const dto = await ApiClient.get<HostDTO>(`/api/hosts/${id}`);
  return toDomain(dto);
};

export const deleteHost = async (id: string): Promise<void> => {
  return ApiClient.delete<void>(`/api/hosts/${id}`);
};

export const runHostBackups = async (
  id: string,
): Promise<{ task_ids: string[] }> => {
  return ApiClient.post<{ task_ids: string[] }>(`/api/hosts/${id}/run`, {});
};
