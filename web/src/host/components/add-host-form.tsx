"use client";

import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { useRouter } from "next/navigation";
import { createHost } from "@/host/infrastructure/host-api";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/shared/ui/card";

import { Switch } from "@/shared/ui/switch";

const hostSchema = z.object({
  name: z.string().min(1, "Name is required"),
  hostname: z.string().min(1, "Hostname is required"),
  user: z.string().min(1, "User is required"),
  port: z.coerce
    .number()
    .int()
    .positive("Port must be a positive integer")
    .default(22),
  path: z.string().min(1, "Path is required"),
  is_workstation: z.boolean().default(false),
});

type HostFormValues = z.infer<typeof hostSchema>;

export function AddHostForm() {
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
    watch,
    setValue,
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

  const nameValue = watch("name");

  useEffect(() => {
    if (nameValue) {
      const slug = nameValue
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, "-")
        .replace(/(^-|-$)+/g, "");
      setValue("path", slug, { shouldDirty: false, shouldTouch: false });
    }
  }, [nameValue, setValue]);

  const onSubmit = async (data: HostFormValues) => {
    setIsLoading(true);
    setError(null);

    try {
      await createHost(data);
      router.push("/hosts"); // Redirect to hosts list (to be implemented)
    } catch (err) {
      setError(err instanceof Error ? err.message : "An error occurred");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Card className="w-full max-w-2xl mx-auto">
      <CardHeader>
        <CardTitle>Add New Host</CardTitle>
        <CardDescription>
          Configure a new remote host for backups.
        </CardDescription>
      </CardHeader>
      <form onSubmit={handleSubmit(onSubmit)}>
        <CardContent className="space-y-4">
          {error && (
            <div className="p-3 text-sm text-red-500 bg-red-50 rounded-md">
              {error}
            </div>
          )}
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input id="name" placeholder="My Server" {...register("name")} />
              {errors.name && (
                <p className="text-sm text-red-500">{errors.name.message}</p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="path">Host Path (Slug)</Label>
              <Input id="path" placeholder="my-server" {...register("path")} />
              {errors.path && (
                <p className="text-sm text-red-500">{errors.path.message}</p>
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
                <p className="text-sm text-red-500">
                  {errors.hostname.message}
                </p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="user">SSH User</Label>
              <Input id="user" placeholder="root" {...register("user")} />
              {errors.user && (
                <p className="text-sm text-red-500">{errors.user.message}</p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="port">SSH Port</Label>
              <Input id="port" type="number" {...register("port")} />
              {errors.port && (
                <p className="text-sm text-red-500">{errors.port.message}</p>
              )}
            </div>
            <div className="flex flex-row items-center justify-between rounded-lg border p-4 shadow-sm col-span-1 md:col-span-2">
              <div className="space-y-0.5">
                <Label htmlFor="is_workstation">Workstation Mode</Label>
                <div className="text-sm text-muted-foreground">
                  Enable this if the host is not always online. Backup failures
                  will be ignored.
                </div>
              </div>
              <Switch
                id="is_workstation"
                checked={watch("is_workstation")}
                onCheckedChange={(checked) =>
                  setValue("is_workstation", checked)
                }
              />
            </div>
          </div>
        </CardContent>
        <CardFooter>
          <Button type="submit" disabled={isLoading}>
            {isLoading ? "Adding Host..." : "Add Host"}
          </Button>
        </CardFooter>
      </form>
    </Card>
  );
}
