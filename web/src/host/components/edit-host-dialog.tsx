"use client";

import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { useRouter } from "next/navigation";
import { updateHost, HostResponse } from "@/host/infrastructure/host-api";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/shared/ui/dialog";

import { Switch } from "@/shared/ui/switch";

const hostSchema = z.object({
  name: z.string().min(1, "Name is required"),
  hostname: z.string().min(1, "Hostname is required"),
  user: z.string().min(1, "User is required"),
  port: z.number().min(1, "Port must be greater than 0"),
  path: z.string().min(1, "Path is required"),
  is_workstation: z.boolean().default(false),
});

type HostFormValues = z.infer<typeof hostSchema>;

interface EditHostDialogProps {
  host: HostResponse | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function EditHostDialog({
  host,
  open,
  onOpenChange,
  onSuccess,
}: EditHostDialogProps) {
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const {
    register,
    handleSubmit,
    reset,
    watch,
    setValue,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(hostSchema),
    defaultValues: {
      name: "",
      hostname: "",
      user: "",
      port: 22,
      path: "",
      is_workstation: false,
    },
  });

  useEffect(() => {
    if (host) {
      reset({
        name: host.name,
        hostname: host.hostname,
        user: host.user,
        port: host.port,
        path: host.path,
        is_workstation: host.isWorkstation,
      });
    }
  }, [host, reset]);

  const onSubmit = async (data: HostFormValues) => {
    if (!host) return;

    setIsSubmitting(true);
    setError(null);

    try {
      await updateHost(host.id, {
        ...data,
      });
      onOpenChange(false);
      router.refresh();
      if (onSuccess) {
        onSuccess();
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update host");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Edit Host</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          {error && (
            <div className="p-3 text-sm text-red-500 bg-red-50 rounded-md">
              {error}
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input id="name" placeholder="My Server" {...register("name")} />
            {errors.name && (
              <p className="text-sm text-red-500">{errors.name.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="hostname">Hostname / IP</Label>
            <Input
              id="hostname"
              placeholder="192.168.1.100"
              {...register("hostname")}
            />
            {errors.hostname && (
              <p className="text-sm text-red-500">{errors.hostname.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="user">User</Label>
            <Input id="user" placeholder="root" {...register("user")} />
            {errors.user && (
              <p className="text-sm text-red-500">{errors.user.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="port">Port</Label>
            <Input
              id="port"
              type="number"
              placeholder="22"
              {...register("port", { valueAsNumber: true })}
            />
            {errors.port && (
              <p className="text-sm text-red-500">{errors.port.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="path">Host Path</Label>
            <Input id="path" placeholder="/backup/root" {...register("path")} />
            {errors.path && (
              <p className="text-sm text-red-500">{errors.path.message}</p>
            )}
          </div>

          <div className="flex flex-row items-center justify-between rounded-lg border p-4 shadow-sm">
            <div className="space-y-0.5">
              <Label htmlFor="is_workstation">Workstation Mode</Label>
              <div className="text-sm text-muted-foreground">
                Enable if host is not always online.
              </div>
            </div>
            <Switch
              id="is_workstation"
              checked={watch("is_workstation")}
              onCheckedChange={(checked) => setValue("is_workstation", checked)}
            />
          </div>

          <div className="flex justify-end space-x-2 pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isSubmitting}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Saving..." : "Save Changes"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
