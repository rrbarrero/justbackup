"use client";

import { useState, useEffect } from "react";
import { useForm, Controller, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { useRouter } from "next/navigation";
import {
  updateBackup,
  measureSize,
  getTaskResult,
  BackupResponse,
} from "@/backup/infrastructure/backup-api";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { Switch } from "@/shared/ui/switch";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/shared/ui/dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/shared/ui/tabs";
import { computeBackupFullPath } from "@/backup/utils/backup-path";
import { BackupDestinationHelper } from "@/backup/components/backup-destination-helper";
import { HookManager } from "./hook-manager";

const backupSchema = z.object({
  path: z.string().min(1, "Path is required"),
  destination: z.string().min(1, "Destination is required"),
  schedule: z
    .string()
    .min(1, "Schedule is required")
    .regex(
      /^(@(annually|yearly|monthly|weekly|daily|midnight|hourly))|(((\*|([0-9\-,/]+)) +){4}(\*|([0-9\-,/]+)))$/,
      "Invalid cron expression (e.g. '0 0 * * *' or '@daily')",
    ),
  excludes: z.string().optional(),
  incremental: z.boolean(),
  retentionCount: z.number().min(1, "Retention must be at least 1"),
  encrypted: z.boolean(),
  hooks: z.array(
    z.object({
      name: z.string().min(1, "Hook name is required"),
      phase: z.enum(["pre", "post"]),
      enabled: z.boolean(),
      params: z.record(z.string(), z.string()),
    }),
  ),
  useEphemeralSource: z.boolean().optional(),
});

type BackupFormValues = z.infer<typeof backupSchema>;

interface EditBackupDialogProps {
  backup: BackupResponse | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
  backupRoot?: string;
  hostPath?: string;
}

export function EditBackupDialog({
  backup,
  open,
  onOpenChange,
  onSuccess,
  backupRoot,
  hostPath,
}: EditBackupDialogProps) {
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [measuredSize, setMeasuredSize] = useState<string | null>(null);
  const [isMeasuring, setIsMeasuring] = useState(false);

  const {
    register,
    handleSubmit,
    reset,
    control,
    setValue,
    getValues,
    formState: { errors },
  } = useForm<BackupFormValues>({
    resolver: zodResolver(backupSchema),
    defaultValues: {
      path: "",
      destination: "",
      schedule: "",
      excludes: "",
      incremental: false,
      retentionCount: 4,
      encrypted: false,
      hooks: [],
      useEphemeralSource: false,
    },
  });

  const useEphemeralSource = useWatch({
    control,
    name: "useEphemeralSource",
  });
  const incremental = useWatch({ control, name: "incremental" });

  const EPHEMERAL_PATH = "{{SESSION_TEMP_DIR}}";

  // Effect to handle path update when ephemeral source toggles
  useEffect(() => {
    // Only run this effect if dirty or touched to avoid overriding initial data loading?
    // Actually, we need to be careful not to override the path when loading the form.
    // The form loading happens in the OTHER useEffect.
    // This effect reacts to user changing the switch.
    // We can check if isSubmitting or something?
    // Better: simplistic approach:
    // If switch is ON and path != placeholder -> set placeholder
    // If switch is OFF and path == placeholder -> clear path

    // However, this needs to NOT run on initial mount if we want to respect default check.
    // But since we control the switch, it should be fine.

    if (useEphemeralSource && getValues("path") !== EPHEMERAL_PATH) {
      setValue("path", EPHEMERAL_PATH, { shouldValidate: true });
    } else if (!useEphemeralSource && getValues("path") === EPHEMERAL_PATH) {
      // Only clear if it was the placeholder
      setValue("path", "", { shouldValidate: true });
    }
  }, [useEphemeralSource, setValue, getValues]);

  useEffect(() => {
    if (backup) {
      reset({
        path: backup.path,
        destination: backup.destination,
        schedule: backup.schedule,
        excludes: backup.excludes ? backup.excludes.join(",") : "",
        incremental: backup.incremental,
        retentionCount: backup.retention || 4,
        encrypted: backup.encrypted,
        hooks: (backup.hooks || []).map((h) => ({
          name: h.name,
          phase: (h.phase === "post" ? "post" : "pre") as "pre" | "post",
          enabled: h.enabled,
          params: h.params || {},
        })),
        useEphemeralSource: backup.path === EPHEMERAL_PATH,
      });
    }
  }, [backup, reset]);

  const onSubmit = async (data: BackupFormValues) => {
    if (!backup) return;

    setIsSubmitting(true);
    setError(null);

    try {
      const excludes = (data.excludes || "")
        .split(",")
        .map((item) => item.trim())
        .filter(Boolean);

      await updateBackup({
        id: backup.id,
        path: data.path,
        destination: data.destination,
        schedule: data.schedule,
        excludes,
        incremental: data.incremental,
        retention: data.retentionCount,
        encrypted: data.encrypted,
        hooks: data.hooks,
      });
      onOpenChange(false);
      router.refresh();
      if (onSuccess) {
        onSuccess();
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update backup");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handlePathBlur = async (e: React.FocusEvent<HTMLInputElement>) => {
    const path = e.target.value;
    if (!path || !backup) return;

    setIsMeasuring(true);
    setMeasuredSize(null);

    try {
      const { task_id } = await measureSize(backup.hostId, path);

      // Poll for result
      const poll = setInterval(async () => {
        try {
          const result = await getTaskResult(task_id);
          if (result) {
            clearInterval(poll);
            if (result.data && result.data.size) {
              setMeasuredSize(result.data.size);
            }
            setIsMeasuring(false);
          }
        } catch (err) {
          // Ignore polling errors
        }
      }, 1000);

      // Stop polling after 30 seconds
      setTimeout(() => {
        clearInterval(poll);
        if (isMeasuring) setIsMeasuring(false);
      }, 30000);
    } catch (err) {
      console.error("Failed to measure size", err);
      setIsMeasuring(false);
    }
  };

  const destinationValue = useWatch({ control, name: "destination" }) || "";
  const fullPath = computeBackupFullPath(
    backupRoot,
    hostPath,
    destinationValue,
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-4xl max-h-[95vh] overflow-hidden flex flex-col p-0 border-none shadow-2xl">
        <DialogHeader className="p-6 bg-muted/30 border-b">
          <DialogTitle className="text-2xl font-bold">
            Edit Backup Task
          </DialogTitle>
          <p className="text-muted-foreground">
            Modify your backup configuration and hooks.
          </p>
        </DialogHeader>

        <form
          onSubmit={handleSubmit(onSubmit)}
          className="flex-1 flex flex-col min-h-0"
        >
          <div className="flex-1 overflow-y-auto p-6">
            {error && (
              <div className="mb-6 p-4 text-sm text-red-500 bg-red-50 border border-red-100 rounded-lg">
                {error}
              </div>
            )}

            <Tabs defaultValue="general" className="w-full">
              <TabsList className="grid w-full grid-cols-3 mb-8 bg-muted/50 p-1 rounded-xl">
                <TabsTrigger
                  value="general"
                  className="rounded-lg data-[state=active]:bg-card data-[state=active]:shadow-sm py-2.5"
                >
                  General Settings
                </TabsTrigger>
                <TabsTrigger
                  value="advanced"
                  className="rounded-lg data-[state=active]:bg-card data-[state=active]:shadow-sm py-2.5"
                >
                  Advanced Settings
                </TabsTrigger>
                <TabsTrigger
                  value="hooks"
                  className="rounded-lg data-[state=active]:bg-card data-[state=active]:shadow-sm py-2.5"
                >
                  Backup Hooks
                </TabsTrigger>
              </TabsList>

              <TabsContent value="general" className="space-y-6 mt-0">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-x-8 gap-y-6 bg-card p-6 rounded-xl border border-border/50 shadow-sm">
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label
                        htmlFor="path-edit"
                        className="text-sm font-semibold"
                      >
                        Source Path
                      </Label>
                      <Input
                        id="path-edit"
                        placeholder="/var/www/html"
                        disabled={useEphemeralSource}
                        className="bg-muted border-input transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                        {...register("path")}
                        onBlur={(e) => {
                          register("path").onBlur(e);
                          handlePathBlur(e);
                        }}
                      />
                      {isMeasuring && (
                        <p className="text-xs text-blue-500 animate-pulse">
                          Measuring size...
                        </p>
                      )}
                      {measuredSize && (
                        <p className="text-xs text-green-600 dark:text-green-500 font-medium">
                          Estimated size: {measuredSize}
                        </p>
                      )}
                      {errors.path && (
                        <p className="text-sm text-red-500">
                          {errors.path.message}
                        </p>
                      )}
                    </div>

                    <div className="flex items-start space-x-3 border p-3 rounded-lg bg-yellow-500/10 border-yellow-500/20">
                      <div className="pt-0.5">
                        <Controller
                          control={control}
                          name="useEphemeralSource"
                          render={({ field }) => (
                            <Switch
                              checked={field.value}
                              onCheckedChange={field.onChange}
                              id="useEphemeralSource-edit"
                            />
                          )}
                        />
                      </div>
                      <div className="space-y-1">
                        <Label
                          htmlFor="useEphemeralSource-edit"
                          className="text-sm font-semibold cursor-pointer"
                        >
                          Use Ephemeral Source
                        </Label>
                        <p className="text-xs text-muted-foreground">
                          Source will be created at runtime by a hook and
                          deleted afterwards (e.g. a database dump).
                        </p>
                      </div>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <Label
                      htmlFor="destination-edit"
                      className="text-sm font-semibold"
                    >
                      Destination Name
                    </Label>
                    <Input
                      id="destination-edit"
                      placeholder="myapp"
                      className="bg-muted/30 border-muted-foreground/20 focus:bg-white transition-colors"
                      {...register("destination")}
                      onChange={(e) => {
                        register("destination").onChange(e);
                        setValue("destination", e.target.value, {
                          shouldValidate: true,
                        });
                      }}
                    />
                    {errors.destination && (
                      <p className="text-sm text-red-500">
                        {errors.destination.message}
                      </p>
                    )}
                    <BackupDestinationHelper fullPath={fullPath} />
                  </div>

                  <div className="space-y-2">
                    <Label
                      htmlFor="schedule-edit"
                      className="text-sm font-semibold"
                    >
                      Schedule (Cron)
                    </Label>
                    <Input
                      id="schedule-edit"
                      placeholder="0 0 * * *"
                      className="bg-muted/30 border-muted-foreground/20 focus:bg-white transition-colors font-mono"
                      {...register("schedule")}
                    />
                    {errors.schedule && (
                      <p className="text-sm text-red-500">
                        {errors.schedule.message}
                      </p>
                    )}
                    <p className="text-[11px] text-muted-foreground uppercase tracking-wider font-medium">
                      Format: Minutes Hours Day Month DayOfWeek
                    </p>
                  </div>

                  <div className="space-y-2">
                    <Label
                      htmlFor="excludes-edit"
                      className="text-sm font-semibold"
                    >
                      Exclude Patterns
                    </Label>
                    <Input
                      id="excludes-edit"
                      placeholder="tmp/,*.log,node_modules/"
                      className="bg-muted/30 border-muted-foreground/20 focus:bg-white transition-colors"
                      {...register("excludes")}
                    />
                    <p className="text-xs text-muted-foreground italic">
                      Comma-separated rsync exclusions (e.g.{" "}
                      <code>tmp/*, *.log</code>)
                    </p>
                  </div>
                </div>
              </TabsContent>

              <TabsContent value="advanced" className="space-y-6 mt-0">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                  <div className="space-y-6 bg-card p-6 rounded-xl border border-border/50 shadow-sm transition-all hover:border-primary/20">
                    <div className="flex items-start space-x-4">
                      <div className="pt-1">
                        <Controller
                          control={control}
                          name="incremental"
                          render={({ field }) => (
                            <Switch
                              checked={field.value}
                              onCheckedChange={field.onChange}
                              id="incremental-edit"
                            />
                          )}
                        />
                      </div>
                      <div className="space-y-1">
                        <Label
                          htmlFor="incremental-edit"
                          className="text-base font-bold"
                        >
                          Incremental Backup
                        </Label>
                        <p className="text-sm text-muted-foreground leading-relaxed">
                          If enabled, backups will use hard links to save space.
                          Previous backups are required for this to work
                          effectively.
                        </p>
                      </div>
                    </div>

                    {incremental && (
                      <div className="space-y-3 pl-14 pt-2 border-l-2 border-primary/10 ml-6">
                        <Label
                          htmlFor="retentionCount-edit"
                          className="text-sm font-semibold"
                        >
                          Retention Count
                        </Label>
                        <div className="flex items-center space-x-3">
                          <Input
                            id="retentionCount-edit"
                            type="number"
                            min="1"
                            className="max-w-[100px] h-10"
                            {...register("retentionCount", {
                              valueAsNumber: true,
                            })}
                          />
                          <span className="text-sm text-muted-foreground font-medium">
                            backups to keep
                          </span>
                        </div>
                        {errors.retentionCount && (
                          <p className="text-sm text-red-500">
                            {errors.retentionCount.message}
                          </p>
                        )}
                      </div>
                    )}
                  </div>

                  <div className="space-y-6 bg-card p-6 rounded-xl border border-border/50 shadow-sm transition-all hover:border-primary/20">
                    <div className="flex items-start space-x-4">
                      <div className="pt-1">
                        <Controller
                          control={control}
                          name="encrypted"
                          render={({ field }) => (
                            <Switch
                              checked={field.value}
                              onCheckedChange={field.onChange}
                              id="encrypted-edit"
                            />
                          )}
                        />
                      </div>
                      <div className="space-y-1">
                        <Label
                          htmlFor="encrypted-edit"
                          className="text-base font-bold"
                        >
                          Encrypt Backup
                        </Label>
                        <p className="text-sm text-muted-foreground leading-relaxed">
                          The backup will be compressed and encrypted using
                          AES-GCM for maximum data protection.
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
              </TabsContent>

              <TabsContent value="hooks" className="mt-0">
                <div className="bg-card p-6 rounded-xl border border-border/50 shadow-sm">
                  <HookManager
                    control={control}
                    register={register}
                    setValue={setValue}
                  />
                </div>
              </TabsContent>
            </Tabs>
          </div>

          <DialogFooter className="p-6 bg-muted/30 border-t flex items-center space-x-4">
            <Button
              type="button"
              variant="ghost"
              size="lg"
              className="px-8 text-muted-foreground hover:bg-muted"
              onClick={() => onOpenChange(false)}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              size="lg"
              disabled={isSubmitting}
              className="px-10 h-12 text-base font-bold shadow-lg shadow-primary/20 transition-all hover:-translate-y-0.5"
            >
              {isSubmitting ? "Saving..." : "Save Changes"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
