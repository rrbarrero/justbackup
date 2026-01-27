"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { BackupResponse, getBackups } from "@/backup/infrastructure/backup-api";
import { BackupsTable } from "@/backup/components/backups-table";
import { Spinner } from "@/shared/ui/spinner";

function BackupsContent() {
  const searchParams = useSearchParams();
  const status = searchParams.get("status") || undefined;

  const [backups, setBackups] = useState<BackupResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchBackups = () => {
    setLoading(true);
    getBackups(undefined, status)
      .then(setBackups)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchBackups();
  }, [status]);

  if (loading) {
    return (
      <div className="flex justify-center p-8">
        <Spinner className="h-8 w-8" />
      </div>
    );
  }

  if (error) {
    return <div className="p-8 text-center text-red-500">Error: {error}</div>;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">
          {status
            ? `${status.charAt(0).toUpperCase() + status.slice(1)} Backups`
            : "All Backups"}
        </h1>
      </div>

      <BackupsTable backups={backups} onRefresh={fetchBackups} />
    </div>
  );
}

export default function BackupsPage() {
  return (
    <Suspense
      fallback={
        <div className="flex justify-center p-8">
          <Spinner className="h-8 w-8" />
        </div>
      }
    >
      <BackupsContent />
    </Suspense>
  );
}
