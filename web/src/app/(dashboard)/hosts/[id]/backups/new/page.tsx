import { CreateBackupFormClient } from "./create-backup-form-client";

interface PageProps {
  params: Promise<{ id: string }>;
}

export default async function NewBackupPage({ params }: PageProps) {
  const { id } = await params;
  const backupRoot = process.env.BACKUP_ROOT;

  return <CreateBackupFormClient hostId={id} backupRoot={backupRoot} />;
}
