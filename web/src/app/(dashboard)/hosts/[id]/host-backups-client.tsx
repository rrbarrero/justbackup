"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { getBackups, BackupResponse } from "@/backup/infrastructure/backup-api";
import { getHost, HostResponse } from "@/host/infrastructure/host-api";
import { BackupsTable } from "@/backup/components/backups-table";
import { Button } from "@/shared/ui/button";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/shared/ui/breadcrumb";

interface HostBackupsClientProps {
  hostId: string;
  backupRoot?: string;
}

export function HostBackupsClient({
  hostId,
  backupRoot,
}: HostBackupsClientProps) {
  const router = useRouter();
  const [backups, setBackups] = useState<BackupResponse[]>([]);
  const [host, setHost] = useState<HostResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      if (!hostId) return;

      try {
        const [backupsData, hostData] = await Promise.all([
          getBackups(hostId),
          getHost(hostId),
        ]);
        setBackups(backupsData);
        setHost(hostData);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load data");
      } finally {
        setIsLoading(false);
      }
    };

    fetchData();
  }, [hostId]);

  if (isLoading) {
    return <div className="text-center p-4">Loading...</div>;
  }

  if (error) {
    return <div className="text-red-500 p-4">Error: {error}</div>;
  }

  return (
    <div className="space-y-6">
      <Breadcrumb>
        <BreadcrumbList>
          <BreadcrumbItem>
            <BreadcrumbLink href="/hosts">Hosts</BreadcrumbLink>
          </BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>
            <BreadcrumbPage>{host?.name || hostId}</BreadcrumbPage>
          </BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>

      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold">Host Backups</h1>
        <Button onClick={() => router.push(`/hosts/${hostId}/backups/new`)}>
          Add Backup
        </Button>
      </div>
      <BackupsTable
        backups={backups}
        backupRoot={backupRoot}
        hostPath={host?.path}
        onRefresh={() => {
          const fetchBackups = async () => {
            if (!hostId) return;
            try {
              const data = await getBackups(hostId);
              setBackups(data);
            } catch (err) {
              console.error(err);
            }
          };
          fetchBackups();
        }}
      />
    </div>
  );
}
