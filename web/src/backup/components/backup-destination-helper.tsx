import { Alert, AlertDescription } from "@/shared/ui/alert";
import { InfoIcon } from "lucide-react";

interface BackupDestinationHelperProps {
  fullPath: string | null;
}

/**
 * Displays a helper alert showing the full destination path on the backup server.
 * Used in both create and edit backup forms to provide visual feedback to users.
 */
export function BackupDestinationHelper({
  fullPath,
}: BackupDestinationHelperProps) {
  if (!fullPath) return null;

  return (
    <Alert className="mt-2">
      <InfoIcon className="h-4 w-4" />
      <AlertDescription className="pl-7">
        Full path on backup server:{" "}
        <code className="text-xs bg-muted px-1 py-0.5 rounded">{fullPath}</code>
      </AlertDescription>
    </Alert>
  );
}
