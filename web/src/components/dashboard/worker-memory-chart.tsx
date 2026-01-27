"use client";

import { useEffect, useState, useCallback } from "react";
import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/shared/ui/card";
import {
  ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
  ChartLegend,
  ChartLegendContent,
} from "@/shared/ui/chart";
import {
  getWorkersStats,
  WorkerStatsWindow,
} from "@/services/worker-stats-service";
import { useWebSocket } from "@/shared/hooks/use-websocket";

export function WorkerMemoryChart() {
  const [stats, setStats] = useState<WorkerStatsWindow[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeWorkers, setActiveWorkers] = useState<Set<string>>(new Set());

  const fetchStats = useCallback(() => {
    getWorkersStats()
      .then((data) => {
        setStats(data);
        // Initialize active workers if not already set
        if (activeWorkers.size === 0 && data.length > 0) {
          const initial = new Set<string>();
          data.forEach((_, i) => initial.add(`worker_${i}`));
          setActiveWorkers(initial);
        }
      })
      .catch((err) => console.error("Failed to fetch worker stats:", err))
      .finally(() => setLoading(false));
  }, [activeWorkers.size]);

  useEffect(() => {
    fetchStats();
  }, [fetchStats]);

  useWebSocket((msg) => {
    if (msg.type === "worker_stats_updated") {
      fetchStats();
    }
  });

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>System Memory Usage</CardTitle>
          <CardDescription>Loading stats...</CardDescription>
        </CardHeader>
      </Card>
    );
  }

  const chartConfig: ChartConfig = {};
  const colors = [
    "#00FFFF", // Cyan
    "#FF00FF", // Magenta
    "#00FF00", // Green
    "#FFFF00", // Yellow
    "#FF4500", // OrangeRed
  ];

  // Create a mapping from worker_id to a stable key like worker_0, worker_1
  // Sort workers by ID to ensure stable indexing regardless of API response order
  const sortedStats = [...stats].sort((a, b) =>
    a.worker_id.localeCompare(b.worker_id),
  );
  const workerKeyMap = new Map<string, string>();
  sortedStats.forEach((worker, index) => {
    const key = `worker_${index}`;
    workerKeyMap.set(worker.worker_id, key);
    const isBackend = worker.worker_id === "backend";
    chartConfig[key] = {
      label: isBackend
        ? "Backend"
        : `Worker ${worker.worker_id.substring(0, 8)}`,
      color: isBackend ? "#FFD700" : colors[index % colors.length], // Gold for backend, or normal colors
    };
  });

  const toggleWorker = (key: string) => {
    const next = new Set(activeWorkers);
    if (next.has(key)) {
      if (next.size > 1) {
        // Keep at least one visible
        next.delete(key);
      }
    } else {
      next.add(key);
    }
    setActiveWorkers(next);
  };

  // Transform data for AreaChart
  // Group reports by 30-second intervals to avoid gaps between workers reporting at slightly different times
  const BUCKET_MS = 30 * 1000;

  // Map to store merged points: timestamp_bucket -> { workerKey -> value }
  const timeMap = new Map<number, Record<string, number>>();

  sortedStats.forEach((worker) => {
    const workerKey = workerKeyMap.get(worker.worker_id)!;
    worker.reports.forEach((report) => {
      const date = new Date(report.timestamp);
      const bucket = Math.floor(date.getTime() / BUCKET_MS) * BUCKET_MS;

      if (!timeMap.has(bucket)) {
        timeMap.set(bucket, {});
      }

      const point = timeMap.get(bucket)!;
      point[workerKey] = parseFloat(report.memory_percent.toFixed(2));
    });
  });

  // Convert map to sorted array for Recharts
  const sortedBuckets = Array.from(timeMap.keys()).sort((a, b) => a - b);

  // First pass: Find the first available value for each worker to perform "backfilling"
  // This ensures all lines start at the left edge of the chart
  const initialValues: Record<string, number> = {};
  Array.from(workerKeyMap.values()).forEach((key) => {
    for (const bucket of sortedBuckets) {
      const val = timeMap.get(bucket)?.[key];
      if (val !== undefined) {
        initialValues[key] = val;
        break;
      }
    }
  });

  const lastValues: Record<string, number> = { ...initialValues };

  const chartData = sortedBuckets.map((bucket) => {
    const bucketValues = timeMap.get(bucket)!;

    // Fill in missing values using the last known value (forward-filling)
    // or the initial value (back-filling)
    Array.from(workerKeyMap.values()).forEach((key) => {
      if (bucketValues[key] !== undefined) {
        lastValues[key] = bucketValues[key];
      } else if (lastValues[key] !== undefined) {
        bucketValues[key] = lastValues[key];
      }
    });

    return {
      timestamp: new Date(bucket).toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
      }),
      rawTime: bucket,
      ...bucketValues,
    };
  });

  return (
    <Card className="flex flex-col">
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <div className="space-y-1">
          <CardTitle>System Memory Usage</CardTitle>
          <CardDescription>
            Real-time memory allocation (Backend & Workers)
          </CardDescription>
        </div>
      </CardHeader>
      <CardContent className="flex-1 pb-0">
        <ChartContainer
          config={chartConfig}
          className="mx-auto aspect-[4/3] max-h-[350px] w-full"
        >
          <AreaChart
            data={chartData}
            margin={{
              left: 10,
              right: 12,
              top: 20,
              bottom: 0,
            }}
          >
            <defs>
              {Array.from(workerKeyMap.entries()).map(([_, workerKey]) => {
                const isBackend = workerKeyMap.get("backend") === workerKey;
                const baseColor = isBackend
                  ? "#FFD700"
                  : colors[parseInt(workerKey.split("_")[1]) % colors.length];
                return (
                  <linearGradient
                    key={`fill-${workerKey}`}
                    id={`fill-${workerKey}`}
                    x1="0"
                    y1="0"
                    x2="0"
                    y2="1"
                  >
                    <stop offset="5%" stopColor={baseColor} stopOpacity={0.8} />
                    <stop
                      offset="95%"
                      stopColor={baseColor}
                      stopOpacity={0.2}
                    />
                  </linearGradient>
                );
              })}
            </defs>
            <CartesianGrid
              vertical={false}
              strokeDasharray="3 3"
              className="stroke-muted/30"
            />
            <XAxis
              dataKey="timestamp"
              tickLine={false}
              axisLine={false}
              tickMargin={12}
              minTickGap={32}
              fontSize={12}
            />
            <YAxis
              tickLine={false}
              axisLine={false}
              tickMargin={12}
              tickFormatter={(value) => `${value}%`}
              fontSize={12}
              domain={[0, "auto"]}
            />
            <ChartTooltip
              cursor={{
                stroke: "rgba(var(--foreground), 0.1)",
                strokeWidth: 2,
              }}
              content={<ChartTooltipContent indicator="dot" />}
            />
            {Array.from(workerKeyMap.entries()).map(([workerId, workerKey]) => {
              const isBackend = workerId === "backend";
              const seriesColor = isBackend
                ? "#FFD700"
                : colors[parseInt(workerKey.split("_")[1]) % colors.length];
              return (
                <Area
                  key={workerId}
                  dataKey={workerKey}
                  type="monotone"
                  fill={`url(#fill-${workerKey})`}
                  stroke={seriesColor}
                  strokeWidth={1.5}
                  connectNulls={true}
                  hide={!activeWorkers.has(workerKey)}
                  activeDot={{
                    r: 8,
                    strokeWidth: 2,
                    stroke: "var(--background)",
                    fill: seriesColor,
                  }}
                  animationDuration={1000}
                  isAnimationActive={true}
                />
              );
            })}
          </AreaChart>
        </ChartContainer>
        {/* Custom Interactive Legend */}
        <div className="flex flex-wrap items-center justify-center gap-4 py-4 border-t border-muted/20">
          {Array.from(workerKeyMap.entries()).map(([workerId, workerKey]) => {
            const isActive = activeWorkers.has(workerKey);
            const isBackend = workerId === "backend";
            const color = isBackend
              ? "#FFD700"
              : colors[parseInt(workerKey.split("_")[1]) % colors.length];
            return (
              <button
                key={workerKey}
                onClick={() => toggleWorker(workerKey)}
                className={`flex items-center gap-2 px-3 py-1.5 rounded-full border transition-all duration-200 ${
                  isActive
                    ? "bg-muted/10 border-muted-foreground/30 opacity-100"
                    : "bg-transparent border-transparent opacity-40 hover:opacity-100"
                }`}
              >
                <div
                  className="w-3 h-3 rounded-full shadow-[0_0_8px_rgba(0,0,0,0.5)]"
                  style={{
                    backgroundColor: color,
                    boxShadow: isActive ? `0 0 10px ${color}` : "none",
                  }}
                />
                <span
                  className={`text-xs font-medium ${isActive ? "text-foreground" : "text-muted-foreground"}`}
                >
                  {isBackend ? "Backend" : `Worker ${workerId.substring(0, 8)}`}
                </span>
              </button>
            );
          })}
        </div>
      </CardContent>
      <div className="px-6 pb-6 text-[10px] text-muted-foreground italic text-center">
        * Data aggregated in 30s windows. Click legend to toggle workers.
      </div>
    </Card>
  );
}
