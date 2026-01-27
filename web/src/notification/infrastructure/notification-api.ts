import { ApiClient } from "../../shared/infrastructure/api-client";

export interface NotificationSettings {
  provider_type: string;
  config: Record<string, any>;
  enabled: boolean;
}

export interface GotifyConfig {
  url: string;
  token: string;
}

export const notificationApi = {
  getSettings: async (providerType: string): Promise<NotificationSettings> => {
    return ApiClient.get<NotificationSettings>(
      `/api/settings/notifications?provider=${providerType}`,
    );
  },

  updateSettings: async (
    settings: NotificationSettings,
  ): Promise<NotificationSettings> => {
    return ApiClient.put<NotificationSettings>(
      "/api/settings/notifications",
      settings,
    );
  },
};
