"use client";

import { useEffect, useState } from "react";
import { Pie, PieChart, Label } from "recharts";
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
import { getDiskUsage, DiskUsage } from "@/services/system-service";

const chartConfig = {
  free: {
    label: "Free Space",
    color: "#10b981", // Emerald green for available space
  },
  used: {
    label: "Used Space",
    color: "#f59e0b", // Amber/orange for used space
  },
} satisfies ChartConfig;

export function DiskUsageChart() {
  const [data, setData] = useState<DiskUsage | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getDiskUsage()
      .then(setData)
      .catch((err) => console.error(err))
      .finally(() => setLoading(false));
  }, []);

  if (loading)
    return (
      <Card className="flex flex-col">
        <CardHeader className="items-center pb-0">
          <CardTitle>Disk Usage</CardTitle>
          <CardDescription>Loading...</CardDescription>
        </CardHeader>
      </Card>
    );

  if (!data) return null;

  // Data from backend is in bytes (statfs blocks * bsize)
  // Converting to GB for display, but Keeping values for chart
  const chartData = [
    { name: "free", value: parseInt(data.free), fill: chartConfig.free.color },
    { name: "used", value: parseInt(data.used), fill: chartConfig.used.color },
  ];

  const totalGB = (parseInt(data.total) / 1024 / 1024 / 1024).toFixed(2);
  const freeGB = (parseInt(data.free) / 1024 / 1024 / 1024).toFixed(2);
  const usedPercentage = Math.round(
    (parseInt(data.used) / parseInt(data.total)) * 100,
  );

  return (
    <Card className="flex flex-col">
      <CardHeader className="items-center pb-0">
        <CardTitle>Disk Usage</CardTitle>
        <CardDescription>/mnt/backups</CardDescription>
      </CardHeader>
      <CardContent className="flex-1 pb-0">
        <ChartContainer
          config={chartConfig}
          className="mx-auto aspect-square max-h-[250px]"
        >
          <PieChart>
            <ChartTooltip
              cursor={false}
              content={<ChartTooltipContent hideLabel />}
            />
            <Pie
              data={chartData}
              dataKey="value"
              nameKey="name"
              innerRadius={60}
              strokeWidth={5}
            >
              <Label
                content={({ viewBox }) => {
                  if (viewBox && "cx" in viewBox && "cy" in viewBox) {
                    return (
                      <text
                        x={viewBox.cx}
                        y={viewBox.cy}
                        textAnchor="middle"
                        dominantBaseline="middle"
                      >
                        <tspan
                          x={viewBox.cx}
                          y={viewBox.cy}
                          className="fill-foreground text-3xl font-bold"
                        >
                          {usedPercentage}%
                        </tspan>
                        <tspan
                          x={viewBox.cx}
                          y={(viewBox.cy || 0) + 24}
                          className="fill-muted-foreground"
                        >
                          Used
                        </tspan>
                      </text>
                    );
                  }
                }}
              />
            </Pie>
            <ChartLegend
              content={<ChartLegendContent nameKey="name" />}
              className="-translate-y-2 flex-wrap gap-2 [&>*]:basis-1/4 [&>*]:justify-center"
            />
          </PieChart>
        </ChartContainer>
      </CardContent>
      <div className="flex-col gap-2 text-sm text-center pb-4 text-muted-foreground">
        <div>Total: {totalGB} GB</div>
        <div>Free: {freeGB} GB</div>
      </div>
    </Card>
  );
}
