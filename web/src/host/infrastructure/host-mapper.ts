import { HostDTO } from "./host-dtos";
import { Host } from "../domain/models";

export const toDomain = (dto: HostDTO): Host => {
  return {
    id: dto.id,
    name: dto.name,
    hostname: dto.hostname,
    user: dto.user,
    port: dto.port,
    path: dto.path,
    isWorkstation: dto.is_workstation,
    failedBackupsCount: dto.failed_backups_count,
  };
};
