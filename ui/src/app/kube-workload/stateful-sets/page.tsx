"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { RefreshCw, CheckCircle2, Database } from "lucide-react";
import { cn, formatAge } from "@/lib/utils";
import { NamespaceBadge } from "@/components/NamespaceBadge";
import { ResourceDetailsSheet } from "@/components/ResourceDetailsSheet";
import { api } from "@/lib/api";

interface StatefulSet {
    name: string;
    namespace: string;
    replicas: number;
    ready_replicas: number;
    current_replicas: number;
    updated_replicas: number;
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

function StatefulSetsContent() {
    const [searchParams] = useSearchParams();
    const selectedContext = searchParams.get("context");
    const selectedNamespace = searchParams.get("namespace");

    const [statefulsets, setStatefulSets] = useState<StatefulSet[]>([]);
    const [loading, setLoading] = useState(false);
    const searchQuery = searchParams.get("q") || "";
    const [selectedStatefulSet, setSelectedStatefulSet] = useState<StatefulSet | null>(null);

    useEffect(() => {
        if (selectedContext && selectedNamespace) {
            fetchStatefulSets();
        }
    }, [selectedContext, selectedNamespace]);

    const fetchStatefulSets = async () => {
        setLoading(true);
        setStatefulSets([]);
        try {
            const data = await api.get<any>(`/kube/stateful-sets?namespace=${selectedNamespace}`, {
                headers: { "x-kube-context": selectedContext || "" }
            });
            setStatefulSets(data.statefulsets || []);
        } catch (error) {
            console.error("Failed to fetch statefulsets:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleRefresh = () => {
        if (selectedContext && selectedNamespace) {
            fetchStatefulSets();
        }
    };

    const filteredStatefulSets = statefulsets.filter(ss =>
        ss.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        ss.namespace.toLowerCase().includes(searchQuery.toLowerCase())
    );

    return (
        <div className="flex flex-col items-center w-full min-h-screen bg-background/50 p-4 md:p-0">
            <div className="w-full max-w-6xl space-y-8">

                <Card className="border-none shadow-2xl shadow-black/5 bg-card/50 backdrop-blur-sm overflow-hidden rounded-3xl">
                    <CardHeader className="border-b bg-card/50 px-8 py-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <CardTitle className="text-lg font-bold flex items-center gap-3">
                                    <Database className="h-5 w-5 text-indigo-500" />
                                    {filteredStatefulSets.length} StatefulSets Found
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
                        ) : filteredStatefulSets.length === 0 ? (
                            <div className="text-center py-20 text-muted-foreground">
                                {selectedContext && selectedNamespace ? "No statefulsets found in selected namespace(s)." : "Select a context and namespace to view statefulsets."}
                            </div>
                        ) : (
                            <div className="grid grid-cols-1 gap-3">
                                {filteredStatefulSets.map((ss) => (
                                    <div
                                        key={`${ss.namespace}-${ss.name}`}
                                        className="group flex flex-col md:flex-row md:items-center justify-between p-5 bg-card/60 backdrop-blur-sm border rounded-2xl shadow-sm hover:shadow-lg hover:border-primary/30 transition-all duration-300 cursor-pointer"
                                        onClick={() => setSelectedStatefulSet(ss)}
                                    >
                                        <div className="flex items-center gap-4 min-w-0">
                                            <div className={cn(
                                                "p-3 rounded-xl",
                                                ss.ready_replicas === ss.replicas ? "bg-green-500/10" : "bg-amber-500/10"
                                            )}>
                                                {ss.ready_replicas === ss.replicas ? (
                                                    <CheckCircle2 className="h-6 w-6 text-green-500" />
                                                ) : (
                                                    <RefreshCw className="h-6 w-6 text-amber-500 animate-pulse" />
                                                )}
                                            </div>
                                            <div className="min-w-0">
                                                <p className="font-semibold text-base truncate">{ss.name}</p>
                                                <div className="flex items-center gap-2 mt-1 flex-wrap">
                                                    <span className="text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full bg-indigo-500/10 text-indigo-600">
                                                        {ss.ready_replicas}/{ss.replicas} ready
                                                    </span>
                                                </div>
                                            </div>
                                        </div>

                                        {/* Namespace */}
                                        <div
                                            className="flex flex-col items-end min-w-[120px] mt-4 md:mt-0"
                                            onClick={(e) => e.stopPropagation()}
                                        >
                                            <NamespaceBadge namespace={ss.namespace} />
                                        </div>

                                        {/* Age */}
                                        <div className="flex flex-col items-end min-w-[80px]">
                                            <span className="text-xs text-muted-foreground">{formatAge(ss.age)}</span>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </CardContent>
                </Card>
            </div>

            <ResourceDetailsSheet
                isOpen={!!selectedStatefulSet}
                onClose={() => setSelectedStatefulSet(null)}
                context={selectedContext || ""}
                namespace={selectedStatefulSet?.namespace || ""}
                name={selectedStatefulSet?.name || ""}
                kind="StatefulSet"
                onUpdate={handleRefresh}
            />
        </div>
    );
}

export default function StatefulSetsPage() {
    return (
        <Suspense fallback={<div className="flex items-center justify-center min-h-screen"><RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" /></div>}>
            <StatefulSetsContent />
        </Suspense>
    );
}
