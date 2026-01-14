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
import { Terminal as TerminalIcon, FileText, Ban, Trash2, CheckCircle2 } from "lucide-react";
import { LogViewerModal } from "@/components/LogViewerModal";
import { api } from "@/lib/api";
import { toast } from "sonner";

interface ResourceDetailsSheetProps {
    isOpen: boolean;
    onClose: () => void;
    context: string;
    namespace: string;
    name: string;
    kind: string;
}

interface EventSimple {
    type: string;
    reason: string;
    message: string;
    count: number;
    last_seen: string;
    age: string;
}

interface ResourceDetails {
    manifest: string;
    events: EventSimple[];
    raw: any;
}

export function ResourceDetailsSheet({
    isOpen,
    onClose,
    context,
    namespace,
    name,
    kind,
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

    const [scopes, setScopes] = useState<Record<string, string>>({});
    const [logResource, setLogResource] = useState<{
        name: string,
        namespace: string,
        selector?: string,
        pods: Array<{ name: string, status: string }>,
        containers: string[],
        initContainers: string[]
    } | null>(null);

    useEffect(() => {
        // Fetch scopes once
        api.get<any>("/kube/scopes")
            .then(data => setScopes(data.scopes || {}))
            .catch(err => console.error("Failed to fetch scopes:", err));
    }, []);

    const fetchDetails = () => {
        if (isOpen && context && name && kind && Object.keys(scopes).length > 0) {
            const scope = scopes[kind];
            const isClusterScoped = scope === "Cluster";

            if (isClusterScoped || namespace) {
                setLoading(true);
                setError("");

                api.get<ResourceDetails>(
                    `/kube/resource?context=${context}&namespace=${namespace}&name=${name}&kind=${kind}`
                )
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
    }, [isOpen, context, namespace, name, kind, scopes]);

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
                                                    const data = await api.get<any>(`/kube/pods?context=${context}&namespace=${namespace}&selector=${encodeURIComponent(selector)}`);
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
                                                        await api.post(`/kube/nodes/cordon?context=${context}&name=${name}`, {
                                                            unschedulable: !isUnschedulable
                                                        });
                                                        toast.success(`Node ${name} ${isUnschedulable ? "uncordoned" : "cordoned"}`);
                                                        fetchDetails();
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
                                                        const res = await api.post<any>(`/kube/nodes/drain?context=${context}&name=${name}`, {});
                                                        toast.success(`Drain started: ${res.evicted} pods evicted, ${res.skipped} skipped.`);
                                                        fetchDetails();
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
                                                await api.del(`/kube/resource?context=${context}&namespace=${namespace}&name=${name}&kind=${kind}`);
                                                toast.success(`${kind} ${name} deleted successfully`);
                                                onClose();
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
                            {/* Properties Section */}
                            <KubeProperties resource={details.raw} />

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
                                <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
                                    YAML Manifest
                                </h3>
                                <div className="relative rounded-md border border-border bg-card p-4 text-xs font-mono text-card-foreground overflow-auto max-h-[600px] shadow-sm">
                                    <pre className="whitespace-pre-wrap break-all">{details.manifest}</pre>
                                </div>
                            </div>
                        </div>
                    )}
                </div>
            </SheetContent>

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
        </Sheet >
    );
}
