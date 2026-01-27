import { ApiClient } from "@/shared/infrastructure/api-client";

export interface SSHKeyResponse {
  publicKey: string;
}

export interface TokenResponse {
  token?: string;
  created_at?: string;
  exists: boolean;
}

export const getSSHKey = async (): Promise<SSHKeyResponse> => {
  return ApiClient.get<SSHKeyResponse>("/api/settings/ssh-key");
};

export const generateToken = async (): Promise<TokenResponse> => {
  return ApiClient.post<TokenResponse>("/api/settings/token", {});
};

export const getTokenStatus = async (): Promise<TokenResponse> => {
  return ApiClient.get<TokenResponse>("/api/settings/token");
};

export const revokeToken = async (): Promise<void> => {
  return ApiClient.delete<void>("/api/settings/token");
};
