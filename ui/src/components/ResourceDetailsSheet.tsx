"use client";

import { useEffect, useState } from "react";
import {
    Sheet,
    SheetContent,
    SheetHeader,
    SheetTitle,
    SheetDescription,
} from "@/components/ui/sheet";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogDescription,
    DialogFooter,
} from "@/components/ui/dialog";
import { ConfirmDialog } from "@/components/ConfirmDialog";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Badge } from "@/components/ui/badge";
import { KubeProperties } from "@/components/KubeProperties";
import { Button } from "@/components/ui/button";
import { Terminal as TerminalIcon, FileText, Ban, Trash2, CheckCircle2, Edit, PauseCircle, PlayCircle, Play, Maximize2, Minus, Plus, RotateCcw, Copy, Activity } from "lucide-react";
import { Textarea } from "@/components/ui/textarea";
import { LogViewerModal } from "@/components/LogViewerModal";
import { api } from "@/lib/api";
import { toast } from "sonner";
import { cn } from "@/lib/utils";
import { RelatedPodsTable } from "@/components/RelatedPodsTable";
import { RelatedPVsTable } from "@/components/RelatedPVsTable";
import { RelatedJobsTable } from "@/components/RelatedJobsTable";
import { MetricsChart } from "@/components/shared/MetricsChart";


interface ResourceDetailsSheetProps {
    isOpen: boolean;
    onClose: () => void;
    context: string;
    namespace: string;
    name: string;
    kind: string;
    crdName?: string;
    onUpdate?: () => void;
}

interface EventSimple {
    type: string;
    reason: string;
    message: string;
    count: number;
    last_seen: string;
    age: string;
}

interface ResourceInfo {
    group: string;
    version: string;
    resource: string;
    scope: string;
    kind: string;
}

interface Anomaly {
    id: string;
    type: string;
    severity: "Critical" | "Warning" | "Info";
    message: string;
    description: string;
    suggestion: string;
    created_at: string;
}

interface ResourceAnalysis {
    anomalies: Anomaly[];
    summary: string;
    score: number;
}

interface ResourceDetails {
    manifest: string;
    events: EventSimple[];
    analysis?: ResourceAnalysis;
    raw: any;
}

