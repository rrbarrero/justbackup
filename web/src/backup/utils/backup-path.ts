/**
 * Computes the full backup destination path by combining backup root, host path, and destination.
 * Handles path cleaning (removes trailing/leading slashes) to avoid double slashes.
 *
 * @param backupRoot - The root backup directory (from BACKUP_ROOT env var)
 * @param hostPath - The host-specific path segment
 * @param destination - The user-specified destination path
 * @returns The complete backup path or null if backupRoot is not provided
 */
export function computeBackupFullPath(
  backupRoot: string | undefined,
  hostPath: string | undefined,
  destination: string,
): string | null {
  if (!backupRoot) return null;

  const cleanRoot = backupRoot.replace(/\/+$/, "");
  const cleanHostPath = hostPath ? hostPath.replace(/(^\/+|\/+$)/g, "") : "";
  const cleanDest = destination.replace(/^\/+/, "");

  const parts = [cleanRoot];
  if (cleanHostPath) parts.push(cleanHostPath);
  if (cleanDest) parts.push(cleanDest);

  return parts.join("/");
}
