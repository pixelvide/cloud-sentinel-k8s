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
import { Loader2 } from "lucide-react";

interface RelatedPodsTableProps {
    resource: any;
    context: string;
}

interface Pod {
    name: string;
    namespace: string;
    status: string;
    age: string;
    ready: string;
    restarts: number;
    node: string;
    ip: string;
    controlled_by?: string;
}

export function RelatedPodsTable({ resource, context }: RelatedPodsTableProps) {
    const [pods, setPods] = useState<Pod[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState("");

    const { kind, metadata, spec } = resource;

    useEffect(() => {
        const fetchPods = async () => {
            setLoading(true);
            setError("");
            setPods([]);

            try {
                let queryParams = "";

                if (kind === "Node") {
                    queryParams = `fieldSelector=spec.nodeName=${metadata.name}`;
                } else if (kind === "Service" || kind === "ReplicationController") {
                    if (spec?.selector) {
                        const selector = Object.entries(spec.selector)
                            .map(([k, v]) => `${k}=${v}`)
                            .join(",");
                        queryParams = `namespace=${metadata.namespace}&selector=${encodeURIComponent(selector)}`;
                    }
                } else if (["Deployment", "ReplicaSet", "DaemonSet", "StatefulSet", "Job"].includes(kind)) {
                    if (spec?.selector?.matchLabels) {
                        const selector = Object.entries(spec.selector.matchLabels)
                            .map(([k, v]) => `${k}=${v}`)
                            .join(",");
                        queryParams = `namespace=${metadata.namespace}&selector=${encodeURIComponent(selector)}`;
                    }
                    // Handle Job's matchLabels if slightly different? usually strict matchLabels. 
                    // Job selector is usually controller-uid but matchLabels works.
                }

                if (!queryParams) {
                    setLoading(false);
                    return; // No strategy to fetch pods
                }

                const data = await api.get<{ pods: Pod[] }>(`/kube/pods?${queryParams}`, {
                    headers: { "x-kube-context": context || "" }
                });
                setPods(data.pods || []);
            } catch (err: any) {
                console.error("Failed to fetch related pods:", err);
                setError(err.message || "Failed to fetch pods");
            } finally {
                setLoading(false);
            }
        };

        fetchPods();
    }, [resource, context, kind, metadata, spec]);

    if (!loading && pods.length === 0 && !error) return null;

    const getStatusVariant = (status: string) => {
        switch (status) {
            case "Running":
            case "Succeeded":
                return "default";
            case "Pending":
            case "ContainerCreating":
                return "secondary";
            case "Failed":
            case "ErrImagePull":
            case "CrashLoopBackOff":
            case "Error":
                return "destructive";
            default:
                return "outline";
        }
    };

    // Status text color helper
    const getStatusColor = (status: string) => {
        if (status === "Running" || status === "Succeeded") return "bg-emerald-50 text-emerald-700 border-emerald-200";
        if (status === "Pending" || status === "ContainerCreating") return "bg-amber-50 text-amber-700 border-amber-200";
        if (status === "Failed" || status === "CrashLoopBackOff" || status === "Error") return "bg-red-50 text-red-700 border-red-200";
        return "bg-muted text-muted-foreground border-border";
    };

    return (
        <div className="space-y-3">
            <div className="flex items-center justify-between">
                <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
                    Pods
                </h3>
                {pods.length > 0 && (
                    <Badge variant="secondary" className="bg-muted text-muted-foreground border-none text-[10px] font-bold px-1.5 h-4">
                        {pods.length}
                    </Badge>
                )}
            </div>

            <div className="rounded-md border border-border bg-card overflow-hidden shadow-sm overflow-x-auto">
                {loading ? (
                    <div className="p-4 flex items-center justify-center text-muted-foreground text-xs gap-2">
                        <Loader2 className="h-3 w-3 animate-spin" />
                        Fetching related pods...
                    </div>
                ) : error ? (
                    <div className="p-4 text-destructive text-xs">
                        {error}
                    </div>
                ) : pods.length > 0 ? (
                    <Table>
                        <TableHeader className="bg-muted text-muted-foreground">
                            <TableRow>
                                <TableHead className="h-8 text-[11px] font-medium whitespace-nowrap">Name</TableHead>
                                {kind !== "Namespace" && <TableHead className="h-8 text-[11px] font-medium whitespace-nowrap">Namespace</TableHead>}
                                <TableHead className="h-8 text-[11px] font-medium whitespace-nowrap">Ready</TableHead>
                                <TableHead className="h-8 text-[11px] font-medium whitespace-nowrap">Restarts</TableHead>
                                <TableHead className="h-8 text-[11px] font-medium whitespace-nowrap">Status</TableHead>
                                {kind !== "Node" && <TableHead className="h-8 text-[11px] font-medium whitespace-nowrap">Node</TableHead>}
                                <TableHead className="h-8 text-[11px] font-medium text-right whitespace-nowrap">Age</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {pods.map((pod) => (
                                <TableRow key={pod.name + pod.namespace} className="hover:bg-muted/50 border-border">
                                    <TableCell className="py-2 text-xs font-mono font-medium whitespace-nowrap">{pod.name}</TableCell>
                                    {kind !== "Namespace" && <TableCell className="py-2 text-xs text-muted-foreground whitespace-nowrap">{pod.namespace}</TableCell>}
                                    <TableCell className="py-2 text-xs text-muted-foreground whitespace-nowrap">{pod.ready}</TableCell>
                                    <TableCell className="py-2 text-xs text-muted-foreground whitespace-nowrap">{pod.restarts}</TableCell>
                                    <TableCell className="py-2 whitespace-nowrap">
                                        <Badge
                                            variant="outline"
                                            className={cn("text-[9px] font-bold px-1.5 py-0 h-4 border", getStatusColor(pod.status))}
                                        >
                                            {pod.status}
                                        </Badge>
                                    </TableCell>
                                    {kind !== "Node" && <TableCell className="py-2 text-xs text-muted-foreground whitespace-nowrap">{pod.node}</TableCell>}
                                    <TableCell className="py-2 text-xs text-muted-foreground text-right whitespace-nowrap">
                                        {formatAge(pod.age)}
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                ) : (
                    <div className="p-4 text-center text-xs text-muted-foreground italic">
                        No related pods found.
                    </div>
                )}
            </div>
        </div>
    );
}
