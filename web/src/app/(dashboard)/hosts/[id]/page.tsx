import { HostBackupsClient } from "./host-backups-client";

interface PageProps {
  params: Promise<{ id: string }>;
}

export default async function HostBackupsPage({ params }: PageProps) {
  const { id } = await params;
  const backupRoot = process.env.BACKUP_ROOT;

  return <HostBackupsClient hostId={id} backupRoot={backupRoot} />;
}
