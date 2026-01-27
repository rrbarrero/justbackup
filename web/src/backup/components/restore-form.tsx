"use client";

import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import {
  BackupResponse,
  restoreBackup,
} from "@/backup/infrastructure/backup-api";
import { HostResponse, getHosts } from "@/host/infrastructure/host-api";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";
import { Spinner } from "@/shared/ui/spinner";

const restoreSchema = z.object({
  path: z.string().min(1, "Path inside backup is required"),
  restore_type: z.enum(["local", "remote"]),
  target_host_id: z.string().optional(),
  target_path: z.string().optional(),
});

type RestoreFormValues = z.infer<typeof restoreSchema>;

interface RestoreFormProps {
  backup: BackupResponse;
  onSuccess: (taskId: string) => void;
}

export function RestoreForm({ backup, onSuccess }: RestoreFormProps) {
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [hosts, setHosts] = useState<HostResponse[]>([]);
  const [loadingHosts, setLoadingHosts] = useState(false);

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<RestoreFormValues>({
    resolver: zodResolver(restoreSchema),
    defaultValues: {
      path: "/",
      restore_type: "remote",
      target_host_id: backup.hostId,
      target_path: backup.path,
    },
  });

  const restoreType = watch("restore_type");

  useEffect(() => {
    if (restoreType === "remote") {
      setLoadingHosts(true);
      getHosts()
        .then(setHosts)
        .catch((err) => setError("Failed to load hosts"))
        .finally(() => setLoadingHosts(false));
    }
  }, [restoreType]);

  const onSubmit = async (data: RestoreFormValues) => {
    setIsSubmitting(true);
    setError(null);

    try {
      if (data.restore_type === "remote" && !data.target_path) {
        throw new Error("Target path is required for remote restoration");
      }

      const { task_id } = await restoreBackup({
        backup_id: backup.id,
        path: data.path,
        restore_type: data.restore_type,
        target_host_id: data.target_host_id,
        target_path: data.target_path,
      });

      onSuccess(task_id);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to trigger restore",
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Card className="bg-card shadow-sm border-border">
      <CardHeader>
        <CardTitle className="text-xl font-bold">
          Restore Configuration
        </CardTitle>
        <p className="text-sm text-muted-foreground">
          Backup: <span className="font-mono">{backup.path}</span> on{" "}
          {backup.hostName}
        </p>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          {error && (
            <div className="p-4 text-sm text-destructive bg-destructive/10 border border-destructive/20 rounded-lg">
              {error}
            </div>
          )}

          <div className="space-y-3">
            <div className="space-y-1">
              <Label htmlFor="path" className="text-sm font-semibold">
                Specific Path to Restore
              </Label>
              <p className="text-xs text-muted-foreground">
                Enter <strong>/</strong> to restore the entire backup, or
                specify a subfolder or file (e.g., <code>uploads/</code> or{" "}
                <code>config.json</code>).
              </p>
            </div>
            <Input
              id="path"
              placeholder="/"
              className="bg-muted border-input transition-colors"
              {...register("path")}
            />
            {errors.path && (
              <p className="text-sm text-red-500">{errors.path.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label className="text-sm font-semibold">Restore Type</Label>
            <div className="flex gap-4">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  value="remote"
                  {...register("restore_type")}
                  className="w-4 h-4 text-primary"
                />
                <span className="text-sm">Remote (Rsync to host)</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  value="local"
                  {...register("restore_type")}
                  className="w-4 h-4 text-primary"
                />
                <span className="text-sm">Local (via CLI)</span>
              </label>
            </div>
          </div>

          {restoreType === "remote" && (
            <div className="space-y-6 border-l-2 border-primary/10 pl-4 mt-4">
              <div className="space-y-2">
                <Label
                  htmlFor="target_host_id"
                  className="text-sm font-semibold"
                >
                  Target Host
                </Label>
                {loadingHosts ? (
                  <Spinner className="h-4 w-4" />
                ) : (
                  <select
                    id="target_host_id"
                    className="w-full rounded-md border border-input bg-muted px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                    {...register("target_host_id")}
                  >
                    {hosts.map((h) => (
                      <option key={h.id} value={h.id}>
                        {h.name} ({h.hostname})
                      </option>
                    ))}
                  </select>
                )}
              </div>

              <div className="space-y-2">
                <Label htmlFor="target_path" className="text-sm font-semibold">
                  Target Path on Host
                </Label>
                <Input
                  id="target_path"
                  placeholder="/tmp/restore"
                  className="bg-muted border-input"
                  {...register("target_path")}
                />
                {errors.target_path && (
                  <p className="text-sm text-red-500">
                    {errors.target_path.message}
                  </p>
                )}
              </div>
            </div>
          )}

          {restoreType === "local" && (
            <div className="p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg text-blue-600 dark:text-blue-400 text-sm">
              <p className="font-bold mb-1">Local Restoration Instructions:</p>
              <p>
                Run the following command on your local machine to receive the
                files:
              </p>
              <code className="block mt-2 p-2 bg-muted border border-border rounded font-mono text-xs text-foreground">
                justbackup restore {backup.id} --local --path / --dest
                ./restored
              </code>
              <p className="mt-2 text-xs opacity-80 italic">
                * Path inside backup will be used from the command line
                argument.
              </p>
            </div>
          )}

          <Button
            type="submit"
            className="w-full"
            disabled={isSubmitting || (restoreType === "local" && true)}
          >
            {isSubmitting
              ? "Triggering..."
              : restoreType === "local"
                ? "Use CLI for Local Restore"
                : "Start Remote Restoration"}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
