"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { RefreshCw, CheckCircle2, Server } from "lucide-react";
import { cn, formatAge } from "@/lib/utils";
import { NamespaceBadge } from "@/components/NamespaceBadge";
import { ResourceDetailsSheet } from "@/components/ResourceDetailsSheet";
import { api } from "@/lib/api";

interface DaemonSet {
    name: string;
    namespace: string;
    desired_scheduled: number;
    current_scheduled: number;
    ready: number;
    up_to_date: number;
    available: number;
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

function DaemonSetsContent() {
    const [searchParams] = useSearchParams();
    const selectedContext = searchParams.get("context");
    const selectedNamespace = searchParams.get("namespace");

    const [daemonsets, setDaemonSets] = useState<DaemonSet[]>([]);
    const [loading, setLoading] = useState(false);
    const searchQuery = searchParams.get("q") || "";
    const [selectedDaemonSet, setSelectedDaemonSet] = useState<DaemonSet | null>(null);

    useEffect(() => {
        if (selectedContext && selectedNamespace) {
            fetchDaemonSets();
        }
    }, [selectedContext, selectedNamespace]);

    const fetchDaemonSets = async () => {
        setLoading(true);
        setDaemonSets([]);
        try {
            const data = await api.get<any>(`/kube/daemon-sets?namespace=${selectedNamespace}`, {
                headers: { "x-kube-context": selectedContext || "" },
            });
            setDaemonSets(data.daemonsets || []);
        } catch (error) {
            console.error("Failed to fetch daemonsets:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleRefresh = () => {
        if (selectedContext && selectedNamespace) {
            fetchDaemonSets();
        }
    };

    const filteredDaemonSets = daemonsets.filter(
        (ds) =>
            ds.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
            ds.namespace.toLowerCase().includes(searchQuery.toLowerCase())
    );

    return (
        <div className="flex flex-col items-center w-full min-h-screen bg-background/50 p-4 md:p-0">
            <div className="w-full max-w-6xl space-y-8">
                <Card className="border-none shadow-2xl shadow-black/5 bg-card/50 backdrop-blur-sm overflow-hidden rounded-3xl">
                    <CardHeader className="border-b bg-card/50 px-8 py-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <CardTitle className="text-lg font-bold flex items-center gap-3">
                                    <Server className="h-5 w-5 text-cyan-500" />
                                    {filteredDaemonSets.length} DaemonSets Found
                                </CardTitle>
                            </div>
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={handleRefresh}
                                disabled={loading || !selectedContext || !selectedNamespace}
                                className="h-10 px-4 rounded-xl"
                            >
                                <RefreshCw className={cn("h-4 w-4 mr-2", loading && "animate-spin")} />
                                Refresh
                            </Button>
                        </div>
                    </CardHeader>
                    <CardContent className="p-6">
                        {loading ? (
                            <div className="flex items-center justify-center py-20">
                                <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
                            </div>
                        ) : filteredDaemonSets.length === 0 ? (
                            <div className="text-center py-20 text-muted-foreground">
                                {selectedContext && selectedNamespace
                                    ? "No daemonsets found in selected namespace(s)."
                                    : "Select a context and namespace to view daemonsets."}
                            </div>
                        ) : (
                            <div className="grid grid-cols-1 gap-3">
                                {filteredDaemonSets.map((ds) => (
                                    <div
                                        key={`${ds.namespace}-${ds.name}`}
                                        className="group flex flex-col md:flex-row md:items-center justify-between p-5 bg-card/60 backdrop-blur-sm border rounded-2xl shadow-sm hover:shadow-lg hover:border-primary/30 transition-all duration-300 cursor-pointer"
                                        onClick={() => setSelectedDaemonSet(ds)}
                                    >
                                        <div className="flex items-center gap-4 min-w-0">
                                            <div
                                                className={cn(
                                                    "p-3 rounded-xl",
                                                    ds.ready === ds.desired_scheduled
                                                        ? "bg-green-500/10"
                                                        : "bg-amber-500/10"
                                                )}
                                            >
                                                {ds.ready === ds.desired_scheduled ? (
                                                    <CheckCircle2 className="h-6 w-6 text-green-500" />
                                                ) : (
                                                    <RefreshCw className="h-6 w-6 text-amber-500 animate-pulse" />
                                                )}
                                            </div>
                                            <div className="min-w-0">
                                                <p className="font-semibold text-base truncate">{ds.name}</p>
                                                <div className="flex items-center gap-2 mt-1 flex-wrap">
                                                    <span className="text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full bg-cyan-500/10 text-cyan-600">
                                                        {ds.ready}/{ds.desired_scheduled} ready
                                                    </span>
                                                </div>
                                            </div>
                                        </div>

                                        {/* Namespace */}
                                        <div
                                            className="flex flex-col items-end min-w-[120px] mt-4 md:mt-0"
                                            onClick={(e) => e.stopPropagation()}
                                        >
                                            <NamespaceBadge namespace={ds.namespace} />
                                        </div>

                                        {/* Age */}
                                        <div className="flex flex-col items-end min-w-[80px]">
                                            <span className="text-xs text-muted-foreground">{formatAge(ds.age)}</span>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </CardContent>
                </Card>
            </div>

            <ResourceDetailsSheet
                isOpen={!!selectedDaemonSet}
                onClose={() => setSelectedDaemonSet(null)}
                context={selectedContext || ""}
                namespace={selectedDaemonSet?.namespace || ""}
                name={selectedDaemonSet?.name || ""}
                kind="DaemonSet"
                onUpdate={handleRefresh}
            />
        </div>
    );
}

export default function DaemonSetsPage() {
    return (
        <Suspense
            fallback={
                <div className="flex items-center justify-center min-h-screen">
                    <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
                </div>
            }
        >
            <DaemonSetsContent />
        </Suspense>
    );
}
