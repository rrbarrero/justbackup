"use client";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/shared/ui/alert-dialog";
import { deleteBackupErrors } from "@/backup/infrastructure/backup-api";
import { useState } from "react";

interface DeleteBackupLogsDialogProps {
  backupId: string | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function DeleteBackupLogsDialog({
  backupId,
  open,
  onOpenChange,
  onSuccess,
}: DeleteBackupLogsDialogProps) {
  const [isDeleting, setIsDeleting] = useState(false);

  const handleDelete = async () => {
    if (!backupId) return;

    setIsDeleting(true);
    try {
      await deleteBackupErrors(backupId);
      onOpenChange(false);
      onSuccess?.();
    } catch (error) {
      console.error("Failed to delete backup logs:", error);
      // Ideally show a toast notification here
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Clear Error Logs?</AlertDialogTitle>
          <AlertDialogDescription>
            This action cannot be undone. This will permanently delete all error
            logs for this backup task.
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isDeleting}>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={(e: React.MouseEvent<HTMLButtonElement>) => {
              e.preventDefault();
              handleDelete();
            }}
            disabled={isDeleting}
            className="bg-red-600 hover:bg-red-700 focus:ring-red-600"
          >
            {isDeleting ? "Clearing..." : "Clear Logs"}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
