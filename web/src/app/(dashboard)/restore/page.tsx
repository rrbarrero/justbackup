"use client";

import { useState, useEffect } from "react";
import { HostResponse, getHosts } from "@/host/infrastructure/host-api";
import { BackupResponse, getBackups } from "@/backup/infrastructure/backup-api";
import { RestoreForm } from "@/backup/components/restore-form";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";
import { Button } from "@/shared/ui/button";
import { Label } from "@/shared/ui/label";
import { Spinner } from "@/shared/ui/spinner";
import { RotateCcw, Server, Archive } from "lucide-react";

export default function RestorePage() {
  const [hosts, setHosts] = useState<HostResponse[]>([]);
  const [backups, setBackups] = useState<BackupResponse[]>([]);
  const [selectedHostId, setSelectedHostId] = useState<string>("");
  const [selectedBackupId, setSelectedBackupId] = useState<string>("");
  const [loadingHosts, setLoadingHosts] = useState(true);
  const [loadingBackups, setLoadingBackups] = useState(false);
  const [restorationTaskId, setRestorationTaskId] = useState<string | null>(
    null,
  );
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    getHosts()
      .then(setHosts)
      .catch((err) => setError("Failed to load hosts"))
      .finally(() => setLoadingHosts(false));
  }, []);

  useEffect(() => {
    if (selectedHostId) {
      setLoadingBackups(true);
      setBackups([]);
      setSelectedBackupId("");
      getBackups(selectedHostId)
        .then(setBackups)
        .catch((err) => setError("Failed to load backups"))
        .finally(() => setLoadingBackups(false));
    }
  }, [selectedHostId]);

  const selectedBackup = backups.find((b) => b.id === selectedBackupId);

  return (
    <div className="space-y-6 max-w-4xl mx-auto">
      <div className="flex flex-col gap-2">
        <h1 className="text-3xl font-bold flex items-center gap-2">
          <RotateCcw className="h-8 w-8 text-primary" />
          Restoration Manager
        </h1>
        <p className="text-muted-foreground">
          Select a host and a backup to start the restoration process.
        </p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card className="bg-card shadow-sm border-border">
          <CardHeader>
            <CardTitle className="text-lg font-bold flex items-center gap-2">
              <Server className="h-5 w-5 text-primary" />
              1. Select Source Host
            </CardTitle>
          </CardHeader>
          <CardContent>
            {loadingHosts ? (
              <Spinner className="h-6 w-6 mx-auto" />
            ) : (
              <div className="space-y-2">
                <Label htmlFor="host-select">Host</Label>
                <select
                  id="host-select"
                  className="w-full rounded-md border border-input bg-muted px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                  value={selectedHostId}
                  onChange={(e) => setSelectedHostId(e.target.value)}
                >
                  <option value="">Select a host...</option>
                  {hosts.map((h) => (
                    <option key={h.id} value={h.id}>
                      {h.name} ({h.hostname})
                    </option>
                  ))}
                </select>
              </div>
            )}
          </CardContent>
        </Card>

        <Card className="bg-card shadow-sm border-border">
          <CardHeader>
            <CardTitle className="text-lg font-bold flex items-center gap-2">
              <Archive className="h-5 w-5 text-primary" />
              2. Select Backup
            </CardTitle>
          </CardHeader>
          <CardContent>
            {!selectedHostId ? (
              <p className="text-sm text-muted-foreground text-center py-4">
                Please select a host first.
              </p>
            ) : loadingBackups ? (
              <Spinner className="h-6 w-6 mx-auto" />
            ) : backups.length === 0 ? (
              <p className="text-sm text-yellow-500 text-center py-4 bg-yellow-500/10 border border-yellow-500/20 rounded-lg">
                No backups found for this host.
              </p>
            ) : (
              <div className="space-y-2">
                <Label htmlFor="backup-select">Backup Task</Label>
                <select
                  id="backup-select"
                  className="w-full rounded-md border border-input bg-muted px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                  value={selectedBackupId}
                  onChange={(e) => setSelectedBackupId(e.target.value)}
                >
                  <option value="">Select a backup task...</option>
                  {backups.map((b) => (
                    <option key={b.id} value={b.id}>
                      {b.path} (→ {b.destination})
                    </option>
                  ))}
                </select>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {selectedBackup && !restorationTaskId && (
        <div className="animate-in fade-in slide-in-from-top-4 duration-500">
          <RestoreForm
            backup={selectedBackup}
            onSuccess={(taskId) => setRestorationTaskId(taskId)}
          />
        </div>
      )}

      {restorationTaskId && (
        <Card className="bg-green-500/5 border-green-500/20 shadow-sm">
          <CardHeader>
            <CardTitle className="text-green-600 dark:text-green-500 flex items-center gap-2">
              <div className="h-2 w-2 bg-green-500 rounded-full animate-pulse" />
              Restoration Task Triggered
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-green-700 dark:text-green-400">
              Your restoration task has been successfully triggered and is being
              processed by the worker.
            </p>
            <div className="p-3 bg-muted border border-border rounded-lg font-mono text-sm break-all">
              Task ID: {restorationTaskId}
            </div>
            <Button
              variant="outline"
              className="w-full border-green-500/20 text-green-700 dark:text-green-400 hover:bg-green-500/10"
              onClick={() => {
                setRestorationTaskId(null);
                setSelectedBackupId("");
              }}
            >
              Start Another Restoration
            </Button>
          </CardContent>
        </Card>
      )}

      {error && (
        <div className="p-4 text-sm text-destructive bg-destructive/10 border border-destructive/20 rounded-lg">
          {error}
        </div>
      )}
    </div>
  );
}
