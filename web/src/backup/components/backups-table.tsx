"use client";

import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  SortingState,
  VisibilityState,
  useReactTable,
} from "@tanstack/react-table";
import cronstrue from "cronstrue";
import { BackupResponse, runBackup } from "@/backup/infrastructure/backup-api";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/shared/ui/table";
import { Card, CardContent } from "@/shared/ui/card";
import { Button } from "@/shared/ui/button";
import { useState } from "react";
import { EditBackupDialog } from "./edit-backup-dialog";
import { DeleteBackupDialog } from "./delete-backup-dialog";
import { BackupDetailsSheet } from "./backup-details-sheet";
import { RowData } from "@tanstack/react-table";
import { ArrowUpDown, ChevronDown, Play } from "lucide-react";
import { useWebSocket } from "@/shared/hooks/use-websocket";
import { Spinner } from "@/shared/ui/spinner";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/shared/ui/dropdown-menu";
import { formatSize } from "@/shared/lib/utils";
import { Input } from "@/shared/ui/input";

declare module "@tanstack/react-table" {
  interface TableMeta<TData extends RowData> {
    onEdit: (backup: TData) => void;
    onDelete: (backup: TData) => void;
  }
}

const columnHelper = createColumnHelper<BackupResponse>();

interface BackupsTableProps {
  backups: BackupResponse[];
  backupRoot?: string;
  hostPath?: string;
  onRefresh?: () => void;
}

