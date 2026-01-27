"use client";

import { useEffect, useState } from "react";
import {
  BackupResponse,
  BackupErrorResponse,
  getBackupErrors,
} from "@/backup/infrastructure/backup-api";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/shared/ui/sheet";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/shared/ui/table";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";
import { Alert, AlertDescription } from "@/shared/ui/alert";
import { Spinner } from "@/shared/ui/spinner";
import { Button } from "@/shared/ui/button";
import {
  InfoIcon,
  FolderIcon,
  CalendarIcon,
  CheckCircle2,
  XCircle,
  Clock,
  AlertTriangle,
  Trash2,
} from "lucide-react";
import cronstrue from "cronstrue";
import { DeleteBackupLogsDialog } from "./delete-backup-logs-dialog";
import { formatSize } from "@/shared/lib/utils";

interface BackupDetailsSheetProps {
  backup: BackupResponse | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  backupRoot?: string;
  hostPath?: string;
}

export function BackupDetailsSheet({
  backup,
  open,
  onOpenChange,
  backupRoot,
  hostPath,
}: BackupDetailsSheetProps) {
  const [errors, setErrors] = useState<BackupErrorResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);

  const fetchErrors = () => {
    if (backup) {
      setLoading(true);
      setErrorMessage(null);
      getBackupErrors(backup.id)
        .then((data) => setErrors(Array.isArray(data) ? data : []))
        .catch((err) => setErrorMessage(err.message))
        .finally(() => setLoading(false));
    }
  };

  useEffect(() => {
    if (backup && open) {
      fetchErrors();
    } else {
      setErrors([]);
      setErrorMessage(null);
    }
  }, [backup, open]);

  if (!backup) return null;

  const formatSchedule = () => {
    try {
      return cronstrue.toString(backup.schedule);
    } catch {
      return backup.schedule;
    }
  };

  const getFullDestination = () => {
    if (!backupRoot) return backup.destination;

    const cleanRoot = backupRoot.replace(/\/+$/, "");
    const cleanHostPath = hostPath ? hostPath.replace(/(^\/+|\/+$)/g, "") : "";
    const cleanDest = backup.destination.replace(/^\/+/, "");

    const parts = [cleanRoot];
    if (cleanHostPath) parts.push(cleanHostPath);
    parts.push(cleanDest);

    return parts.join("/");
  };

  const getStatusIcon = () => {
    switch (backup.status) {
      case "completed":
        return <CheckCircle2 className="h-5 w-5 text-green-600" />;
      case "failed":
        return <XCircle className="h-5 w-5 text-red-600" />;
      case "running":
        return <Clock className="h-5 w-5 text-blue-600" />;
      default:
        return <AlertTriangle className="h-5 w-5 text-yellow-600" />;
    }
  };

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="sm:max-w-2xl overflow-y-auto">
        <SheetHeader className="pb-6">
          <SheetTitle className="text-2xl">Backup Details</SheetTitle>
          <SheetDescription>
            Detailed information and error logs for this backup task
          </SheetDescription>
        </SheetHeader>

        <div className="space-y-6">
          {/* Status Card */}
          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center gap-3">
                {getStatusIcon()}
                <div className="flex-1">
                  <p className="text-sm text-muted-foreground">
                    Current Status
                  </p>
                  <p className="text-lg font-semibold capitalize">
                    {backup.status}
                  </p>
                </div>
                <div className="text-right">
                  <p className="text-sm text-muted-foreground">Last Run</p>
                  <p className="text-sm font-medium">
                    {new Date(backup.lastRun).getFullYear() === 1
                      ? "Never"
                      : new Date(backup.lastRun).toLocaleString()}
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Backup Configuration */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base flex items-center gap-2">
                <FolderIcon className="h-4 w-4" />
                Backup Configuration
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-3">
                <div className="flex flex-col gap-1">
                  <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                    Backup ID
                  </span>
                  <span className="text-sm font-medium font-mono text-xs text-muted-foreground select-all">
                    {backup.id}
                  </span>
                </div>

                <div className="grid grid-cols-2 gap-3">
                  <div className="flex flex-col gap-1">
                    <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                      Host Name
                    </span>
                    <span className="text-sm font-medium">
                      {backup.hostName}
                    </span>
                  </div>
                  <div className="flex flex-col gap-1">
                    <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                      Hostname / IP
                    </span>
                    <span className="text-sm font-medium font-mono">
                      {backup.hostAddress}
                    </span>
                  </div>
                </div>

                <div className="flex flex-col gap-1">
                  <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                    Source Path
                  </span>
                  <code className="text-sm bg-muted px-2 py-1 rounded font-mono">
                    {backup.path}
                  </code>
                </div>

                <div className="flex flex-col gap-1">
                  <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                    Destination
                  </span>
                  <code className="text-sm bg-muted px-2 py-1 rounded font-mono">
                    {getFullDestination()}
                  </code>
                </div>

                <div className="flex flex-col gap-1">
                  <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                    Size
                  </span>
                  <code className="text-sm bg-muted px-2 py-1 rounded font-mono">
                    {backup.size ? formatSize(backup.size) : "-"}
                  </code>
                </div>

                <div className="grid grid-cols-2 gap-3">
                  <div className="flex flex-col gap-1">
                    <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                      Incremental
                    </span>
                    <span className="text-sm font-medium">
                      {backup.incremental ? "Enabled" : "Disabled"}
                    </span>
                  </div>
                  <div className="flex flex-col gap-1">
                    <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                      Encrypted
                    </span>
                    <span className="text-sm font-medium">
                      {backup.encrypted ? "Yes" : "No"}
                    </span>
                  </div>
                </div>

                <div className="flex flex-col gap-1">
                  <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                    Excludes
                  </span>
                  <span className="text-sm font-medium">
                    {backup.excludes && backup.excludes.length > 0
                      ? `${backup.excludes.length} pattern${backup.excludes.length > 1 ? "s" : ""}`
                      : "None"}
                  </span>
                </div>

                {backup.excludes && backup.excludes.length > 0 && (
                  <div className="flex flex-col gap-2 pt-2 border-t">
                    <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                      Exclude Patterns
                    </span>
                    <div className="flex flex-wrap gap-1.5">
                      {backup.excludes.map((exclude, idx) => (
                        <code
                          key={idx}
                          className="text-xs bg-muted px-2 py-1 rounded font-mono"
                        >
                          {exclude}
                        </code>
                      ))}
                    </div>
                  </div>
                )}

                <div className="flex flex-col gap-1 pt-2 border-t">
                  <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                    Hooks
                  </span>
                  <div className="flex flex-wrap gap-2 mt-1">
                    {backup.hooks && backup.hooks.length > 0 ? (
                      backup.hooks.map((hook, idx) => (
                        <div
                          key={idx}
                          className="flex items-center gap-1.5 bg-muted px-2 py-1 rounded text-xs"
                        >
                          <span className="font-medium">{hook.name}</span>
                          <span className="text-muted-foreground text-[10px] uppercase border border-border px-1 rounded-sm bg-background">
                            {hook.phase}
                          </span>
                        </div>
                      ))
                    ) : (
                      <span className="text-sm font-medium text-muted-foreground">
                        None
                      </span>
                    )}
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Schedule Card */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base flex items-center gap-2">
                <CalendarIcon className="h-4 w-4" />
                Schedule
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex flex-col gap-1">
                <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                  Cron Expression
                </span>
                <code className="text-sm bg-muted px-2 py-1 rounded font-mono">
                  {backup.schedule}
                </code>
              </div>
              <div className="flex flex-col gap-1">
                <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                  Human Readable
                </span>
                <p className="text-sm">{formatSchedule()}</p>
              </div>
            </CardContent>
          </Card>

          {/* Error Logs */}
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <CardTitle className="text-base flex items-center gap-2">
                  <AlertTriangle className="h-4 w-4" />
                  Error Logs
                  {errors.length > 0 && (
                    <span className="ml-2 text-xs bg-red-100 text-red-800 px-2 py-0.5 rounded-full font-medium">
                      {errors.length} error{errors.length > 1 ? "s" : ""}
                    </span>
                  )}
                </CardTitle>
                {errors.length > 0 && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => setShowDeleteDialog(true)}
                    className="text-red-600 hover:text-red-700 hover:bg-red-50"
                  >
                    <Trash2 className="h-4 w-4 mr-1" />
                    Clear Logs
                  </Button>
                )}
              </div>
            </CardHeader>
            <CardContent>
              {loading && (
                <div className="flex items-center justify-center py-8">
                  <Spinner className="h-6 w-6" />
                </div>
              )}

              {!loading && errorMessage && (
                <Alert variant="destructive">
                  <XCircle className="h-4 w-4" />
                  <AlertDescription>{errorMessage}</AlertDescription>
                </Alert>
              )}

              {!loading && !errorMessage && errors.length === 0 && (
                <Alert>
                  <CheckCircle2 className="h-4 w-4 text-green-600" />
                  <AlertDescription>
                    No errors found for this backup task. Everything is running
                    smoothly!
                  </AlertDescription>
                </Alert>
              )}

              {!loading && !errorMessage && errors.length > 0 && (
                <div className="rounded-md border">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead className="w-[180px]">Time</TableHead>
                        <TableHead>Error Message</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {errors.map((error) => (
                        <TableRow key={error.id}>
                          <TableCell className="whitespace-nowrap text-xs text-muted-foreground">
                            {new Date(error.occurred_at).toLocaleString()}
                          </TableCell>
                          <TableCell className="font-mono text-xs">
                            {error.error_message}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        <DeleteBackupLogsDialog
          backupId={backup?.id || null}
          open={showDeleteDialog}
          onOpenChange={setShowDeleteDialog}
          onSuccess={fetchErrors}
        />
      </SheetContent>
    </Sheet>
  );
}
