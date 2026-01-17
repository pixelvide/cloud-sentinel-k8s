"use client";

import { useMemo } from "react";
import {
    LineChart,
    Line,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    ResponsiveContainer,
    Area,
    AreaChart,
} from "recharts";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

interface UsageDataPoint {
    timestamp: string;
    value: number;
}

interface MetricsChartProps {
    title: string;
    data: UsageDataPoint[];
    unit: string;
    color?: string;
    className?: string;
    fallback?: boolean;
}

export function MetricsChart({
    title,
    data,
    unit,
    color = "#3b82f6", // blue-500
    className,
    fallback,
}: MetricsChartProps) {
    const formattedData = useMemo(() => {
        if (!data) return [];
        return data.map((d) => ({
            ...d,
            formattedTime: new Date(d.timestamp).toLocaleTimeString([], {
                hour: "2-digit",
                minute: "2-digit",
            }),
        }));
    }, [data]);

    if (!data || data.length === 0) {
        return (
            <Card className={cn("flex flex-col", className)}>
                <CardHeader className="pb-2">
                    <CardTitle className="text-sm font-medium text-muted-foreground">
                        {title}
                    </CardTitle>
                </CardHeader>
                <CardContent className="flex-1 flex items-center justify-center min-h-[200px] text-muted-foreground text-sm">
                    No data available
                </CardContent>
            </Card>
        );
    }

    const latestValue = data[data.length - 1].value;

    return (
        <Card className={cn("flex flex-col", className)}>
            <CardHeader className="pb-2 flex flex-row items-center justify-between space-y-0">
                <CardTitle className="text-sm font-medium text-muted-foreground">
                    {title}
                </CardTitle>
                <div className="flex items-center gap-2">
                    {fallback && (
                        <Badge variant="outline" className="text-[10px] h-5 border-amber-500/50 text-amber-500 bg-amber-500/10">
                            Cached
                        </Badge>
                    )}
                    <span className="text-2xl font-bold">
                        {latestValue.toFixed(3)}
                        <span className="text-xs font-normal text-muted-foreground ml-1">
                            {unit}
                        </span>
                    </span>
                </div>
            </CardHeader>
            <CardContent className="flex-1 min-h-[200px] p-0 pb-4 pr-4">
                <ResponsiveContainer width="100%" height="100%">
                    <AreaChart data={formattedData}>
                        <defs>
                            <linearGradient id={`gradient-${title}`} x1="0" y1="0" x2="0" y2="1">
                                <stop offset="5%" stopColor={color} stopOpacity={0.3} />
                                <stop offset="95%" stopColor={color} stopOpacity={0} />
                            </linearGradient>
                        </defs>
                        <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="var(--border)" opacity={0.4} />
                        <XAxis
                            dataKey="formattedTime"
                            tick={{ fontSize: 10, fill: "var(--muted-foreground)" }}
                            tickLine={false}
                            axisLine={false}
                            minTickGap={30}
                        />
                        <YAxis
                            tick={{ fontSize: 10, fill: "var(--muted-foreground)" }}
                            tickLine={false}
                            axisLine={false}
                            tickFormatter={(value) => value.toFixed(2)}
                            width={40}
                        />
                        <Tooltip
                            contentStyle={{
                                backgroundColor: "var(--popover)",
                                borderColor: "var(--border)",
                                borderRadius: "var(--radius)",
                                fontSize: "12px",
                            }}
                            itemStyle={{ color: "var(--popover-foreground)" }}
                            labelStyle={{ color: "var(--muted-foreground)", marginBottom: "4px" }}
                        />
                        <Area
                            type="monotone"
                            dataKey="value"
                            stroke={color}
                            strokeWidth={2}
                            fillOpacity={1}
                            fill={`url(#gradient-${title})`}
                            isAnimationActive={false}
                        />
                    </AreaChart>
                </ResponsiveContainer>
            </CardContent>
        </Card>
    );
}
