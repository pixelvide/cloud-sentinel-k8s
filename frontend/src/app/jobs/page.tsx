"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { PlayCircle, RefreshCw, CheckCircle2, XCircle, Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { cn, formatAge } from "@/lib/utils";
import { API_URL } from "@/lib/config";
import { NamespaceBadge } from "@/components/NamespaceBadge";
import { ResourceDetailsSheet } from "@/components/ResourceDetailsSheet";
import { LogViewerModal } from "@/components/LogViewerModal";
import { FileText } from "lucide-react";

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
    const searchParams = useSearchParams();
    const selectedContext = searchParams.get("context") || "";
    const selectedNamespace = searchParams.get("namespace") || "";

    const [jobs, setJobs] = useState<JobInfo[]>([]);
    const [loading, setLoading] = useState(false);
    const [searchQuery, setSearchQuery] = useState("");
    const [selectedJob, setSelectedJob] = useState<JobInfo | null>(null);
    const [logResource, setLogResource] = useState<{ namespace: string, selector: string, pods: Array<{ name: string, status: string }> } | null>(null);

    const filteredJobs = jobs.filter(j =>
        j.name.toLowerCase().includes(searchQuery.toLowerCase())
    );

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
            const res = await fetch(`${API_URL}/kube/jobs?context=${selectedContext}&namespace=${selectedNamespace}`, { credentials: "include" });
            if (res.status === 401) {
                window.location.href = "/login";
                return;
            }
            if (res.ok) {
                const data = await res.json();
                setJobs(data.jobs || []);
            }
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
                                    {!selectedContext || !selectedNamespace ? "Select a namespace from the top bar to view jobs" : null}
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
                        {jobs.length > 0 && (
                            <div className="relative mt-4">
                                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                                <Input
                                    placeholder="Search jobs by name..."
                                    value={searchQuery}
                                    onChange={(e) => setSearchQuery(e.target.value)}
                                    className="pl-10"
                                />
                            </div>
                        )}
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
                                <p className="text-muted-foreground text-sm">
                                    No jobs found in this namespace.
                                </p>
                            </div>
                        ) : (
                            <div className="grid grid-cols-1 gap-4">
                                {filteredJobs.map(j => {
                                    const status = getJobStatus(j);
                                    return (
                                        <div
                                            key={j.name + j.namespace}
                                            className="p-6 bg-muted/30 rounded-2xl border border-muted/20 hover:bg-muted/50 transition-colors cursor-pointer"
                                            onClick={() => setSelectedJob(j)}
                                        >
                                            <div className="flex flex-col lg:flex-row lg:items-center gap-4 justify-between">
                                                <div className="flex items-center gap-4 min-w-0">
                                                    <div className={cn(
                                                        "p-3 rounded-xl",
                                                        status === "Complete" ? "bg-green-500/10" :
                                                            status === "Failed" ? "bg-red-500/10" :
                                                                status === "Running" ? "bg-amber-500/10" : "bg-gray-500/10"
                                                    )}>
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
                                                            <span className={cn(
                                                                "text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full",
                                                                status === "Complete" ? "bg-green-500/10 text-green-600" :
                                                                    status === "Failed" ? "bg-red-500/10 text-red-600" :
                                                                        "bg-amber-500/10 text-amber-600"
                                                            )}>
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
                                                    <span className="text-xs text-muted-foreground">{formatAge(j.age)}</span>
                                                </div>

                                                {/* Action Buttons */}
                                                <div
                                                    className="flex flex-row items-center gap-2 min-w-[100px] justify-end"
                                                    onClick={(e) => e.stopPropagation()}
                                                >
                                                    <Button
                                                        variant="outline"
                                                        size="sm"
                                                        className="h-8 rounded-lg gap-2 text-xs font-semibold"
                                                        onClick={async (e) => {
                                                            e.stopPropagation();
                                                            if (j.selector) {
                                                                try {
                                                                    const res = await fetch(`${API_URL}/kube/pods?context=${selectedContext}&namespace=${j.namespace}&selector=${encodeURIComponent(j.selector)}`, { credentials: "include" });
                                                                    if (res.ok) {
                                                                        const data = await res.json();
                                                                        const pods = data.pods || [];
                                                                        setLogResource({
                                                                            namespace: j.namespace,
                                                                            selector: j.selector,
                                                                            pods: pods.map((p: any) => ({ name: p.name, status: p.status }))
                                                                        });
                                                                    }
                                                                } catch (error) {
                                                                    console.error("Failed to fetch pods:", error);
                                                                }
                                                            }
                                                        }}
                                                    >
                                                        <FileText className="h-3.5 w-3.5" />
                                                        Logs
                                                    </Button>
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
            />
            {logResource && (
                <LogViewerModal
                    isOpen={!!logResource}
                    onClose={() => setLogResource(null)}
                    context={selectedContext}
                    namespace={logResource.namespace}
                    pod={logResource.pods.length > 0 ? logResource.pods[0].name : ""}
                    selector={logResource.selector}
                    containers={["__all__"]}
                    initContainers={[]}
                    pods={logResource.pods}
                />
            )}
        </div >
    );
}

export default function JobsPage() {
    return (
        <Suspense fallback={<div className="flex items-center justify-center min-h-screen"><RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" /></div>}>
            <JobsContent />
        </Suspense>
    );
}
