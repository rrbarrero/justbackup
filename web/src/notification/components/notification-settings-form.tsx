"use client";

import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { Switch } from "@/shared/ui/switch";
import { NotificationPluginCard } from "./notification-plugin-card";
import { notificationApi } from "@/notification/infrastructure/notification-api";

const notificationSchema = z
  .object({
    enabled: z.boolean(),
    gotifyUrl: z.string().optional(),
    gotifyToken: z.string().optional(),
    gotifyNotifyOnSuccess: z.boolean(),
    smtpEnabled: z.boolean(),
    smtpHost: z.string().optional(),
    smtpPort: z.number().optional(),
    smtpUser: z.string().optional(),
    smtpPassword: z.string().optional(),
    smtpFrom: z.string().optional(),
    smtpTo: z.string().optional(),
    smtpNotifyOnSuccess: z.boolean(),
    pushbulletEnabled: z.boolean(),
    pushbulletAccessToken: z.string().optional(),
    pushbulletNotifyOnSuccess: z.boolean(),
  })
  .refine(
    (data) => {
      if (data.enabled) {
        return !!data.gotifyUrl && !!data.gotifyToken;
      }
      return true;
    },
    {
      message: "URL and Token are required when Gotify is enabled",
      path: ["gotifyUrl"],
    },
  )
  .refine(
    (data) => {
      if (data.smtpEnabled) {
        return (
          !!data.smtpHost &&
          !!data.smtpPort &&
          !!data.smtpUser &&
          !!data.smtpPassword &&
          !!data.smtpFrom &&
          !!data.smtpTo
        );
      }
      return true;
    },
    {
      message: "All SMTP fields are required when enabled",
      path: ["smtpHost"],
    },
  )
  .refine(
    (data) => {
      if (data.pushbulletEnabled) {
        return !!data.pushbulletAccessToken;
      }
      return true;
    },
    {
      message: "Access Token is required when Pushbullet is enabled",
      path: ["pushbulletAccessToken"],
    },
  );

type NotificationFormValues = z.infer<typeof notificationSchema>;

