"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
  VisibilityState,
} from "@tanstack/react-table";
import {
  getHosts,
  HostResponse,
  runHostBackups,
} from "@/host/infrastructure/host-api";
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
import { EditHostDialog } from "./edit-host-dialog";
import { DeleteHostDialog } from "./delete-host-dialog";
import { ArrowUpDown, ChevronDown, Play } from "lucide-react";
import { Spinner } from "@/shared/ui/spinner";
import { useWebSocket } from "@/shared/hooks/use-websocket";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuTrigger,
} from "@/shared/ui/dropdown-menu";
import { Input } from "@/shared/ui/input";

const columnHelper = createColumnHelper<HostResponse>();

const columns = [
  columnHelper.accessor("name", {
    header: ({ column }) => (
      <Button
        variant="ghost"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
      >
        Name
        <ArrowUpDown className="ml-2 h-4 w-4" />
      </Button>
    ),
    cell: (info) => <span className="font-medium">{info.getValue()}</span>,
  }),
  columnHelper.accessor("hostname", {
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
  columnHelper.accessor("user", {
    header: ({ column }) => (
      <Button
        variant="ghost"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
      >
        User
        <ArrowUpDown className="ml-2 h-4 w-4" />
      </Button>
    ),
    cell: (info) => info.getValue(),
  }),
  columnHelper.accessor("port", {
    header: ({ column }) => (
      <Button
        variant="ghost"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
      >
        Port
        <ArrowUpDown className="ml-2 h-4 w-4" />
      </Button>
    ),
    cell: (info) => info.getValue(),
  }),
  columnHelper.accessor("failedBackupsCount", {
    header: ({ column }) => (
      <Button
        variant="ghost"
        onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
      >
        Failed Backups
        <ArrowUpDown className="ml-2 h-4 w-4" />
      </Button>
    ),
    cell: (info) => {
      const count = info.getValue() ?? 0;
      if (count > 0) {
        return <span className="text-red-500 font-medium">{count}</span>;
      }
      return <span className="text-muted-foreground">0</span>;
    },
  }),
];

export function HostsTable() {
  const router = useRouter();
  const [hosts, setHosts] = useState<HostResponse[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [editingHost, setEditingHost] = useState<HostResponse | null>(null);
  const [hostToDelete, setHostToDelete] = useState<HostResponse | null>(null);
  const [runningBackups, setRunningBackups] = useState<Set<string>>(new Set());
  const [hostToBackups, setHostToBackups] = useState<Map<string, string[]>>(
    new Map(),
  );
  const [sorting, setSorting] = useState<SortingState>([]);
  const [globalFilter, setGlobalFilter] = useState<string>("");
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>(
    () => {
      if (typeof window !== "undefined") {
        const saved = localStorage.getItem("hostsTableColumnVisibility");
        if (saved) {
          try {
            return JSON.parse(saved);
          } catch {
            // Invalid JSON, use defaults
          }
        }
      }
      return {};
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
        "hostsTableColumnVisibility",
        JSON.stringify(newValue),
      );
    }
  };

  const fetchHosts = async () => {
    try {
      const data = await getHosts();
      setHosts(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load hosts");
    } finally {
      setIsLoading(false);
    }
  };

  const handleWebSocketMessage = (data: any) => {
    if (data.type === "backup_completed" || data.type === "backup_failed") {
      setRunningBackups((prev) => {
        const newSet = new Set(prev);
        newSet.delete(data.backup_id);
        return newSet;
      });
      fetchHosts();
    }
  };

  useWebSocket(handleWebSocketMessage);

  useEffect(() => {
    fetchHosts();
  }, []);

  const handleEdit = (e: React.MouseEvent, host: HostResponse) => {
    e.stopPropagation();
    setEditingHost(host);
  };

  const handleRunAll = async (e: React.MouseEvent, host: HostResponse) => {
    e.stopPropagation();
    try {
      const { task_ids } = await runHostBackups(host.id);
      setRunningBackups((prev) => {
        const newSet = new Set(prev);
        task_ids.forEach((id) => newSet.add(id));
        return newSet;
      });
      setHostToBackups((prev) => {
        const newMap = new Map(prev);
        newMap.set(host.id, task_ids);
        return newMap;
      });
    } catch (err) {
      console.error("Failed to run host backups", err);
    }
  };

  const isHostRunning = (hostId: string) => {
    const backupIds = hostToBackups.get(hostId);
    if (!backupIds) return false;
    return backupIds.some((id) => runningBackups.has(id));
  };

  const table = useReactTable({
    data: hosts,
    columns: [
      ...columns,
      columnHelper.display({
        id: "actions",
        cell: (info) => (
          <div className="flex justify-end gap-2">
            <Button
              variant="ghost"
              size="sm"
              disabled={isHostRunning(info.row.original.id)}
              onClick={(e) => handleRunAll(e, info.row.original)}
            >
              {isHostRunning(info.row.original.id) ? (
                <Spinner className="h-4 w-4 mr-1" />
              ) : (
                <Play className="h-4 w-4 mr-1" />
              )}
              Run All
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={(e) => handleEdit(e, info.row.original)}
            >
              Edit
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className="text-red-600 hover:text-red-700 hover:bg-red-50"
              onClick={(e) => {
                e.stopPropagation();
                setHostToDelete(info.row.original);
              }}
            >
              Delete
            </Button>
          </div>
        ),
      }),
    ],
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
  });

  if (isLoading) {
    return <div className="text-center p-4">Loading hosts...</div>;
  }

  if (error) {
    return <div className="text-red-500 p-4">Error: {error}</div>;
  }

  if (!hosts || hosts.length === 0) {
    return (
      <Card>
        <CardContent className="p-6 text-center text-gray-500">
          No hosts found. Add your first host to get started.
        </CardContent>
      </Card>
    );
  }

  return (
    <>
      <div className="flex items-center gap-4 py-4">
        <Input
          placeholder="Filter hosts..."
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
                className="cursor-pointer hover:bg-muted/50"
                onClick={() => router.push(`/hosts/${row.original.id}`)}
              >
                {row.getVisibleCells().map((cell) => (
                  <TableCell
                    key={cell.id}
                    onClick={(e) => {
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

      <EditHostDialog
        host={editingHost}
        open={!!editingHost}
        onOpenChange={(open) => !open && setEditingHost(null)}
        onSuccess={() => {
          fetchHosts();
          setEditingHost(null);
        }}
      />
      <DeleteHostDialog
        host={hostToDelete}
        open={!!hostToDelete}
        onOpenChange={(open) => !open && setHostToDelete(null)}
        onSuccess={() => {
          fetchHosts();
          setHostToDelete(null);
        }}
      />
    </>
  );
}
