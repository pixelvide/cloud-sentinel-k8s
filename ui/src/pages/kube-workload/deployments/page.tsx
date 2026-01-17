"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Layers, RefreshCw, CheckCircle2, XCircle, Send } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn, formatAge } from "@/lib/utils";
import { NamespaceBadge } from "@/components/NamespaceBadge";
import { ResourceDetailsSheet } from "@/components/ResourceDetailsSheet";
import { api } from "@/lib/api";

interface DeploymentInfo {
    name: string;
    namespace: string;
    replicas: number;
    ready_replicas: number;
    available_replicas: number;
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

function DeploymentsContent() {
    const [searchParams] = useSearchParams();
    const selectedContext = searchParams.get("context") || "";
    const selectedNamespace = searchParams.get("namespace") || "";

    const [deployments, setDeployments] = useState<DeploymentInfo[]>([]);
    const [loading, setLoading] = useState(false);
    const searchQuery = searchParams.get("q") || "";
    const [selectedDeployment, setSelectedDeployment] = useState<DeploymentInfo | null>(null);

    const filteredDeployments = deployments.filter((d) => d.name.toLowerCase().includes(searchQuery.toLowerCase()));

    useEffect(() => {
        if (!selectedContext || !selectedNamespace) {
            setDeployments([]);
            return;
        }
        fetchDeployments();
    }, [selectedContext, selectedNamespace]);

    const fetchDeployments = async () => {
        setLoading(true);
        setDeployments([]);
        try {
            const data = await api.get<any>(`/kube/deployments?namespace=${selectedNamespace}`, {
                headers: { "x-kube-context": selectedContext || "" },
            });
            setDeployments(data.deployments || []);
        } catch (error) {
            console.error("Failed to fetch deployments:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleRefresh = () => {
        if (selectedContext && selectedNamespace) {
            fetchDeployments();
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
                                    <Layers className="h-5 w-5 text-blue-500" />
                                    {deployments.length} Deployments Found
                                </CardTitle>
                                <CardDescription>
                                    {!selectedContext || !selectedNamespace
                                        ? "Select a namespace from the top bar to view deployments"
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
                                <Layers className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    Select a cluster from the top bar to view deployments.
                                </p>
                            </div>
                        ) : !selectedNamespace ? (
                            <div className="text-center py-12">
                                <Layers className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    Select a namespace from the top bar to view deployments.
                                </p>
                            </div>
                        ) : deployments.length === 0 ? (
                            <div className="text-center py-12">
                                <Layers className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">No deployments found in this namespace.</p>
                            </div>
                        ) : (
                            <div className="grid grid-cols-1 gap-4">
                                {filteredDeployments.map((d) => (
                                    <div
                                        key={d.name + d.namespace}
                                        className="p-6 bg-muted/30 rounded-2xl border border-muted/20 hover:bg-muted/50 transition-colors cursor-pointer"
                                        onClick={() => setSelectedDeployment(d)}
                                    >
                                        <div className="flex flex-col lg:flex-row lg:items-center gap-4 justify-between">
                                            <div className="flex items-center gap-4 min-w-0">
                                                <div
                                                    className={cn(
                                                        "p-3 rounded-xl",
                                                        d.ready_replicas === d.replicas
                                                            ? "bg-green-500/10"
                                                            : "bg-amber-500/10"
                                                    )}
                                                >
                                                    {d.ready_replicas === d.replicas ? (
                                                        <CheckCircle2 className="h-6 w-6 text-green-500" />
                                                    ) : (
                                                        <RefreshCw className="h-6 w-6 text-amber-500 animate-pulse" />
                                                    )}
                                                </div>
                                                <div className="min-w-0">
                                                    <p className="font-semibold text-base truncate">{d.name}</p>
                                                    <div className="flex items-center gap-2 mt-1 flex-wrap">
                                                        <span className="text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full bg-blue-500/10 text-blue-600">
                                                            {d.ready_replicas}/{d.replicas} ready
                                                        </span>
                                                    </div>
                                                </div>
                                            </div>

                                            {/* Namespace */}
                                            <div
                                                className="flex flex-col items-end min-w-[120px]"
                                                onClick={(e) => e.stopPropagation()}
                                            >
                                                <NamespaceBadge namespace={d.namespace} />
                                            </div>

                                            {/* Age */}
                                            <div className="flex flex-col items-end min-w-[80px]">
                                                <span className="text-xs text-muted-foreground">
                                                    {formatAge(d.age)}
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
                isOpen={!!selectedDeployment}
                onClose={() => setSelectedDeployment(null)}
                context={selectedContext}
                namespace={selectedDeployment?.namespace || ""}
                name={selectedDeployment?.name || ""}
                kind="Deployment"
                onUpdate={handleRefresh}
            />
        </div>
    );
}

export default function DeploymentsPage() {
    return (
        <Suspense
            fallback={
                <div className="flex items-center justify-center min-h-screen">
                    <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
                </div>
            }
        >
            <DeploymentsContent />
        </Suspense>
    );
}
