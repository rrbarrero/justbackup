"use client";

import { useEffect, useState, useCallback } from "react";
import Link from "next/link";
import { useWebSocket } from "@/shared/hooks/use-websocket";
import { Card, CardContent, CardHeader, CardTitle } from "@/shared/ui/card";
import {
  getDashboardStats,
  DashboardStats,
} from "@/services/dashboard-service";
import { DiskUsageChart } from "@/components/dashboard/disk-usage-chart";
import { WorkerMemoryChart } from "@/components/dashboard/worker-memory-chart";
import {
  Server,
  Database,
  HardDrive,
  AlertCircle,
  CheckCircle,
  Clock,
  Activity,
  Cpu,
} from "lucide-react";

export default function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchStats = useCallback(() => {
    getDashboardStats()
      .then(setStats)
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    fetchStats();
  }, [fetchStats]);

  useWebSocket((msg) => {
    if (
      msg.type === "worker_stats_updated" ||
      msg.type === "backup_completed" ||
      msg.type === "backup_failed"
    ) {
      fetchStats();
    }
  });

  if (loading) {
    return (
      <div className="p-8 text-center text-muted-foreground">
        Loading dashboard stats...
      </div>
    );
  }

  if (error) {
    return <div className="p-8 text-center text-red-500">Error: {error}</div>;
  }

  if (!stats) {
    return null;
  }

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Hosts</CardTitle>
            <Server className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.total_hosts}</div>
            <p className="text-xs text-muted-foreground">Managed servers</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Backups</CardTitle>
            <Database className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.total_backups}</div>
            <p className="text-xs text-muted-foreground">Configured tasks</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Active Workers
            </CardTitle>
            <Cpu className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.active_workers}</div>
            <p className="text-xs text-muted-foreground">Connected instances</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Data Volume</CardTitle>
            <HardDrive className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {stats.backup_stats.total_size || "0 B"}
            </div>
            <p className="text-xs text-muted-foreground">Total backed up</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Success Rate</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {stats.backup_stats.total > 0
                ? Math.round(
                    (stats.backup_stats.completed / stats.backup_stats.total) *
                      100,
                  )
                : 0}
              %
            </div>
            <p className="text-xs text-muted-foreground">Completed tasks</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Failed Backups
            </CardTitle>
            <AlertCircle className="h-4 w-4 text-red-500" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-500">
              {stats.backup_stats.failed}
            </div>
            <p className="text-xs text-muted-foreground">Requires attention</p>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        <div className="col-span-1">
          <DiskUsageChart />
        </div>
        <Card className="col-span-1">
          <CardHeader>
            <CardTitle>Backup Status Breakdown</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <Link
                href="/backups?status=completed"
                className="flex items-center hover:bg-muted/50 p-2 rounded-md transition-colors cursor-pointer"
              >
                <CheckCircle className="mr-2 h-4 w-4 text-green-500" />
                <div className="flex-1 space-y-1">
                  <p className="text-sm font-medium leading-none">Completed</p>
                  <p className="text-sm text-muted-foreground">
                    Successfully finished backups
                  </p>
                </div>
                <div className="font-medium">
                  {stats.backup_stats.completed}
                </div>
              </Link>
              <Link
                href="/backups?status=pending"
                className="flex items-center hover:bg-muted/50 p-2 rounded-md transition-colors cursor-pointer"
              >
                <Clock className="mr-2 h-4 w-4 text-yellow-500" />
                <div className="flex-1 space-y-1">
                  <p className="text-sm font-medium leading-none">Pending</p>
                  <p className="text-sm text-muted-foreground">
                    Scheduled or running
                  </p>
                </div>
                <div className="font-medium">{stats.backup_stats.pending}</div>
              </Link>
              <Link
                href="/backups?status=failed"
                className="flex items-center hover:bg-muted/50 p-2 rounded-md transition-colors cursor-pointer"
              >
                <AlertCircle className="mr-2 h-4 w-4 text-red-500" />
                <div className="flex-1 space-y-1">
                  <p className="text-sm font-medium leading-none">Failed</p>
                  <p className="text-sm text-muted-foreground">
                    Encountered errors
                  </p>
                </div>
                <div className="font-medium">{stats.backup_stats.failed}</div>
              </Link>
            </div>
          </CardContent>
        </Card>
        <div className="col-span-1 lg:col-span-1 xl:col-span-2">
          <WorkerMemoryChart />
        </div>
      </div>
    </div>
  );
}
