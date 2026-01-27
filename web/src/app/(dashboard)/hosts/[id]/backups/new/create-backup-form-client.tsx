"use client";

import { useEffect, useState } from "react";
import { CreateBackupForm } from "@/backup/components/create-backup-form";
import { getHost, HostResponse } from "@/host/infrastructure/host-api";

interface CreateBackupFormClientProps {
  hostId: string;
  backupRoot?: string;
}

export function CreateBackupFormClient({
  hostId,
  backupRoot,
}: CreateBackupFormClientProps) {
  const [host, setHost] = useState<HostResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchHost = async () => {
      try {
        const hostData = await getHost(hostId);
        setHost(hostData);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load host");
      } finally {
        setIsLoading(false);
      }
    };

    fetchHost();
  }, [hostId]);

  if (isLoading) {
    return <div className="text-center p-4">Loading...</div>;
  }

  if (error) {
    return <div className="text-red-500 p-4">Error: {error}</div>;
  }

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold text-center">Add New Backup</h1>
      <CreateBackupForm
        hostId={hostId}
        backupRoot={backupRoot}
        hostPath={host?.path}
      />
    </div>
  );
}
