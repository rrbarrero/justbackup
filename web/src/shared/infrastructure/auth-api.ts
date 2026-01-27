import { ApiClient } from "./api-client";

export interface LoginResponse {
  token: string;
}

export interface SetupStatusResponse {
  setupRequired: boolean;
}

export interface SetupRequest {
  username?: string;
  password?: string;
  confirmPassword?: string;
}

export const login = async (
  username: string,
  password: string,
): Promise<LoginResponse> => {
  return ApiClient.post<LoginResponse>("/api/login", { username, password });
};

export const setup = async (data: SetupRequest): Promise<void> => {
  await ApiClient.post("/api/setup", data);
};

export const getSetupStatus = async (): Promise<SetupStatusResponse> => {
  return ApiClient.get<SetupStatusResponse>("/api/setup-status");
};
