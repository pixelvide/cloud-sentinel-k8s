"use client";

import { useEffect, useState } from "react";
import { Badge } from "@/components/ui/badge";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { api } from "@/lib/api";
import { formatAge } from "@/lib/utils";
import { cn } from "@/lib/utils";
import { Loader2, CheckCircle2, XCircle, Clock } from "lucide-react";

interface RelatedJobsTableProps {
    resource: any;
    context: string;
}

interface Job {
    name: string;
    namespace: string;
    completions: number;
    succeeded: number;
    failed: number;
    active: number;
    age: string;
}

export function RelatedJobsTable({ resource, context }: RelatedJobsTableProps) {
    const [jobs, setJobs] = useState<Job[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState("");

    const { kind, metadata } = resource;

    useEffect(() => {
        const fetchJobs = async () => {
            if (kind !== "CronJob") return;

            setLoading(true);
            setError("");
            setJobs([]);

            try {
                // For CronJobs, we can find Jobs by owner reference UID.
                // But since our API doesn't support ownerUid filtering directly, 
                // and we've added label selector support, we'll try to find by labels.
                // CronJob controller adds 'controller-uid' label to Jobs.

                let queryParams = `namespace=${metadata.namespace}`;

                if (metadata.uid) {
                    queryParams += `&ownerUid=${metadata.uid}`;
                } else {
                    setLoading(false);
                    return;
                }

                const data = await api.get<{ jobs: Job[] }>(`/kube/jobs?${queryParams}`, {
                    headers: { "x-kube-context": context || "" }
                });

                // Sort by age (newest first)
                const sortedJobs = (data.jobs || []).sort((a, b) =>
                    new Date(b.age).getTime() - new Date(a.age).getTime()
                );

                setJobs(sortedJobs);
            } catch (err: any) {
                console.error("Failed to fetch related jobs:", err);
                setError(err.message || "Failed to fetch jobs");
            } finally {
                setLoading(false);
            }
        };

        fetchJobs();
    }, [resource, context, kind, metadata]);

    if (kind !== "CronJob") return null;
    if (!loading && jobs.length === 0 && !error) return null;

    const getStatusIcon = (job: Job) => {
        if (job.succeeded > 0) return <CheckCircle2 className="h-3 w-3 text-emerald-500" />;
        if (job.failed > 0) return <XCircle className="h-3 w-3 text-red-500" />;
        if (job.active > 0) return <Clock className="h-3 w-3 text-amber-500 animate-pulse" />;
        return <Clock className="h-3 w-3 text-muted-foreground" />;
    };

    const getStatusText = (job: Job) => {
        if (job.succeeded > 0) return "Succeeded";
        if (job.failed > 0) return "Failed";
        if (job.active > 0) return "Running";
        return "Unknown";
    };

    const getStatusColor = (job: Job) => {
        if (job.succeeded > 0) return "bg-emerald-50 text-emerald-700 border-emerald-200";
        if (job.failed > 0) return "bg-red-50 text-red-700 border-red-200";
        if (job.active > 0) return "bg-amber-50 text-amber-700 border-amber-200";
        return "bg-muted text-muted-foreground border-border";
    };

    return (
        <div className="space-y-3">
            <div className="flex items-center justify-between">
                <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
                    Related Job History
                </h3>
                {jobs.length > 0 && (
                    <Badge variant="secondary" className="bg-muted text-muted-foreground border-none text-[10px] font-bold px-1.5 h-4">
                        {jobs.length}
                    </Badge>
                )}
            </div>

            <div className="rounded-md border border-border bg-card overflow-hidden shadow-sm overflow-x-auto">
                {loading ? (
                    <div className="p-4 flex items-center justify-center text-muted-foreground text-xs gap-2">
                        <Loader2 className="h-3 w-3 animate-spin" />
                        Fetching job history...
                    </div>
                ) : error ? (
                    <div className="p-4 text-destructive text-xs">
                        {error}
                    </div>
                ) : jobs.length > 0 ? (
                    <Table>
                        <TableHeader className="bg-muted text-muted-foreground">
                            <TableRow>
                                <TableHead className="h-8 text-[11px] font-medium whitespace-nowrap">Job Name</TableHead>
                                <TableHead className="h-8 text-[11px] font-medium whitespace-nowrap">Completions</TableHead>
                                <TableHead className="h-8 text-[11px] font-medium whitespace-nowrap">Status</TableHead>
                                <TableHead className="h-8 text-[11px] font-medium text-right whitespace-nowrap">Age</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {jobs.map((job) => (
                                <TableRow key={job.name} className="hover:bg-muted/50 border-border">
                                    <TableCell className="py-2 text-xs font-mono font-medium whitespace-nowrap">
                                        {job.name}
                                    </TableCell>
                                    <TableCell className="py-2 text-xs text-muted-foreground whitespace-nowrap">
                                        {job.succeeded}/{job.completions}
                                    </TableCell>
                                    <TableCell className="py-2 whitespace-nowrap">
                                        <Badge
                                            variant="outline"
                                            className={cn("text-[9px] font-bold px-1.5 py-0 h-4 border flex items-center gap-1 w-fit", getStatusColor(job))}
                                        >
                                            {getStatusIcon(job)}
                                            {getStatusText(job)}
                                        </Badge>
                                    </TableCell>
                                    <TableCell className="py-2 text-xs text-muted-foreground text-right whitespace-nowrap">
                                        {formatAge(job.age)}
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                ) : (
                    <div className="p-4 text-center text-xs text-muted-foreground italic">
                        No job history found.
                    </div>
                )}
            </div>
        </div>
    );
}
