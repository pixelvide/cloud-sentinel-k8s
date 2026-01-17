"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Clock, RefreshCw, PauseCircle, PlayCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn, formatAge } from "@/lib/utils";
import { NamespaceBadge } from "@/components/NamespaceBadge";
import { ResourceDetailsSheet } from "@/components/ResourceDetailsSheet";
import { api } from "@/lib/api";
import cronParser from "cron-parser";

interface CronJobInfo {
    name: string;
    namespace: string;
    schedule: string;
    timezone?: string;
    suspend: boolean;
    active: number;
    last_schedule: string;
    age: string;
}

function CronJobsContent() {
    const [searchParams] = useSearchParams();
    const selectedContext = searchParams.get("context") || "";
    const selectedNamespace = searchParams.get("namespace") || "";

    const [cronjobs, setCronjobs] = useState<CronJobInfo[]>([]);
    const [loading, setLoading] = useState(false);
    const searchQuery = searchParams.get("q") || "";
    const [selectedCronJob, setSelectedCronJob] = useState<CronJobInfo | null>(null);

    const filteredCronJobs = cronjobs.filter((cj) => cj.name.toLowerCase().includes(searchQuery.toLowerCase()));

    useEffect(() => {
        if (!selectedContext || !selectedNamespace) {
            setCronjobs([]);
            return;
        }
        fetchCronJobs();
    }, [selectedContext, selectedNamespace]);

    const fetchCronJobs = async () => {
        setLoading(true);
        setCronjobs([]);
        try {
            const data = await api.get<any>(`/kube/cron-jobs?namespace=${selectedNamespace}`, {
                headers: { "x-kube-context": selectedContext || "" },
            });
            setCronjobs(data.cronjobs || []);
        } catch (error) {
            console.error("Failed to fetch cronjobs:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleRefresh = () => {
        if (selectedContext && selectedNamespace) {
            fetchCronJobs();
        }
    };

    return (
        <div className="flex flex-col items-center w-full min-h-screen bg-background/50 p-4 md:p-0">
            <div className="w-full max-w-6xl space-y-8">
                <Card className="border-none shadow-2xl shadow-black/5 bg-card/50 backdrop-blur-sm overflow-hidden rounded-3xl">
                    <CardHeader className="border-b bg-card/50 px-8 py-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <CardTitle className="text-lg font-bold flex items-center gap-3">
                                    <Clock className="h-5 w-5 text-teal-500" />
                                    {cronjobs.length} CronJobs Found
                                </CardTitle>
                                <CardDescription>
                                    {!selectedContext || !selectedNamespace
                                        ? "Select a namespace from the top bar to view cronjobs"
                                        : null}
                                </CardDescription>
                            </div>
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={handleRefresh}
                                disabled={!selectedContext || !selectedNamespace || loading}
                                className="rounded-xl"
                            >
                                <RefreshCw className={cn("h-4 w-4 mr-2", loading && "animate-spin")} />
                                Refresh
                            </Button>
                        </div>
                    </CardHeader>
                    <CardContent className="p-8">
                        {loading ? (
                            <div className="flex items-center justify-center py-12">
                                <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
                            </div>
                        ) : !selectedContext ? (
                            <div className="text-center py-12">
                                <Clock className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    Select a cluster from the top bar to view cronjobs.
                                </p>
                            </div>
                        ) : !selectedNamespace ? (
                            <div className="text-center py-12">
                                <Clock className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    Select a namespace from the top bar to view cronjobs.
                                </p>
                            </div>
                        ) : cronjobs.length === 0 ? (
                            <div className="text-center py-12">
                                <Clock className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">No cronjobs found in this namespace.</p>
                            </div>
                        ) : (
                            <div className="grid grid-cols-1 gap-4">
                                {filteredCronJobs.map((cj) => (
                                    <div
                                        key={cj.name + cj.namespace}
                                        className="p-6 bg-muted/30 rounded-2xl border border-muted/20 hover:bg-muted/50 transition-colors cursor-pointer"
                                        onClick={() => setSelectedCronJob(cj)}
                                    >
                                        <div className="flex flex-col lg:flex-row lg:items-center gap-4 justify-between">
                                            <div className="flex items-center gap-4 min-w-0">
                                                <div
                                                    className={cn(
                                                        "p-3 rounded-xl",
                                                        cj.suspend ? "bg-gray-500/10" : "bg-teal-500/10"
                                                    )}
                                                >
                                                    {cj.suspend ? (
                                                        <PauseCircle className="h-6 w-6 text-gray-500" />
                                                    ) : (
                                                        <PlayCircle className="h-6 w-6 text-teal-500" />
                                                    )}
                                                </div>
                                                <div className="min-w-0">
                                                    <p className="font-semibold text-base truncate">{cj.name}</p>
                                                    <div className="flex items-center gap-2 mt-1 flex-wrap">
                                                        <code className="text-[10px] font-mono bg-muted px-2 py-0.5 rounded">
                                                            {cj.schedule}
                                                        </code>
                                                        {cj.timezone && (
                                                            <span className="text-[10px] font-medium text-muted-foreground bg-muted/50 px-2 py-0.5 rounded">
                                                                {cj.timezone}
                                                            </span>
                                                        )}
                                                        {cj.suspend && (
                                                            <span className="text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full bg-gray-500/10 text-gray-600">
                                                                Suspended
                                                            </span>
                                                        )}
                                                        {cj.last_schedule && (
                                                            <span className="text-xs text-muted-foreground">
                                                                Last: {formatAge(cj.last_schedule)}
                                                            </span>
                                                        )}
                                                        <span className="text-xs text-muted-foreground">
                                                            Next:{" "}
                                                            {cj.suspend
                                                                ? "N/A"
                                                                : (() => {
                                                                      try {
                                                                          const parser =
                                                                              (cronParser as any).parse ||
                                                                              (cronParser as any).default?.parse;
                                                                          if (!parser) return "Parser Error";
                                                                          const tz = cj.timezone || "UTC";
                                                                          const nextDate = parser(cj.schedule, { tz })
                                                                              .next()
                                                                              .toDate();
                                                                          return nextDate.toLocaleString();
                                                                      } catch (e) {
                                                                          return "Invalid Schedule";
                                                                      }
                                                                  })()}
                                                        </span>
                                                    </div>
                                                </div>
                                            </div>

                                            {/* Namespace */}
                                            <div
                                                className="flex flex-col items-end min-w-[120px]"
                                                onClick={(e) => e.stopPropagation()}
                                            >
                                                <NamespaceBadge namespace={cj.namespace} />
                                            </div>

                                            {/* Age */}
                                            <div className="flex flex-col items-end min-w-[80px]">
                                                <span className="text-xs text-muted-foreground">
                                                    {formatAge(cj.age)}
                                                </span>
                                            </div>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </CardContent>
                </Card>
            </div>

            <ResourceDetailsSheet
                isOpen={!!selectedCronJob}
                onClose={() => setSelectedCronJob(null)}
                context={selectedContext}
                namespace={selectedCronJob?.namespace || ""}
                name={selectedCronJob?.name || ""}
                kind="CronJob"
                onUpdate={handleRefresh}
            />
        </div>
    );
}

export default function CronJobsPage() {
    return (
        <Suspense
            fallback={
                <div className="flex items-center justify-center min-h-screen">
                    <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
                </div>
            }
        >
            <CronJobsContent />
        </Suspense>
    );
}
