import { BackupDTO } from "./backup-dtos";
import { Backup, Hook } from "../domain/models";

export const toDomain = (dto: BackupDTO): Backup => {
  return {
    id: dto.id,
    hostId: dto.host_id,
    hostName: dto.host_name,
    hostAddress: dto.host_address,
    path: dto.path,
    destination: dto.destination,
    status: dto.status,
    schedule: dto.schedule,
    lastRun: dto.last_run,
    excludes: dto.excludes || [],
    incremental: dto.incremental,
    size: dto.size ? parseInt(dto.size) : 0,
    retention: dto.retention,
    encrypted: dto.encrypted,
    hooks: (dto.hooks || []).map(
      (h): Hook => ({
        id: h.id,
        name: h.name,
        phase: h.phase,
        enabled: h.enabled,
        params: h.params,
      }),
    ),
  };
};