export function NotificationSettingsForm() {
  const [isLoading, setIsLoading] = useState(false);
  const [isFetching, setIsFetching] = useState(true);
  const [status, setStatus] = useState<{
    type: "success" | "error";
    message: string;
  } | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors },
    watch,
    setValue,
    reset,
  } = useForm<NotificationFormValues>({
    resolver: zodResolver(notificationSchema),
    defaultValues: {
      enabled: false,
      gotifyUrl: "",
      gotifyToken: "",
      gotifyNotifyOnSuccess: false,
      smtpEnabled: false,
      smtpHost: "",
      smtpPort: 587,
      smtpUser: "",
      smtpPassword: "",
      smtpFrom: "",
      smtpTo: "",
      smtpNotifyOnSuccess: false,
      pushbulletEnabled: false,
      pushbulletAccessToken: "",
      pushbulletNotifyOnSuccess: false,
    },
  });

  const enabled = watch("enabled");
  const smtpEnabled = watch("smtpEnabled");

  useEffect(() => {
    async function loadSettings() {
      try {
        const [gotifySettings, smtpSettings, pushbulletSettings] =
          await Promise.allSettled([
            notificationApi.getSettings("gotify"),
            notificationApi.getSettings("smtp"),
            notificationApi.getSettings("pushbullet"),
          ]);

        const newValues: Partial<NotificationFormValues> = {};

        if (gotifySettings.status === "fulfilled") {
          newValues.enabled = gotifySettings.value.enabled;
          newValues.gotifyUrl = gotifySettings.value.config.url || "";
          newValues.gotifyToken = gotifySettings.value.config.token || "";
          newValues.gotifyNotifyOnSuccess =
            gotifySettings.value.config.notify_on_success || false;
        }

        if (smtpSettings.status === "fulfilled") {
          newValues.smtpEnabled = smtpSettings.value.enabled;
          newValues.smtpHost = smtpSettings.value.config.host || "";
          newValues.smtpPort = smtpSettings.value.config.port || 587;
          newValues.smtpUser = smtpSettings.value.config.user || "";
          newValues.smtpPassword = smtpSettings.value.config.password || "";
          newValues.smtpFrom = smtpSettings.value.config.from || "";
          newValues.smtpTo = smtpSettings.value.config.to
            ? smtpSettings.value.config.to.join(", ")
            : "";
          newValues.smtpNotifyOnSuccess =
            smtpSettings.value.config.notify_on_success || false;
        }

        if (pushbulletSettings.status === "fulfilled") {
          newValues.pushbulletEnabled = pushbulletSettings.value.enabled;
          newValues.pushbulletAccessToken =
            pushbulletSettings.value.config.access_token || "";
          newValues.pushbulletNotifyOnSuccess =
            pushbulletSettings.value.config.notify_on_success || false;
        }

        reset((prev) => ({ ...prev, ...newValues }));
      } catch (error) {
        console.error("Failed to load notification settings", error);
      } finally {
        setIsFetching(false);
      }
    }
    loadSettings();
  }, [reset]);

  const onSubmit = async (data: NotificationFormValues) => {
    setIsLoading(true);
    setStatus(null);
    try {
      const promises = [];

      // Save Gotify settings
      promises.push(
        notificationApi.updateSettings({
          provider_type: "gotify",
          enabled: data.enabled,
          config: {
            url: data.gotifyUrl,
            token: data.gotifyToken,
            notify_on_success: data.gotifyNotifyOnSuccess,
          },
        }),
      );

      // Save SMTP settings
      promises.push(
        notificationApi.updateSettings({
          provider_type: "smtp",
          enabled: data.smtpEnabled,
          config: {
            host: data.smtpHost,
            port: data.smtpPort,
            user: data.smtpUser,
            password: data.smtpPassword,
            from: data.smtpFrom,
            to: data.smtpTo ? data.smtpTo.split(",").map((s) => s.trim()) : [],
            notify_on_success: data.smtpNotifyOnSuccess,
          },
        }),
      );

      // Save Pushbullet settings
      promises.push(
        notificationApi.updateSettings({
          provider_type: "pushbullet",
          enabled: data.pushbulletEnabled,
          config: {
            access_token: data.pushbulletAccessToken,
            notify_on_success: data.pushbulletNotifyOnSuccess,
          },
        }),
      );

      await Promise.all(promises);

      setStatus({
        type: "success",
        message: "Notification settings have been updated successfully.",
      });
    } catch (error) {
      console.error(error);
      setStatus({ type: "error", message: "Failed to save settings." });
    } finally {
      setIsLoading(false);
    }
  };

  if (isFetching) {
    return <div>Loading settings...</div>;
  }

  return (
    <div className="space-y-6">
      <div>
        <h3 className="text-lg font-medium">Notification Plugins</h3>
        <p className="text-sm text-muted-foreground">
          Configure external services to receive alerts about your backups.
        </p>
      </div>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        {status && (
          <div
            className={`p-3 text-sm rounded-md ${
              status.type === "success"
                ? "bg-green-50 text-green-600 border border-green-200"
                : "bg-red-50 text-red-600 border border-red-200"
            }`}
          >
            {status.message}
          </div>
        )}

        <NotificationPluginCard
          title="Gotify"
          description="Receive push notifications via your self-hosted Gotify server."
          logoSrc="/gotify_logo.png"
          isEnabled={enabled}
          onEnableChange={(checked) => setValue("enabled", checked)}
        >
          <div className="grid gap-4">
            <div className="grid gap-2">
              <Label htmlFor="gotifyUrl">Server URL</Label>
              <Input
                id="gotifyUrl"
                placeholder="https://gotify.example.com"
                {...register("gotifyUrl")}
              />
              {errors.gotifyUrl && (
                <p className="text-sm text-red-500">
                  {errors.gotifyUrl.message}
                </p>
              )}
            </div>
            <div className="grid gap-2">
              <Label htmlFor="gotifyToken">App Token</Label>
              <Input
                id="gotifyToken"
                type="password"
                placeholder="A..."
                {...register("gotifyToken")}
              />
              {errors.gotifyToken && (
                <p className="text-sm text-red-500">
                  {errors.gotifyToken.message}
                </p>
              )}
            </div>
            <div className="flex items-center space-x-2">
              <Switch
                id="gotifyNotifyOnSuccess"
                checked={watch("gotifyNotifyOnSuccess")}
                onCheckedChange={(checked: boolean) =>
                  setValue("gotifyNotifyOnSuccess", checked)
                }
              />
              <Label htmlFor="gotifyNotifyOnSuccess">Notify on Success</Label>
            </div>
            <div className="flex justify-end">
              <Button type="submit" disabled={isLoading}>
                {isLoading ? "Saving..." : "Save Configuration"}
              </Button>
            </div>
          </div>
        </NotificationPluginCard>

        <NotificationPluginCard
          title="SMTP"
          description="Receive email notifications via an SMTP server."
          logoSrc="/smtp_logo.png"
          isEnabled={smtpEnabled}
          onEnableChange={(checked) => setValue("smtpEnabled", checked)}
        >
          <div className="grid gap-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="smtpHost">SMTP Host</Label>
                <Input
                  id="smtpHost"
                  placeholder="smtp.example.com"
                  {...register("smtpHost")}
                />
                {errors.smtpHost && (
                  <p className="text-sm text-red-500">
                    {errors.smtpHost.message}
                  </p>
                )}
              </div>
              <div className="grid gap-2">
                <Label htmlFor="smtpPort">Port</Label>
                <Input
                  id="smtpPort"
                  type="number"
                  placeholder="587"
                  {...register("smtpPort", { valueAsNumber: true })}
                />
                {errors.smtpPort && (
                  <p className="text-sm text-red-500">
                    {errors.smtpPort.message}
                  </p>
                )}
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="smtpUser">Username</Label>
                <Input
                  id="smtpUser"
                  placeholder="user@example.com"
                  {...register("smtpUser")}
                />
                {errors.smtpUser && (
                  <p className="text-sm text-red-500">
                    {errors.smtpUser.message}
                  </p>
                )}
              </div>
              <div className="grid gap-2">
                <Label htmlFor="smtpPassword">Password</Label>
                <Input
                  id="smtpPassword"
                  type="password"
                  placeholder="********"
                  {...register("smtpPassword")}
                />
                {errors.smtpPassword && (
                  <p className="text-sm text-red-500">
                    {errors.smtpPassword.message}
                  </p>
                )}
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="grid gap-2">
                <Label htmlFor="smtpFrom">From Email</Label>
                <Input
                  id="smtpFrom"
                  placeholder="backup@example.com"
                  {...register("smtpFrom")}
                />
                {errors.smtpFrom && (
                  <p className="text-sm text-red-500">
                    {errors.smtpFrom.message}
                  </p>
                )}
              </div>
              <div className="grid gap-2">
                <Label htmlFor="smtpTo">To Email(s)</Label>
                <Input
                  id="smtpTo"
                  placeholder="admin@example.com"
                  {...register("smtpTo")}
                />
                <p className="text-xs text-muted-foreground">
                  Comma separated for multiple recipients
                </p>
                {errors.smtpTo && (
                  <p className="text-sm text-red-500">
                    {errors.smtpTo.message}
                  </p>
                )}
              </div>
            </div>
            <div className="flex items-center space-x-2">
              <Switch
                id="smtpNotifyOnSuccess"
                checked={watch("smtpNotifyOnSuccess")}
                onCheckedChange={(checked: boolean) =>
                  setValue("smtpNotifyOnSuccess", checked)
                }
              />
              <Label htmlFor="smtpNotifyOnSuccess">Notify on Success</Label>
            </div>
            <div className="flex justify-end">
              <Button type="submit" disabled={isLoading}>
                {isLoading ? "Saving..." : "Save Configuration"}
              </Button>
            </div>
          </div>
        </NotificationPluginCard>

        <NotificationPluginCard
          title="Pushbullet"
          description="Receive push notifications on your devices via Pushbullet."
          logoSrc="/pushbullet_logo.png"
          isEnabled={watch("pushbulletEnabled")}
          onEnableChange={(checked) => setValue("pushbulletEnabled", checked)}
        >
          <div className="grid gap-4">
            <div className="grid gap-2">
              <Label htmlFor="pushbulletAccessToken">Access Token</Label>
              <Input
                id="pushbulletAccessToken"
                type="password"
                placeholder="o.pG..."
                {...register("pushbulletAccessToken")}
              />
              {errors.pushbulletAccessToken && (
                <p className="text-sm text-red-500">
                  {errors.pushbulletAccessToken.message}
                </p>
              )}
            </div>
            <div className="flex items-center space-x-2">
              <Switch
                id="pushbulletNotifyOnSuccess"
                checked={watch("pushbulletNotifyOnSuccess")}
                onCheckedChange={(checked: boolean) =>
                  setValue("pushbulletNotifyOnSuccess", checked)
                }
              />
              <Label htmlFor="pushbulletNotifyOnSuccess">
                Notify on Success
              </Label>
            </div>
            <div className="flex justify-end">
              <Button type="submit" disabled={isLoading}>
                {isLoading ? "Saving..." : "Save Configuration"}
              </Button>
            </div>
          </div>
        </NotificationPluginCard>
      </form>
    </div>
  );
}
