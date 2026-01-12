"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Box, RefreshCw, CheckCircle2, XCircle, LayoutGrid, Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { cn, formatAge } from "@/lib/utils";
import { API_URL } from "@/lib/config";
import { NamespaceBadge } from "@/components/NamespaceBadge";

interface PodInfo {
    name: string;
    containers: string[];
    status: string;
    namespace: string;
    age: string;
}

function PodsContent() {
    const searchParams = useSearchParams();
    const selectedContext = searchParams.get("context") || "";
    const selectedNamespace = searchParams.get("namespace") || "";

    const [pods, setPods] = useState<PodInfo[]>([]);
    const [podsLoading, setPodsLoading] = useState(false);
    const [searchQuery, setSearchQuery] = useState("");

    const filteredPods = pods.filter(pod =>
        pod.name.toLowerCase().includes(searchQuery.toLowerCase())
    );

    useEffect(() => {
        if (!selectedContext || !selectedNamespace) {
            setPods([]);
            return;
        }

        fetchPods();
    }, [selectedContext, selectedNamespace]);

    const fetchPods = async () => {
        setPodsLoading(true);
        setPods([]);
        try {
            const res = await fetch(`${API_URL}/kube/pods?context=${selectedContext}&namespace=${selectedNamespace}`, { credentials: "include" });
            if (res.status === 401) {
                window.location.href = "/login";
                return;
            }
            if (res.ok) {
                const data = await res.json();
                setPods(data.pods || []);
            }
        } catch (error) {
            console.error("Failed to fetch pods:", error);
        } finally {
            setPodsLoading(false);
        }
    };

    const handleRefresh = () => {
        if (selectedContext && selectedNamespace) {
            fetchPods();
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
                                    <Box className="h-5 w-5 text-purple-500" />
                                    {pods.length} Pods Found
                                </CardTitle>
                                <CardDescription>
                                    {!selectedContext || !selectedNamespace ? "Select a namespace from the top bar to view pods" : null}
                                </CardDescription>
                            </div>
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={handleRefresh}
                                disabled={!selectedContext || !selectedNamespace || podsLoading}
                                className="rounded-xl"
                            >
                                <RefreshCw className={cn("h-4 w-4 mr-2", podsLoading && "animate-spin")} />
                                Refresh
                            </Button>
                        </div>
                        {pods.length > 0 && (
                            <div className="relative mt-4">
                                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                                <Input
                                    placeholder="Search pods by name..."
                                    value={searchQuery}
                                    onChange={(e) => setSearchQuery(e.target.value)}
                                    className="pl-10"
                                />
                            </div>
                        )}
                    </CardHeader>
                    <CardContent className="p-8">
                        {podsLoading ? (
                            <div className="flex items-center justify-center py-12">
                                <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
                            </div>
                        ) : !selectedContext ? (
                            <div className="text-center py-12">
                                <Box className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    Select a cluster from the top bar to view pods.
                                </p>
                            </div>
                        ) : !selectedNamespace ? (
                            <div className="text-center py-12">
                                <Box className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    Select a namespace from the top bar to view pods.
                                </p>
                            </div>
                        ) : pods.length === 0 ? (
                            <div className="text-center py-12">
                                <Box className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    No pods found in this namespace.
                                </p>
                            </div>
                        ) : (
                            <div className="grid grid-cols-1 gap-4">
                                {filteredPods.map(pod => (
                                    <div
                                        key={pod.name}
                                        className="p-6 bg-muted/30 rounded-2xl border border-muted/20 hover:bg-muted/50 transition-colors"
                                    >
                                        <div className="flex flex-col lg:flex-row lg:items-center gap-4 justify-between">
                                            {/* Pod Name and Status */}
                                            <div className="min-w-0 flex-1">
                                                <p className="font-semibold text-base truncate">{pod.name}</p>
                                                <div className="flex items-center gap-2 mt-1 flex-wrap">
                                                    <span className={cn(
                                                        "text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full",
                                                        (pod.status === "Running" || pod.status === "Succeeded") ? "bg-green-500/10 text-green-600" :
                                                            pod.status === "Pending" ? "bg-amber-500/10 text-amber-600" :
                                                                "bg-red-500/10 text-red-600"
                                                    )}>
                                                        {pod.status}
                                                    </span>
                                                </div>
                                            </div>

                                            {/* Actions & Containers */}
                                            <div className="flex flex-col gap-3 items-start lg:items-end min-w-[200px]">
                                                <div className="flex items-center gap-2">
                                                    <Button
                                                        variant="outline"
                                                        size="sm"
                                                        className="h-8 rounded-lg gap-2 text-xs font-semibold"
                                                        disabled={pod.status !== "Running" || pod.containers.length === 0}
                                                        onClick={() => {
                                                            const container = pod.containers.length > 0 ? pod.containers[0] : "";
                                                            window.open(`/exec?context=${selectedContext}&namespace=${pod.namespace}&pod=${pod.name}&container=${container}`, "_blank");
                                                        }}
                                                    >
                                                        <Box className="h-3.5 w-3.5" />
                                                        Terminal
                                                    </Button>
                                                </div>

                                                <div className="flex flex-col gap-1 items-end">
                                                    <div className="flex flex-wrap gap-2 justify-end">
                                                        {pod.containers.map(c => (
                                                            <span key={c} className="text-xs bg-muted border rounded-md px-2 py-1 font-mono">
                                                                {c}
                                                            </span>
                                                        ))}
                                                    </div>
                                                </div>
                                            </div>

                                            {/* Namespace */}
                                            <div className="flex flex-col items-end min-w-[120px]">
                                                <NamespaceBadge namespace={pod.namespace} />
                                            </div>

                                            {/* Age */}
                                            <div className="flex flex-col items-end min-w-[80px]">
                                                <span className="text-xs text-muted-foreground">{formatAge(pod.age)}</span>
                                            </div>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </CardContent>
                </Card>
            </div>
        </div >
    );
}

export default function PodsPage() {
    return (
        <Suspense fallback={<div className="flex items-center justify-center min-h-screen"><RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" /></div>}>
            <PodsContent />
        </Suspense>
    );
}
