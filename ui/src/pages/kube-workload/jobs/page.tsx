"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { PlayCircle, RefreshCw, CheckCircle2, XCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn, formatAge } from "@/lib/utils";
import { NamespaceBadge } from "@/components/NamespaceBadge";
import { ResourceDetailsSheet } from "@/components/ResourceDetailsSheet";
import { api } from "@/lib/api";

interface JobInfo {
    name: string;
    namespace: string;
    completions: number;
    succeeded: number;
    failed: number;
    active: number;
    age: string;
    selector?: string;
}

interface PodInfo {
    name: string;
    containers: string[];
    init_containers?: string[];
    namespace: string;
    status: string;
}

function JobsContent() {
    const [searchParams] = useSearchParams();
    const selectedContext = searchParams.get("context") || "";
    const selectedNamespace = searchParams.get("namespace") || "";

    const [jobs, setJobs] = useState<JobInfo[]>([]);
    const [loading, setLoading] = useState(false);
    const searchQuery = searchParams.get("q") || "";
    const [selectedJob, setSelectedJob] = useState<JobInfo | null>(null);

    const filteredJobs = jobs.filter((j) => j.name.toLowerCase().includes(searchQuery.toLowerCase()));

    useEffect(() => {
        if (!selectedContext || !selectedNamespace) {
            setJobs([]);
            return;
        }
        fetchJobs();
    }, [selectedContext, selectedNamespace]);

    const fetchJobs = async () => {
        setLoading(true);
        setJobs([]);
        try {
            const data = await api.get<any>(`/kube/jobs?namespace=${selectedNamespace}`, {
                headers: { "x-kube-context": selectedContext || "" },
            });
            setJobs(data.jobs || []);
        } catch (error) {
            console.error("Failed to fetch jobs:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleRefresh = () => {
        if (selectedContext && selectedNamespace) {
            fetchJobs();
        }
    };

    const getJobStatus = (job: JobInfo) => {
        if (job.succeeded >= job.completions) return "Complete";
        if (job.failed > 0) return "Failed";
        if (job.active > 0) return "Running";
        return "Pending";
    };

    return (
        <div className="flex flex-col items-center w-full min-h-screen bg-background/50 p-4 md:p-0">
            <div className="w-full max-w-6xl space-y-8">
                <Card className="border-none shadow-2xl shadow-black/5 bg-card/50 backdrop-blur-sm overflow-hidden rounded-3xl">
                    <CardHeader className="border-b bg-card/50 px-8 py-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <CardTitle className="text-lg font-bold flex items-center gap-3">
                                    <PlayCircle className="h-5 w-5 text-orange-500" />
                                    {jobs.length} Jobs Found
                                </CardTitle>
                                <CardDescription>
                                    {!selectedContext || !selectedNamespace
                                        ? "Select a namespace from the top bar to view jobs"
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
                                <PlayCircle className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    Select a cluster from the top bar to view jobs.
                                </p>
                            </div>
                        ) : !selectedNamespace ? (
                            <div className="text-center py-12">
                                <PlayCircle className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    Select a namespace from the top bar to view jobs.
                                </p>
                            </div>
                        ) : jobs.length === 0 ? (
                            <div className="text-center py-12">
                                <PlayCircle className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">No jobs found in this namespace.</p>
                            </div>
                        ) : (
                            <div className="grid grid-cols-1 gap-4">
                                {filteredJobs.map((j) => {
                                    const status = getJobStatus(j);
                                    return (
                                        <div
                                            key={j.name + j.namespace}
                                            className="p-6 bg-muted/30 rounded-2xl border border-muted/20 hover:bg-muted/50 transition-colors cursor-pointer"
                                            onClick={() => setSelectedJob(j)}
                                        >
                                            <div className="flex flex-col lg:flex-row lg:items-center gap-4 justify-between">
                                                <div className="flex items-center gap-4 min-w-0">
                                                    <div
                                                        className={cn(
                                                            "p-3 rounded-xl",
                                                            status === "Complete"
                                                                ? "bg-green-500/10"
                                                                : status === "Failed"
                                                                  ? "bg-red-500/10"
                                                                  : status === "Running"
                                                                    ? "bg-amber-500/10"
                                                                    : "bg-gray-500/10"
                                                        )}
                                                    >
                                                        {status === "Complete" ? (
                                                            <CheckCircle2 className="h-6 w-6 text-green-500" />
                                                        ) : status === "Failed" ? (
                                                            <XCircle className="h-6 w-6 text-red-500" />
                                                        ) : (
                                                            <RefreshCw className="h-6 w-6 text-amber-500 animate-pulse" />
                                                        )}
                                                    </div>
                                                    <div className="min-w-0">
                                                        <p className="font-semibold text-base truncate">{j.name}</p>
                                                        <div className="flex items-center gap-2 mt-1 flex-wrap">
                                                            <span
                                                                className={cn(
                                                                    "text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full",
                                                                    status === "Complete"
                                                                        ? "bg-green-500/10 text-green-600"
                                                                        : status === "Failed"
                                                                          ? "bg-red-500/10 text-red-600"
                                                                          : "bg-amber-500/10 text-amber-600"
                                                                )}
                                                            >
                                                                {j.succeeded}/{j.completions} completed
                                                            </span>
                                                        </div>
                                                    </div>
                                                </div>

                                                {/* Namespace */}
                                                <div
                                                    className="flex flex-col items-end min-w-[120px]"
                                                    onClick={(e) => e.stopPropagation()}
                                                >
                                                    <NamespaceBadge namespace={j.namespace} />
                                                </div>

                                                {/* Age */}
                                                <div className="flex flex-col items-end min-w-[80px]">
                                                    <span className="text-xs text-muted-foreground">
                                                        {formatAge(j.age)}
                                                    </span>
                                                </div>
                                            </div>
                                        </div>
                                    );
                                })}
                            </div>
                        )}
                    </CardContent>
                </Card>
            </div>

            <ResourceDetailsSheet
                isOpen={!!selectedJob}
                onClose={() => setSelectedJob(null)}
                context={selectedContext}
                namespace={selectedJob?.namespace || ""}
                name={selectedJob?.name || ""}
                kind="Job"
                onUpdate={handleRefresh}
            />
        </div>
    );
}

export default function JobsPage() {
    return (
        <Suspense
            fallback={
                <div className="flex items-center justify-center min-h-screen">
                    <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
                </div>
            }
        >
            <JobsContent />
        </Suspense>
    );
}