export function BackupsTable({
  backups,
  backupRoot,
  hostPath,
  onRefresh,
}: BackupsTableProps) {
  const [selectedBackup, setSelectedBackup] = useState<BackupResponse | null>(
    null,
  );
  const [backupToDelete, setBackupToDelete] = useState<BackupResponse | null>(
    null,
  );
  const [selectedBackupForDetails, setSelectedBackupForDetails] =
    useState<BackupResponse | null>(null);
  const [runningBackups, setRunningBackups] = useState<Set<string>>(new Set());
  const [sorting, setSorting] = useState<SortingState>([]);
  const [globalFilter, setGlobalFilter] = useState<string>("");
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>(
    () => {
      if (typeof window !== "undefined") {
        const saved = localStorage.getItem("backupsTableColumnVisibility");
        if (saved) {
          try {
            return JSON.parse(saved);
          } catch {
            // Invalid JSON, use defaults
          }
        }
      }
      return { destination: false, schedule: false };
    },
  );

  // Persist column visibility to localStorage
  const handleColumnVisibilityChange = (
    updaterOrValue:
      | VisibilityState
      | ((old: VisibilityState) => VisibilityState),
  ) => {
    const newValue =
      typeof updaterOrValue === "function"
        ? updaterOrValue(columnVisibility)
        : updaterOrValue;
    setColumnVisibility(newValue);
    if (typeof window !== "undefined") {
      localStorage.setItem(
        "backupsTableColumnVisibility",
        JSON.stringify(newValue),
      );
    }
  };

  const handleWebSocketMessage = (data: any) => {
    if (data.type === "backup_completed" || data.type === "backup_failed") {
      setRunningBackups((prev) => {
        const newSet = new Set(prev);
        newSet.delete(data.backup_id);
        return newSet;
      });
      onRefresh?.();
    }
  };

  useWebSocket(handleWebSocketMessage);

  const columns = [
    columnHelper.accessor("hostName", {
      enableSorting: true,
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        >
          Host
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: (info) => <span className="font-medium">{info.getValue()}</span>,
    }),
    columnHelper.accessor("hostAddress", {
      enableSorting: true,
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        >
          Hostname / IP
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: (info) => info.getValue(),
    }),
    columnHelper.accessor("path", {
      enableSorting: true,
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        >
          Path
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: (info) => <span className="font-medium">{info.getValue()}</span>,
    }),
    columnHelper.accessor("destination", {
      enableSorting: true,
      header: "Destination",
      cell: (info) => {
        const destination = info.getValue();
        if (!backupRoot) return destination;

        // Combine paths and handle slashes
        const cleanRoot = backupRoot.replace(/\/+$/, "");
        const cleanHostPath = hostPath
          ? hostPath.replace(/(^\/+|\/+$)/g, "")
          : "";
        const cleanDest = destination.replace(/^\/+/, "");

        const parts = [cleanRoot];
        if (cleanHostPath) parts.push(cleanHostPath);
        parts.push(cleanDest);

        return parts.join("/");
      },
    }),
    columnHelper.accessor("schedule", {
      enableSorting: true,
      enableGlobalFilter: false,
      header: "Schedule",
      cell: (info) => {
        try {
          return cronstrue.toString(info.getValue());
        } catch (e) {
          return info.getValue();
        }
      },
    }),
    columnHelper.accessor("size", {
      enableSorting: true,
      enableGlobalFilter: false,
      sortingFn: (rowA, rowB, columnId) => {
        const sizeA = rowA.original.size || 0;
        const sizeB = rowB.original.size || 0;
        return sizeA - sizeB;
      },
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        >
          Size
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: (info) => {
        const size = info.getValue() || 0;
        return <span className="font-medium">{formatSize(size)}</span>;
      },
    }),
    columnHelper.accessor("lastRun", {
      enableSorting: true,
      enableGlobalFilter: false,
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        >
          Last Backup
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: (info) => {
        const date = new Date(info.getValue());
        return date.getFullYear() === 1 ? "Never" : date.toLocaleString();
      },
    }),
    columnHelper.accessor("status", {
      enableSorting: true,
      enableGlobalFilter: false,
      header: ({ column }) => (
        <Button
          variant="ghost"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        >
          Status
          <ArrowUpDown className="ml-2 h-4 w-4" />
        </Button>
      ),
      cell: (info) => (
        <span
          className={`px-2 py-1 rounded-full text-xs font-medium ${
            info.getValue() === "completed"
              ? "bg-green-500/10 text-green-600 dark:text-green-500"
              : info.getValue() === "failed"
                ? "bg-destructive/10 text-destructive"
                : "bg-yellow-500/10 text-yellow-600 dark:text-yellow-500"
          }`}
        >
          {info.getValue()}
        </span>
      ),
    }),
    columnHelper.accessor("encrypted", {
      enableSorting: true,
      header: "Encrypted",
      cell: (info) => (
        <span
          className={`px-2 py-1 rounded-full text-xs font-medium ${
            info.getValue()
              ? "bg-blue-500/10 text-blue-600 dark:text-blue-400"
              : "bg-muted text-muted-foreground"
          }`}
        >
          {info.getValue() ? "Yes" : "No"}
        </span>
      ),
    }),
    columnHelper.display({
      id: "actions",
      cell: (info) => (
        <div className="flex gap-2">
          <Button
            variant="ghost"
            size="sm"
            disabled={runningBackups.has(info.row.original.id)}
            onClick={async () => {
              const id = info.row.original.id;
              setRunningBackups((prev) => new Set(prev).add(id));
              try {
                await runBackup(id);
                // onRefresh?.(); // Removed immediate refresh, waiting for WS
              } catch (error) {
                console.error("Failed to start backup", error);
                setRunningBackups((prev) => {
                  const newSet = new Set(prev);
                  newSet.delete(id);
                  return newSet;
                });
              }
            }}
          >
            {runningBackups.has(info.row.original.id) ? (
              <Spinner className="h-4 w-4 mr-1" />
            ) : (
              <Play className="h-4 w-4 mr-1" />
            )}
            Run
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => info.table.options.meta?.onEdit(info.row.original)}
          >
            Edit
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className="text-destructive hover:text-destructive/90 hover:bg-destructive/10"
            onClick={() => info.table.options.meta?.onDelete(info.row.original)}
          >
            Delete
          </Button>
        </div>
      ),
    }),
  ];

  const table = useReactTable({
    data: backups,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onSortingChange: setSorting,
    onGlobalFilterChange: setGlobalFilter,
    onColumnVisibilityChange: handleColumnVisibilityChange,
    state: {
      sorting,
      columnVisibility,
      globalFilter,
    },
    meta: {
      onEdit: (backup: BackupResponse) => setSelectedBackup(backup),
      onDelete: (backup: BackupResponse) => setBackupToDelete(backup),
    },
  });

  if (!backups || backups.length === 0) {
    return (
      <Card>
        <CardContent className="p-6 text-center text-muted-foreground">
          No backups found for this host.
        </CardContent>
      </Card>
    );
  }

  return (
    <>
      <div className="flex items-center gap-4 py-4">
        <Input
          placeholder="Filter backups..."
          value={globalFilter ?? ""}
          onChange={(event) => setGlobalFilter(event.target.value)}
          className="max-w-sm"
        />
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" className="ml-auto">
              Columns <ChevronDown className="ml-2 h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {table
              .getAllColumns()
              .filter((column) => column.getCanHide())
              .map((column) => {
                return (
                  <DropdownMenuCheckboxItem
                    key={column.id}
                    className="capitalize"
                    checked={column.getIsVisible()}
                    onCheckedChange={(value) =>
                      column.toggleVisibility(!!value)
                    }
                  >
                    {column.id}
                  </DropdownMenuCheckboxItem>
                );
              })}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext(),
                        )}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows.map((row) => (
              <TableRow
                key={row.id}
                data-state={row.getIsSelected() && "selected"}
                className="hover:bg-muted/50 cursor-pointer"
                onClick={() => setSelectedBackupForDetails(row.original)}
              >
                {row.getVisibleCells().map((cell) => (
                  <TableCell
                    key={cell.id}
                    onClick={(e) => {
                      // Prevent row click when clicking on action buttons
                      if (cell.column.id === "actions") {
                        e.stopPropagation();
                      }
                    }}
                  >
                    {flexRender(cell.column.columnDef.cell, cell.getContext())}
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      <EditBackupDialog
        backup={selectedBackup}
        open={!!selectedBackup}
        onOpenChange={(open) => !open && setSelectedBackup(null)}
        onSuccess={onRefresh}
        backupRoot={backupRoot}
        hostPath={hostPath}
      />
      <DeleteBackupDialog
        backup={backupToDelete}
        open={!!backupToDelete}
        onOpenChange={(open) => !open && setBackupToDelete(null)}
        onSuccess={onRefresh}
      />
      <BackupDetailsSheet
        backup={selectedBackupForDetails}
        open={!!selectedBackupForDetails}
        onOpenChange={(open) => !open && setSelectedBackupForDetails(null)}
        backupRoot={backupRoot}
        hostPath={hostPath}
      />
    </>
  );
}
