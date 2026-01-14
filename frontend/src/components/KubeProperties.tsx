"use client";

import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import { formatAge, toYaml } from "@/lib/utils";
import { ChevronDown, ChevronRight, Tags, StickyNote, Anchor, MapPin, Share2 } from "lucide-react";
import { cn } from "@/lib/utils";

interface KubePropertiesProps {
    resource: any;
}

export function KubeProperties({ resource }: KubePropertiesProps) {
    const [showLabels, setShowLabels] = useState(false);
    const [showAnnotations, setShowAnnotations] = useState(false);
    const [showTaints, setShowTaints] = useState(false);
    const [showConditions, setShowConditions] = useState(false);
    const [showTolerations, setShowTolerations] = useState(false);
    const [showNodeSelector, setShowNodeSelector] = useState(false);
    const [showNodeAffinity, setShowNodeAffinity] = useState(false);
    const [showPodAffinity, setShowPodAffinity] = useState(false);
    const [showPodAntiAffinity, setShowPodAntiAffinity] = useState(false);

    if (!resource || !resource.metadata) return null;

    const { metadata } = resource;
    const {
        name,
        namespace,
        creationTimestamp,
        labels,
        annotations
    } = metadata;

    const getResourceStatus = (res: any) => {
        if (!res.status) return null;

        const { status, kind } = res;

        // Specific handling for common kinds
        if (kind === "Pod") return status.phase;
        if (kind === "Node") {
            const readyCondition = status.conditions?.find((c: any) => c.type === "Ready");
            return readyCondition?.status === "True" ? "Ready" : "NotReady";
        }
        if (kind === "Job") {
            if (status.succeeded > 0) return "Succeeded";
            if (status.failed > 0) return "Failed";
            if (status.active > 0) return "Active";
        }
        if (kind === "CronJob") {
            return status.lastScheduleTime ? "Scheduled" : "Never Scheduled";
        }

        // Generic fallback for workloads with replicas
        if (status.replicas !== undefined && status.readyReplicas !== undefined) {
            return `${status.readyReplicas}/${status.replicas} Ready`;
        }

        // Check conditions for a generic "Ready" state
        const readyCond = status.conditions?.find((c: any) => c.type === "Ready" || c.type === "Available");
        if (readyCond) {
            return readyCond.status === "True" ? "Ready" : "NotReady";
        }

        // Phase is a common field for many resources
        if (status.phase) return status.phase;

        return null;
    };

    const statusValue = getResourceStatus(resource);

    const properties = [
        { label: "Name", value: name },
        { label: "Namespace", value: namespace },
        {
            label: "Created",
            value: creationTimestamp ? `${new Date(creationTimestamp).toLocaleString()} (${formatAge(creationTimestamp)} ago)` : null
        },
        { label: "Status", value: statusValue },
    ];

    const labelCount = labels ? Object.keys(labels).length : 0;
    const annotationCount = annotations ? Object.keys(annotations).length : 0;
    const taints = resource.spec?.taints || [];
    const conditions = resource.status?.conditions || [];
    const tolerations = resource.spec?.tolerations || [];
    const nodeSelector = resource.spec?.nodeSelector || {};
    const affinity = resource.spec?.affinity || {};

    const hasAffinity = affinity.nodeAffinity || affinity.podAffinity || affinity.podAntiAffinity;

    const getNodeAffinityCount = (na: any) => {
        let count = 0;
        if (na?.requiredDuringSchedulingIgnoredDuringExecution?.nodeSelectorTerms) {
            count += na.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms.length;
        }
        if (na?.preferredDuringSchedulingIgnoredDuringExecution) {
            count += na.preferredDuringSchedulingIgnoredDuringExecution.length;
        }
        return count;
    };

    const getPodAffinityCount = (pa: any) => {
        let count = 0;
        if (pa?.requiredDuringSchedulingIgnoredDuringExecution) {
            count += pa.requiredDuringSchedulingIgnoredDuringExecution.length;
        }
        if (pa?.preferredDuringSchedulingIgnoredDuringExecution) {
            count += pa.preferredDuringSchedulingIgnoredDuringExecution.length;
        }
        return count;
    };

    const nodeAffinityCount = getNodeAffinityCount(affinity.nodeAffinity);
    const podAffinityCount = getPodAffinityCount(affinity.podAffinity);
    const podAntiAffinityCount = getPodAffinityCount(affinity.podAntiAffinity);

    const getConditionBadgeStyle = (status: string, type: string) => {
        if (status === "True") {
            return type === "Ready" || type === "Available"
                ? "bg-emerald-50 text-emerald-700 border-emerald-200"
                : "bg-blue-50 text-blue-700 border-blue-200";
        }
        if (status === "False") {
            return type === "Ready" || type === "Available"
                ? "bg-red-50 text-red-700 border-red-200"
                : "bg-amber-50 text-amber-700 border-amber-200";
        }
        return "bg-muted text-muted-foreground border-border";
    };

    return (
        <div className="space-y-4">
            <h3 className="text-sm font-bold text-foreground uppercase tracking-wider flex items-center gap-2">
                Properties
            </h3>
            <div className="rounded-xl border border-border bg-card p-4 space-y-4 shadow-sm font-sans">
                <div className="space-y-3">
                    {properties.map((prop) => (
                        prop.value && (
                            <div key={prop.label} className="grid grid-cols-3 gap-2 items-start">
                                <span className="text-muted-foreground text-xs font-semibold uppercase tracking-tight">{prop.label}</span>
                                {prop.label === "Status" ? (
                                    <div className="col-span-2">
                                        <Badge
                                            variant="outline"
                                            className={cn(
                                                "text-[10px] font-bold px-2 py-0 h-auto",
                                                prop.value === "Running" || prop.value === "Ready" || prop.value === "Succeeded"
                                                    ? "bg-emerald-50 text-emerald-700 border-emerald-200"
                                                    : prop.value === "Pending" || prop.value === "Active"
                                                        ? "bg-amber-50 text-amber-700 border-amber-200"
                                                        : prop.value === "Failed" || prop.value === "NotReady"
                                                            ? "bg-red-50 text-red-700 border-red-200"
                                                            : "bg-muted text-muted-foreground border-border"
                                            )}
                                        >
                                            {prop.value}
                                        </Badge>
                                    </div>
                                ) : (
                                    <span className="text-foreground col-span-2 break-all font-bold text-xs">{prop.value}</span>
                                )}
                            </div>
                        )
                    ))}
                </div>

                {labelCount > 0 && (
                    <div className="pt-3 border-t border-border/50">
                        <button
                            onClick={() => setShowLabels(!showLabels)}
                            className="flex items-center justify-between w-full text-left group"
                        >
                            <div className="flex items-center gap-2">
                                <Tags className="h-3.5 w-3.5 text-muted-foreground/60" />
                                <span className="text-muted-foreground text-xs font-semibold uppercase tracking-tight">Labels</span>
                                <Badge variant="secondary" className="bg-muted text-muted-foreground border-none text-[10px] font-bold px-1.5 h-4">
                                    {labelCount}
                                </Badge>
                            </div>
                            {showLabels ? <ChevronDown className="h-3.5 w-3.5 text-muted-foreground/60" /> : <ChevronRight className="h-3.5 w-3.5 text-muted-foreground/60" />}
                        </button>

                        {showLabels && (
                            <div className="mt-3 flex flex-wrap gap-1.5 animate-in fade-in slide-in-from-top-1 duration-200">
                                {Object.entries(labels).map(([k, v]) => (
                                    <Badge key={k} variant="outline" className="text-[10px] font-bold border-border bg-muted/50 text-foreground/80 py-0.5 px-2 h-auto whitespace-normal break-all">
                                        {k}: {String(v)}
                                    </Badge>
                                ))}
                            </div>
                        )}
                    </div>
                )}

                {annotationCount > 0 && (
                    <div className="pt-3 border-t border-border/50">
                        <button
                            onClick={() => setShowAnnotations(!showAnnotations)}
                            className="flex items-center justify-between w-full text-left group"
                        >
                            <div className="flex items-center gap-2">
                                <StickyNote className="h-3.5 w-3.5 text-muted-foreground/60" />
                                <span className="text-muted-foreground text-xs font-semibold uppercase tracking-tight">Annotations</span>
                                <Badge variant="secondary" className="bg-muted text-muted-foreground border-none text-[10px] font-bold px-1.5 h-4">
                                    {annotationCount}
                                </Badge>
                            </div>
                            {showAnnotations ? <ChevronDown className="h-3.5 w-3.5 text-muted-foreground/60" /> : <ChevronRight className="h-3.5 w-3.5 text-muted-foreground/60" />}
                        </button>

                        {showAnnotations && (
                            <div className="mt-3 flex flex-wrap gap-1.5 animate-in fade-in slide-in-from-top-1 duration-200">
                                {Object.entries(annotations).map(([k, v]) => (
                                    <Badge key={k} variant="outline" className="text-[10px] font-bold border-border bg-muted/50 text-foreground/80 py-0.5 px-2 max-w-full h-auto whitespace-normal break-all">
                                        {k}: {String(v)}
                                    </Badge>
                                ))}
                            </div>
                        )}
                    </div>
                )}

                {taints.length > 0 && (
                    <div className="pt-3 border-t border-border/50">
                        <button
                            onClick={() => setShowTaints(!showTaints)}
                            className="flex items-center justify-between w-full text-left group"
                        >
                            <div className="flex items-center gap-2">
                                <Tags className="h-3.5 w-3.5 text-muted-foreground/60 rotate-90" />
                                <span className="text-muted-foreground text-xs font-semibold uppercase tracking-tight">Taints</span>
                                <Badge variant="secondary" className="bg-muted text-muted-foreground border-none text-[10px] font-bold px-1.5 h-4">
                                    {taints.length}
                                </Badge>
                            </div>
                            {showTaints ? <ChevronDown className="h-3.5 w-3.5 text-muted-foreground/60" /> : <ChevronRight className="h-3.5 w-3.5 text-muted-foreground/60" />}
                        </button>

                        {showTaints && (
                            <div className="mt-3 flex flex-wrap gap-1.5 animate-in fade-in slide-in-from-top-1 duration-200">
                                {taints.map((taint: any, idx: number) => (
                                    <Badge key={idx} variant="outline" className="text-[10px] font-bold border-amber-200 bg-amber-50 text-amber-700 py-0.5 px-2 h-auto whitespace-normal break-all">
                                        {taint.key}{taint.value ? `=${taint.value}` : ""}:{taint.effect}
                                    </Badge>
                                ))}
                            </div>
                        )}
                    </div>
                )}

                {conditions.length > 0 && (
                    <div className="pt-3 border-t border-border/50">
                        <button
                            onClick={() => setShowConditions(!showConditions)}
                            className="flex items-center justify-between w-full text-left group"
                        >
                            <div className="flex items-center gap-2">
                                <div className="h-3.5 w-3.5 flex items-center justify-center">
                                    <div className="h-1.5 w-1.5 rounded-full bg-muted-foreground/40" />
                                </div>
                                <span className="text-muted-foreground text-xs font-semibold uppercase tracking-tight">Conditions</span>
                                <Badge variant="secondary" className="bg-muted text-muted-foreground border-none text-[10px] font-bold px-1.5 h-4">
                                    {conditions.length}
                                </Badge>
                            </div>
                            {showConditions ? <ChevronDown className="h-3.5 w-3.5 text-muted-foreground/60" /> : <ChevronRight className="h-3.5 w-3.5 text-muted-foreground/60" />}
                        </button>

                        {showConditions && (
                            <div className="mt-3 space-y-2 animate-in fade-in slide-in-from-top-1 duration-200">
                                {conditions.map((condition: any, idx: number) => (
                                    <div key={idx} className="flex flex-col gap-1 p-2 rounded-lg border border-border/50 bg-muted/30">
                                        <div className="flex items-center justify-between">
                                            <span className="text-[10px] font-bold text-foreground">{condition.type}</span>
                                            <Badge
                                                variant="outline"
                                                className={cn(
                                                    "text-[9px] font-bold px-1.5 py-0 h-3.5",
                                                    getConditionBadgeStyle(condition.status, condition.type)
                                                )}
                                            >
                                                {condition.status}
                                            </Badge>
                                        </div>
                                        {condition.reason && (
                                            <span className="text-[9px] text-muted-foreground font-medium">Reason: {condition.reason}</span>
                                        )}
                                        {condition.message && (
                                            <span className="text-[9px] text-foreground/80 leading-tight">{condition.message}</span>
                                        )}
                                        {condition.lastTransitionTime && (
                                            <span className="text-[9px] text-muted-foreground/60 font-medium mt-0.5">
                                                Last Transition: {new Date(condition.lastTransitionTime).toLocaleString()} ({formatAge(condition.lastTransitionTime)} ago)
                                            </span>
                                        )}
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                )}

                {tolerations.length > 0 && (
                    <div className="pt-3 border-t border-border/50">
                        <button
                            onClick={() => setShowTolerations(!showTolerations)}
                            className="flex items-center justify-between w-full text-left group"
                        >
                            <div className="flex items-center gap-2">
                                <Anchor className="h-3.5 w-3.5 text-muted-foreground/60" />
                                <span className="text-muted-foreground text-xs font-semibold uppercase tracking-tight">Tolerations</span>
                                <Badge variant="secondary" className="bg-muted text-muted-foreground border-none text-[10px] font-bold px-1.5 h-4">
                                    {tolerations.length}
                                </Badge>
                            </div>
                            {showTolerations ? <ChevronDown className="h-3.5 w-3.5 text-muted-foreground/60" /> : <ChevronRight className="h-3.5 w-3.5 text-muted-foreground/60" />}
                        </button>

                        {showTolerations && (
                            <div className="mt-3 flex flex-wrap gap-1.5 animate-in fade-in slide-in-from-top-1 duration-200">
                                {tolerations.map((t: any, idx: number) => (
                                    <Badge key={idx} variant="outline" className="text-[10px] font-bold border-border bg-muted/50 text-foreground/80 py-0.5 px-2 h-auto whitespace-normal break-all">
                                        {t.key}{t.operator === "Exists" ? " (Exists)" : t.value ? `=${t.value}` : ""}{t.effect ? `:${t.effect}` : ""}
                                    </Badge>
                                ))}
                            </div>
                        )}
                    </div>
                )}

                {Object.keys(nodeSelector).length > 0 && (
                    <div className="pt-3 border-t border-border/50">
                        <button
                            onClick={() => setShowNodeSelector(!showNodeSelector)}
                            className="flex items-center justify-between w-full text-left group"
                        >
                            <div className="flex items-center gap-2">
                                <MapPin className="h-3.5 w-3.5 text-muted-foreground/60" />
                                <span className="text-muted-foreground text-xs font-semibold uppercase tracking-tight">Node Selector</span>
                                <Badge variant="secondary" className="bg-muted text-muted-foreground border-none text-[10px] font-bold px-1.5 h-4">
                                    {Object.keys(nodeSelector).length}
                                </Badge>
                            </div>
                            {showNodeSelector ? <ChevronDown className="h-3.5 w-3.5 text-muted-foreground/60" /> : <ChevronRight className="h-3.5 w-3.5 text-muted-foreground/60" />}
                        </button>

                        {showNodeSelector && (
                            <div className="mt-3 flex flex-wrap gap-1.5 animate-in fade-in slide-in-from-top-1 duration-200">
                                {Object.entries(nodeSelector).map(([k, v]) => (
                                    <Badge key={k} variant="outline" className="text-[10px] font-bold border-border bg-muted/50 text-foreground/80 py-0.5 px-2 h-auto whitespace-normal break-all">
                                        {k}: {String(v)}
                                    </Badge>
                                ))}
                            </div>
                        )}
                    </div>
                )}

                {affinity.nodeAffinity && (
                    <div className="pt-3 border-t border-border/50">
                        <button
                            onClick={() => setShowNodeAffinity(!showNodeAffinity)}
                            className="flex items-center justify-between w-full text-left group"
                        >
                            <div className="flex items-center gap-2">
                                <Share2 className="h-3.5 w-3.5 text-muted-foreground/60" />
                                <span className="text-muted-foreground text-xs font-semibold uppercase tracking-tight">Node Affinity</span>
                                <Badge variant="secondary" className="bg-muted text-muted-foreground border-none text-[10px] font-bold px-1.5 h-4">
                                    {nodeAffinityCount}
                                </Badge>
                            </div>
                            {showNodeAffinity ? <ChevronDown className="h-3.5 w-3.5 text-muted-foreground/60" /> : <ChevronRight className="h-3.5 w-3.5 text-muted-foreground/60" />}
                        </button>

                        {showNodeAffinity && (
                            <div className="mt-3 animate-in fade-in slide-in-from-top-1 duration-200">
                                <div className="p-2 rounded-lg border border-border/50 bg-muted/30 text-[10px] text-foreground/80 whitespace-pre-wrap font-mono overflow-auto max-h-40">
                                    {toYaml(affinity.nodeAffinity).trim()}
                                </div>
                            </div>
                        )}
                    </div>
                )}

                {affinity.podAffinity && (
                    <div className="pt-3 border-t border-border/50">
                        <button
                            onClick={() => setShowPodAffinity(!showPodAffinity)}
                            className="flex items-center justify-between w-full text-left group"
                        >
                            <div className="flex items-center gap-2">
                                <Share2 className="h-3.5 w-3.5 text-muted-foreground/60" />
                                <span className="text-muted-foreground text-xs font-semibold uppercase tracking-tight">Pod Affinity</span>
                                <Badge variant="secondary" className="bg-muted text-muted-foreground border-none text-[10px] font-bold px-1.5 h-4">
                                    {podAffinityCount}
                                </Badge>
                            </div>
                            {showPodAffinity ? <ChevronDown className="h-3.5 w-3.5 text-muted-foreground/60" /> : <ChevronRight className="h-3.5 w-3.5 text-muted-foreground/60" />}
                        </button>

                        {showPodAffinity && (
                            <div className="mt-3 animate-in fade-in slide-in-from-top-1 duration-200">
                                <div className="p-2 rounded-lg border border-border/50 bg-muted/30 text-[10px] text-foreground/80 whitespace-pre-wrap font-mono overflow-auto max-h-40">
                                    {toYaml(affinity.podAffinity).trim()}
                                </div>
                            </div>
                        )}
                    </div>
                )}

                {affinity.podAntiAffinity && (
                    <div className="pt-3 border-t border-border/50">
                        <button
                            onClick={() => setShowPodAntiAffinity(!showPodAntiAffinity)}
                            className="flex items-center justify-between w-full text-left group"
                        >
                            <div className="flex items-center gap-2">
                                <Share2 className="h-3.5 w-3.5 text-muted-foreground/60" />
                                <span className="text-muted-foreground text-xs font-semibold uppercase tracking-tight">Pod Anti-Affinity</span>
                                <Badge variant="secondary" className="bg-muted text-muted-foreground border-none text-[10px] font-bold px-1.5 h-4">
                                    {podAntiAffinityCount}
                                </Badge>
                            </div>
                            {showPodAntiAffinity ? <ChevronDown className="h-3.5 w-3.5 text-muted-foreground/60" /> : <ChevronRight className="h-3.5 w-3.5 text-muted-foreground/60" />}
                        </button>

                        {showPodAntiAffinity && (
                            <div className="mt-3 animate-in fade-in slide-in-from-top-1 duration-200">
                                <div className="p-2 rounded-lg border border-border/50 bg-muted/30 text-[10px] text-foreground/80 whitespace-pre-wrap font-mono overflow-auto max-h-40">
                                    {toYaml(affinity.podAntiAffinity).trim()}
                                </div>
                            </div>
                        )}
                    </div>
                )}
            </div>
        </div>
    );
}