export function ResourceDetailsSheet({
    isOpen,
    onClose,
    context,
    namespace,
    name,
    kind,
    crdName,
    onUpdate,
}: ResourceDetailsSheetProps) {
    const [details, setDetails] = useState<ResourceDetails | null>(null);
    const [loading, setLoading] = useState(false);
    const [actioning, setActioning] = useState(false);
    const [confirmConfig, setConfirmConfig] = useState<{
        isOpen: boolean;
        title: string;
        description: React.ReactNode;
        confirmText: string;
        confirmVariant: "default" | "destructive";
        onConfirm: () => void;
    }>({
        isOpen: false,
        title: "",
        description: null,
        confirmText: "Confirm",
        confirmVariant: "default",
        onConfirm: () => { },
    });
    const [error, setError] = useState("");
    const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
    const [editManifest, setEditManifest] = useState("");
    const [isRunDialogOpen, setIsRunDialogOpen] = useState(false);
    const [runJobName, setRunJobName] = useState("");
    const [isScaleDialogOpen, setIsScaleDialogOpen] = useState(false);
    const [scaleReplicas, setScaleReplicas] = useState(0);

    const [scopes, setScopes] = useState<Record<string, ResourceInfo>>({});
    const [logResource, setLogResource] = useState<{
        name: string,
        namespace: string,
        selector?: string,
        pods: Array<{ name: string, status: string }>,
        containers: string[],
        initContainers: string[]
    } | null>(null);

    const [metrics, setMetrics] = useState<{
        cpu: Array<{ timestamp: string; value: number }>;
        memory: Array<{ timestamp: string; value: number }>;
        fallback: boolean;
    } | null>(null);

    useEffect(() => {
        // Fetch scopes once
        api.get<any>("/kube/scopes")
            .then(data => setScopes(data.scopes || {}))
            .catch(err => console.error("Failed to fetch scopes:", err));
    }, []);

    useEffect(() => {
        if (isOpen && kind === "Pod" && namespace && name && context) {
            setMetrics(null);
            api.getPodMetrics(context, namespace, name)
                .then(setMetrics)
                .catch(err => console.error("Failed to fetch metrics:", err));
        } else {
            setMetrics(null);
        }
    }, [isOpen, kind, namespace, name, context]);

    const fetchDetails = () => {
        if (!isOpen || !context || !name) return;

        if (crdName) {
            setLoading(true);
            setError("");
            const url = `/kube/crds/${encodeURIComponent(crdName)}/resources/${encodeURIComponent(name)}?namespace=${encodeURIComponent(namespace)}`;
            api.get<ResourceDetails>(url, {
                headers: { "x-kube-context": context || "" }
            })
                .then((data) => setDetails(data))
                .catch((err) => setError(err.message))
                .finally(() => setLoading(false));
            return;
        }

        if (kind && Object.keys(scopes).length > 0) {
            const resourceInfo = scopes[kind];
            const isClusterScoped = resourceInfo?.scope === "Cluster";

            if (isClusterScoped || namespace) {
                setLoading(true);
                setError("");

                const url = `/kube/resource?namespace=${namespace}&name=${name}&kind=${kind}`;
                api.get<ResourceDetails>(url, {
                    headers: { "x-kube-context": context || "" }
                })
                    .then((data) => setDetails(data))
                    .catch((err) => setError(err.message))
                    .finally(() => setLoading(false));
            }
        }
    };

    useEffect(() => {
        if (isOpen) {
            fetchDetails();
        }
    }, [isOpen, context, namespace, name, kind, scopes, crdName]);

    return (
        <Sheet open={isOpen} onOpenChange={(open) => !open && onClose()}>
            <SheetContent className="w-full sm:max-w-2xl overflow-y-auto bg-background border-l border-border p-0 flex flex-col h-full">
                <SheetHeader className="p-6 border-b border-border shrink-0 bg-background/50 backdrop-blur-sm">
                    <SheetTitle className="text-xl font-bold font-mono text-foreground">
                        {kind}: {name}
                    </SheetTitle>
                    <SheetDescription className="text-muted-foreground font-mono text-xs">
                        {namespace} @ {context}
                    </SheetDescription>

                    {details && (
                        <div className="flex items-center gap-2 mt-4">
                            {kind === "Pod" && (
                                <Button
                                    variant="outline"
                                    size="sm"
                                    className="h-8 rounded-lg gap-2 text-xs font-semibold bg-background shadow-sm border-border text-foreground hover:bg-accent"
                                    disabled={details.raw?.status?.phase !== "Running"}
                                    onClick={() => {
                                        const pod = details.raw;
                                        const container = pod.spec?.containers?.[0]?.name || "";
                                        window.open(`/exec?context=${context}&namespace=${namespace}&pod=${name}&container=${container}`, "_blank");
                                    }}
                                >
                                    <TerminalIcon className="h-3.5 w-3.5 text-muted-foreground" />
                                    Terminal
                                </Button>
                            )}

                            {["Pod", "Deployment", "ReplicaSet", "StatefulSet", "DaemonSet", "Job", "ReplicationController"].includes(kind) && (
                                <Button
                                    variant="outline"
                                    size="sm"
                                    className="h-8 rounded-lg gap-2 text-xs font-semibold bg-background shadow-sm border-border text-foreground hover:bg-accent"
                                    onClick={async () => {
                                        const resource = details.raw;
                                        if (kind === "Pod") {
                                            setLogResource({
                                                name: resource.metadata.name,
                                                namespace: resource.metadata.namespace,
                                                pods: [{ name: resource.metadata.name, status: resource.status?.phase }],
                                                containers: resource.spec?.containers?.map((c: any) => c.name) || [],
                                                initContainers: resource.spec?.initContainers?.map((c: any) => c.name) || [],
                                            });
                                        } else {
                                            const matchLabels = resource.spec?.selector?.matchLabels;
                                            if (matchLabels) {
                                                const selector = Object.entries(matchLabels).map(([k, v]) => `${k}=${v}`).join(",");
                                                try {
                                                    const data = await api.get<any>(`/kube/pods?namespace=${namespace}&selector=${encodeURIComponent(selector)}`, {
                                                        headers: { "x-kube-context": context || "" }
                                                    });
                                                    const pods = data.pods || [];
                                                    setLogResource({
                                                        name: name,
                                                        namespace: namespace,
                                                        selector: selector,
                                                        pods: pods.map((p: any) => ({ name: p.name, status: p.status })),
                                                        containers: (pods.length > 0 && pods[0].containers) ? pods[0].containers : ["__all__"],
                                                        initContainers: (pods.length > 0 && pods[0].init_containers) ? pods[0].init_containers : []
                                                    });
                                                } catch (error) {
                                                    console.error("Failed to fetch pods:", error);
                                                }
                                            }
                                        }
                                    }}
                                >
                                    <FileText className="h-3.5 w-3.5 text-muted-foreground" />
                                    Logs
                                </Button>
                            )}

                            {kind === "Node" && (
                                <>
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        className="h-8 rounded-lg gap-2 text-xs font-semibold bg-background shadow-sm border-border text-foreground hover:bg-accent"
                                        disabled={actioning || !details}
                                        onClick={async () => {
                                            const isUnschedulable = details?.raw?.spec?.unschedulable;
                                            setConfirmConfig({
                                                isOpen: true,
                                                title: isUnschedulable ? "Uncordon Node" : "Cordon Node",
                                                description: (
                                                    <>
                                                        Are you sure you want to {isUnschedulable ? "uncordon" : "cordon"} node <span className="font-mono font-bold text-foreground">{name}</span>?
                                                    </>
                                                ),
                                                confirmText: isUnschedulable ? "Uncordon" : "Cordon",
                                                confirmVariant: "default",
                                                onConfirm: async () => {
                                                    setActioning(true);
                                                    setConfirmConfig(prev => ({ ...prev, isOpen: false }));
                                                    try {
                                                        await api.post(`/kube/nodes/cordon?name=${name}`, {
                                                            unschedulable: !isUnschedulable
                                                        }, {
                                                            headers: { "x-kube-context": context || "" }
                                                        });
                                                        toast.success(`Node ${name} ${isUnschedulable ? "uncordoned" : "cordoned"}`);
                                                        fetchDetails();
                                                        onUpdate?.();
                                                    } catch (err: any) {
                                                        toast.error(err.message || "Action failed");
                                                    } finally {
                                                        setActioning(false);
                                                    }
                                                }
                                            });
                                        }}
                                    >
                                        {details?.raw?.spec?.unschedulable ? (
                                            <>
                                                <CheckCircle2 className="h-3.5 w-3.5 text-green-500" />
                                                Uncordon
                                            </>
                                        ) : (
                                            <>
                                                <Ban className="h-3.5 w-3.5 text-destructive" />
                                                Cordon
                                            </>
                                        )}
                                    </Button>

                                    <Button
                                        variant="outline"
                                        size="sm"
                                        className="h-8 rounded-lg gap-2 text-xs font-semibold bg-background shadow-sm border-border text-foreground hover:bg-accent"
                                        disabled={actioning || !details}
                                        onClick={() => {
                                            setConfirmConfig({
                                                isOpen: true,
                                                title: `Drain Node: ${name}`,
                                                description: (
                                                    <>
                                                        Are you sure you want to drain node <span className="font-mono font-bold text-foreground">{name}</span>?
                                                        This will cordon the node and evict all pods. This action cannot be undone.
                                                    </>
                                                ),
                                                confirmText: "Confirm Drain",
                                                confirmVariant: "destructive",
                                                onConfirm: async () => {
                                                    setActioning(true);
                                                    setConfirmConfig(prev => ({ ...prev, isOpen: false }));
                                                    try {
                                                        const res = await api.post<any>(`/kube/nodes/drain?name=${name}`, {}, {
                                                            headers: { "x-kube-context": context || "" }
                                                        });
                                                        toast.success(`Drain started: ${res.evicted} pods evicted, ${res.skipped} skipped.`);
                                                        fetchDetails();
                                                        onUpdate?.();
                                                    } catch (err: any) {
                                                        toast.error(err.message || "Drain failed");
                                                    } finally {
                                                        setActioning(false);
                                                    }
                                                }
                                            });
                                        }}
                                    >
                                        <Trash2 className="h-3.5 w-3.5 text-muted-foreground" />
                                        Drain
                                    </Button>
                                </>
                            )}

                            {kind === "CronJob" && (
                                <Button
                                    variant="outline"
                                    size="sm"
                                    className="h-8 rounded-lg gap-2 text-xs font-semibold bg-background shadow-sm border-border text-foreground hover:bg-accent"
                                    disabled={actioning || !details}
                                    onClick={async () => {
                                        const isSuspended = details?.raw?.spec?.suspend;
                                        setConfirmConfig({
                                            isOpen: true,
                                            title: isSuspended ? "Resume CronJob" : "Suspend CronJob",
                                            description: (
                                                <>
                                                    Are you sure you want to {isSuspended ? "resume" : "suspend"} CronJob <span className="font-mono font-bold text-foreground">{name}</span>?
                                                </>
                                            ),
                                            confirmText: isSuspended ? "Resume" : "Suspend",
                                            confirmVariant: "default",
                                            onConfirm: async () => {
                                                setActioning(true);
                                                setConfirmConfig(prev => ({ ...prev, isOpen: false }));
                                                try {
                                                    await api.post(`/kube/cron-jobs/suspend?namespace=${namespace}&name=${name}`, {
                                                        suspend: !isSuspended
                                                    }, {
                                                        headers: { "x-kube-context": context || "" }
                                                    });
                                                    toast.success(`CronJob ${name} ${isSuspended ? "resumed" : "suspended"}`);
                                                    fetchDetails();
                                                    onUpdate?.();
                                                } catch (err: any) {
                                                    toast.error(err.message || "Action failed");
                                                } finally {
                                                    setActioning(false);
                                                }
                                            }
                                        });
                                    }}
                                >
                                    {details?.raw?.spec?.suspend ? (
                                        <>
                                            <PlayCircle className="h-3.5 w-3.5 text-teal-500" />
                                            Resume
                                        </>
                                    ) : (
                                        <>
                                            <PauseCircle className="h-3.5 w-3.5 text-orange-500" />
                                            Suspend
                                        </>
                                    )}
                                </Button>
                            )}

                            {kind === "CronJob" && (
                                <Button
                                    variant="outline"
                                    size="sm"
                                    className="h-8 rounded-lg gap-2 text-xs font-semibold bg-background shadow-sm border-border text-foreground hover:bg-accent"
                                    disabled={actioning || !details}
                                    onClick={() => {
                                        const timestamp = Math.floor(Date.now() / 1000);
                                        setRunJobName(`${name}-manual-${timestamp}`);
                                        setIsRunDialogOpen(true);
                                    }}
                                >
                                    <Play className="h-3.5 w-3.5 text-blue-500" />
                                    Run
                                </Button>
                            )}

                            {["Deployment", "StatefulSet", "ReplicaSet"].includes(kind) && (
                                <Button
                                    variant="outline"
                                    size="sm"
                                    className="h-8 rounded-lg gap-2 text-xs font-semibold bg-background shadow-sm border-border text-foreground hover:bg-accent"
                                    disabled={actioning || !details}
                                    onClick={() => {
                                        const replicas = details?.raw?.spec?.replicas || 0;
                                        setScaleReplicas(replicas);
                                        setIsScaleDialogOpen(true);
                                    }}
                                >
                                    <Maximize2 className="h-3.5 w-3.5 text-blue-500" />
                                    Scale
                                </Button>
                            )}

                            {["Deployment", "StatefulSet", "DaemonSet"].includes(kind) && (
                                <Button
                                    variant="outline"
                                    size="sm"
                                    className="h-8 rounded-lg gap-2 text-xs font-semibold bg-background shadow-sm border-border text-foreground hover:bg-accent"
                                    disabled={actioning || !details}
                                    onClick={() => {
                                        setConfirmConfig({
                                            isOpen: true,
                                            title: `Restart ${kind}: ${name}`,
                                            description: (
                                                <>
                                                    Are you sure you want to trigger a rollout restart for <span className="font-mono font-bold text-foreground">{kind}</span> <span className="font-mono font-bold text-foreground">{name}</span>?
                                                    This will restart all pods in the resource.
                                                </>
                                            ),
                                            confirmText: "Restart",
                                            confirmVariant: "default",
                                            onConfirm: async () => {
                                                setActioning(true);
                                                setConfirmConfig(prev => ({ ...prev, isOpen: false }));
                                                try {
                                                    await api.post(`/kube/resource/restart?namespace=${namespace}&name=${name}&kind=${kind}`, {}, {
                                                        headers: { "x-kube-context": context || "" }
                                                    });
                                                    toast.success(`${kind} ${name} restart triggered`);
                                                    fetchDetails();
                                                    onUpdate?.();
                                                } catch (err: any) {
                                                    toast.error(err.message || "Restart failed");
                                                } finally {
                                                    setActioning(false);
                                                }
                                            }
                                        });
                                    }}
                                >
                                    <RotateCcw className="h-3.5 w-3.5 text-orange-500" />
                                    Restart
                                </Button>
                            )}

                            <Button
                                variant="outline"
                                size="sm"
                                className="h-8 rounded-lg gap-2 text-xs font-semibold bg-background shadow-sm border-border text-foreground hover:bg-accent"
                                disabled={actioning || !details}
                                onClick={() => {
                                    setEditManifest(details?.manifest || "");
                                    setIsEditDialogOpen(true);
                                }}
                            >
                                <Edit className="h-3.5 w-3.5 text-muted-foreground" />
                                Edit
                            </Button>

                            <Button
                                variant="outline"
                                size="sm"
                                className="h-8 rounded-lg gap-2 text-xs font-semibold bg-background shadow-sm border-border text-destructive hover:bg-destructive hover:text-destructive-foreground transition-colors"
                                disabled={actioning || !details}
                                onClick={() => {
                                    setConfirmConfig({
                                        isOpen: true,
                                        title: `Delete ${kind}: ${name}`,
                                        description: (
                                            <>
                                                Are you sure you want to delete <span className="font-mono font-bold text-foreground">{kind}</span> <span className="font-mono font-bold text-foreground">{name}</span>?
                                                This action is permanent and cannot be undone.
                                            </>
                                        ),
                                        confirmText: "Delete",
                                        confirmVariant: "destructive",
                                        onConfirm: async () => {
                                            setActioning(true);
                                            setConfirmConfig(prev => ({ ...prev, isOpen: false }));
                                            try {
                                                await api.del(`/kube/resource?namespace=${namespace}&name=${name}&kind=${kind}`, {
                                                    headers: { "x-kube-context": context || "" }
                                                });
                                                toast.success(`${kind} ${name} deleted successfully`);
                                                onClose();
                                                onUpdate?.();
                                            } catch (err: any) {
                                                toast.error(err.message || "Delete failed");
                                            } finally {
                                                setActioning(false);
                                            }
                                        }
                                    });
                                }}
                            >
                                <Trash2 className="h-3.5 w-3.5" />
                                Delete
                            </Button>
                        </div>
                    )}
                </SheetHeader>

                <div className="flex-1 overflow-y-auto">
                    {loading && (
                        <div className="p-6 text-muted-foreground font-mono text-sm animate-pulse">
                            Loading details...
                        </div>
                    )}

                    {error && (
                        <div className="p-6 text-destructive font-mono text-sm">
                            Error: {error}
                        </div>
                    )}

                    {details && (
                        <div className="flex flex-col gap-6 p-6">
                            {/* Metrics Section */}
                            {kind === "Pod" && metrics && (
                                <div className="space-y-3">
                                    <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider flex items-center gap-2">
                                        <Activity className="h-4 w-4" />
                                        Performance
                                    </h3>
                                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                        <MetricsChart
                                            title="CPU Usage"
                                            data={metrics.cpu}
                                            unit=" Cores"
                                            color="#8b5cf6" // violet
                                            fallback={metrics.fallback}
                                        />
                                        <MetricsChart
                                            title="Memory Usage"
                                            data={metrics.memory}
                                            unit=" MiB"
                                            color="#ec4899" // pink
                                            fallback={metrics.fallback}
                                        />
                                    </div>
                                </div>
                            )}

                            <KubeProperties resource={details.raw} />
                            {/* Insights Section */}
                            {details.analysis && details.analysis.anomalies?.length > 0 && (
                                <div className="space-y-4">
                                    <div className="flex items-center justify-between">
                                        <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider flex items-center gap-2">
                                            Insights & Anomalies
                                        </h3>
                                        <Badge variant="outline" className={cn(
                                            "font-mono text-[10px]",
                                            details.analysis.score >= 90 ? "text-emerald-500 border-emerald-500/20" :
                                                details.analysis.score >= 70 ? "text-amber-500 border-amber-500/20" :
                                                    "text-destructive border-destructive/20"
                                        )}>
                                            Score: {details.analysis.score}/100
                                        </Badge>
                                    </div>
                                    <div className="space-y-3">
                                        {details.analysis.anomalies.map((anomaly) => (
                                            <div
                                                key={anomaly.id}
                                                className={cn(
                                                    "p-4 rounded-xl border shadow-sm space-y-2",
                                                    anomaly.severity === "Critical" ? "bg-red-500/5 border-red-500/20" :
                                                        anomaly.severity === "Warning" ? "bg-amber-500/5 border-amber-500/20" :
                                                            "bg-blue-500/5 border-blue-500/20"
                                                )}
                                            >
                                                <div className="flex items-start justify-between gap-2">
                                                    <div className="flex items-center gap-2">
                                                        <Badge className={cn(
                                                            "text-[10px] font-bold uppercase tracking-tight px-1.5 py-0 h-auto",
                                                            anomaly.severity === "Critical" ? "bg-red-500 text-white" :
                                                                anomaly.severity === "Warning" ? "bg-amber-500 text-white" :
                                                                    "bg-blue-500 text-white"
                                                        )}>
                                                            {anomaly.severity}
                                                        </Badge>
                                                        <span className="text-xs font-bold text-foreground">{anomaly.message}</span>
                                                    </div>
                                                </div>
                                                <p className="text-xs text-muted-foreground leading-relaxed">
                                                    {anomaly.description}
                                                </p>
                                                {anomaly.suggestion && (
                                                    <div className="pt-2 flex items-start gap-2 border-t border-border/10 mt-2">
                                                        <span className="text-[10px] font-bold uppercase text-primary shrink-0 mt-0.5">Suggestion:</span>
                                                        <p className="text-xs text-foreground/80 italic">
                                                            {anomaly.suggestion}
                                                        </p>
                                                    </div>
                                                )}
                                            </div>
                                        ))}
                                    </div>
                                </div>
                            )}

                            <RelatedPodsTable resource={details.raw} context={context} />
                            <RelatedPVsTable resource={details.raw} context={context} />
                            <RelatedJobsTable resource={details.raw} context={context} />

                            {/* Events Section */}
                            <div className="space-y-3">
                                <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
                                    Recent Events
                                </h3>
                                {details.events?.length > 0 ? (
                                    <div className="rounded-md border border-border bg-card overflow-hidden shadow-sm">
                                        <table className="w-full text-xs text-left font-mono">
                                            <thead className="bg-muted text-muted-foreground border-b border-border">
                                                <tr>
                                                    <th className="p-2 font-medium">Type</th>
                                                    <th className="p-2 font-medium">Reason</th>
                                                    <th className="p-2 font-medium">Age</th>
                                                    <th className="p-2 font-medium">Message</th>
                                                </tr>
                                            </thead>
                                            <tbody className="divide-y divide-border">
                                                {details.events.map((e, i) => (
                                                    <tr key={i} className="hover:bg-muted/50 transition-colors">
                                                        <td className="p-2">
                                                            <Badge
                                                                variant="outline"
                                                                className={
                                                                    e.type === "Warning"
                                                                        ? "text-destructive border-destructive/20 bg-destructive/10"
                                                                        : "text-muted-foreground border-border bg-muted/30"
                                                                }
                                                            >
                                                                {e.type}
                                                            </Badge>
                                                        </td>
                                                        <td className="p-2 text-foreground/80">{e.reason}</td>
                                                        <td className="p-2 text-muted-foreground whitespace-nowrap text-[10px]">
                                                            {/* TODO: Format age better if needed, backend sends RFC3339 */}
                                                            {new Date(e.last_seen).toLocaleTimeString()}
                                                        </td>
                                                        <td className="p-2 text-muted-foreground break-words max-w-[200px]">
                                                            {e.message} ({e.count})
                                                        </td>
                                                    </tr>
                                                ))}
                                            </tbody>
                                        </table>
                                    </div>
                                ) : (
                                    <div className="text-muted-foreground text-sm italic py-4">
                                        No events found.
                                    </div>
                                )}
                            </div>

                            {/* Manifest Section */}
                            <div className="space-y-3">
                                <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider flex items-center justify-between">
                                    YAML Manifest
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        className="h-6 w-6 ml-auto"
                                        onClick={() => {
                                            if (details.manifest) {
                                                navigator.clipboard.writeText(details.manifest);
                                                toast.success("Copied to clipboard");
                                            }
                                        }}
                                        title="Copy YAML"
                                    >
                                        <Copy className="h-3.5 w-3.5 text-muted-foreground" />
                                    </Button>
                                </h3>
                                <div className="relative rounded-md border border-border bg-card p-4 text-xs font-mono text-card-foreground overflow-auto max-h-[600px] shadow-sm">
                                    <pre className="whitespace-pre-wrap break-all">{details.manifest}</pre>
                                </div>
                            </div>
                        </div>
                    )}
                </div>
            </SheetContent >

            {
                logResource && (
                    <LogViewerModal
                        isOpen={!!logResource}
                        onClose={() => setLogResource(null)}
                        context={context}
                        namespace={logResource.namespace}
                        selector={logResource.selector}
                        containers={logResource.containers}
                        initContainers={logResource.initContainers}
                        pods={logResource.pods}
                        showPodSelector={kind !== "Pod"}
                        title={logResource.name}
                    />
                )
            }

            <ConfirmDialog
                isOpen={confirmConfig.isOpen}
                onClose={() => setConfirmConfig(prev => ({ ...prev, isOpen: false }))}
                onConfirm={confirmConfig.onConfirm}
                title={confirmConfig.title}
                description={confirmConfig.description}
                confirmText={confirmConfig.confirmText}
                confirmVariant={confirmConfig.confirmVariant}
                loading={actioning}
            />

            <Dialog open={isEditDialogOpen} onOpenChange={setIsEditDialogOpen}>
                <DialogContent className="max-w-4xl h-[90vh] flex flex-col p-0 gap-0 border-border bg-background">
                    <DialogHeader className="p-6 border-b border-border shrink-0">
                        <DialogTitle className="text-xl font-bold font-mono flex items-center gap-2">
                            <Edit className="h-5 w-5" />
                            Edit {kind}: {name}
                        </DialogTitle>
                        <DialogDescription className="font-mono text-xs">
                            Modify the YAML manifest below.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="flex-1 min-h-0 p-0 overflow-hidden bg-muted/30">
                        <Textarea
                            value={editManifest}
                            onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setEditManifest(e.target.value)}
                            className="w-full h-full p-6 font-mono text-sm bg-transparent border-none focus-visible:ring-0 resize-none overflow-auto"
                            spellCheck={false}
                        />
                    </div>
                    <DialogFooter className="p-4 border-t border-border bg-background/50 backdrop-blur-sm shrink-0">
                        <Button
                            variant="outline"
                            onClick={() => setIsEditDialogOpen(false)}
                            disabled={actioning}
                        >
                            Cancel
                        </Button>
                        <Button
                            onClick={async () => {
                                setActioning(true);
                                try {
                                    await api.put(`/kube/resource?namespace=${namespace}&name=${name}&kind=${kind}`, {
                                        manifest: editManifest
                                    }, {
                                        headers: { "x-kube-context": context || "" }
                                    });
                                    toast.success(`${kind} ${name} updated successfully`);
                                    setIsEditDialogOpen(false);
                                    fetchDetails();
                                    onUpdate?.();
                                } catch (err: any) {
                                    toast.error(err.message || "Update failed");
                                } finally {
                                    setActioning(false);
                                }
                            }}
                            disabled={actioning}
                        >
                            {actioning ? "Saving..." : "Save Changes"}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            <Dialog open={isRunDialogOpen} onOpenChange={setIsRunDialogOpen}>
                <DialogContent className="sm:max-w-[425px] bg-background border-border">
                    <DialogHeader>
                        <DialogTitle className="font-mono flex items-center gap-2">
                            <Play className="h-5 w-5 text-blue-500" />
                            Trigger CronJob
                        </DialogTitle>
                        <DialogDescription>
                            Manually trigger a job from <span className="font-mono font-bold text-foreground">{name}</span>.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="grid gap-4 py-4">
                        <div className="grid grid-cols-4 items-center gap-4">
                            <label htmlFor="jobName" className="text-right text-sm font-medium">
                                Job Name
                            </label>
                            <input
                                id="jobName"
                                value={runJobName}
                                onChange={(e) => setRunJobName(e.target.value)}
                                className="col-span-3 flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                            />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button
                            variant="outline"
                            onClick={() => setIsRunDialogOpen(false)}
                            disabled={actioning}
                        >
                            Cancel
                        </Button>
                        <Button
                            onClick={async () => {
                                setActioning(true);
                                try {
                                    const res = await api.post<{ jobName: string }>(`/kube/cron-jobs/trigger?namespace=${namespace}&name=${name}`, {
                                        jobName: runJobName
                                    }, {
                                        headers: { "x-kube-context": context || "" }
                                    });
                                    toast.success(`Job ${res.jobName} created successfully`);
                                    setIsRunDialogOpen(false);
                                    // Optionally refresh to show the new job in related lists if any? 
                                    onUpdate?.();
                                } catch (err: any) {
                                    toast.error(err.message || "Trigger failed");
                                } finally {
                                    setActioning(false);
                                }
                            }}
                            disabled={actioning || !runJobName}
                        >
                            {actioning ? "Triggering..." : "Trigger"}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            <Dialog open={isScaleDialogOpen} onOpenChange={setIsScaleDialogOpen}>
                <DialogContent className="sm:max-w-[425px] bg-background border-border">
                    <DialogHeader>
                        <DialogTitle className="font-mono flex items-center gap-2">
                            <Maximize2 className="h-5 w-5 text-blue-500" />
                            Scale {kind}
                        </DialogTitle>
                        <DialogDescription>
                            Set the number of replicas for <span className="font-mono font-bold text-foreground">{name}</span>.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="flex items-center justify-center gap-6 py-10">
                        <Button
                            variant="outline"
                            size="icon"
                            className="h-12 w-12 rounded-full border-2 hover:bg-accent transition-colors"
                            onClick={() => setScaleReplicas(Math.max(0, scaleReplicas - 1))}
                            disabled={scaleReplicas <= 0}
                        >
                            <Minus className="h-6 w-6" />
                        </Button>
                        <div className="flex flex-col items-center gap-1">
                            <input
                                type="number"
                                value={scaleReplicas}
                                onChange={(e) => setScaleReplicas(parseInt(e.target.value) || 0)}
                                className="text-5xl font-black font-mono w-24 text-center bg-transparent border-none focus:ring-0 focus:outline-none [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
                            />
                            <span className="text-[10px] uppercase tracking-widest text-muted-foreground font-bold">Replicas</span>
                        </div>
                        <Button
                            variant="outline"
                            size="icon"
                            className="h-12 w-12 rounded-full border-2 hover:bg-accent transition-colors"
                            onClick={() => setScaleReplicas(scaleReplicas + 1)}
                        >
                            <Plus className="h-6 w-6" />
                        </Button>
                    </div>
                    <DialogFooter className="bg-muted/30 -mx-6 -mb-6 p-6 mt-4 border-t border-border">
                        <Button
                            variant="outline"
                            onClick={() => setIsScaleDialogOpen(false)}
                            disabled={actioning}
                            className="rounded-lg"
                        >
                            Cancel
                        </Button>
                        <Button
                            onClick={async () => {
                                setActioning(true);
                                try {
                                    await api.post(`/kube/resource/scale?namespace=${namespace}&name=${name}&kind=${kind}`, {
                                        replicas: scaleReplicas
                                    }, {
                                        headers: { "x-kube-context": context || "" }
                                    });
                                    toast.success(`${kind} ${name} scaled to ${scaleReplicas}`);
                                    setIsScaleDialogOpen(false);
                                    fetchDetails();
                                    onUpdate?.();
                                } catch (err: any) {
                                    toast.error(err.message || "Scale failed");
                                } finally {
                                    setActioning(false);
                                }
                            }}
                            disabled={actioning}
                            className="rounded-lg bg-blue-600 hover:bg-blue-700 text-white border-none shadow-md"
                        >
                            {actioning ? "Scaling..." : "Scale Replicas"}
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </Sheet >
    );
}
