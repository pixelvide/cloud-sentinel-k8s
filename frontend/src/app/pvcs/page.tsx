"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Database, RefreshCw, HardDrive } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn, formatAge } from "@/lib/utils";
import { API_URL } from "@/lib/config";
import { NamespaceBadge } from "@/components/NamespaceBadge";
import { ResourceDetailsSheet } from "@/components/ResourceDetailsSheet";

interface PVCInfo {
    name: string;
    namespace: string;
    status: string;
    volume: string;
    capacity: string;
    access_modes: string;
    storage_class: string;
    age: string;
}

function PVCContent() {
    const searchParams = useSearchParams();
    const selectedContext = searchParams.get("context") || "";
    const selectedNamespace = searchParams.get("namespace") || "";

    const [pvcs, setPvcs] = useState<PVCInfo[]>([]);
    const [loading, setLoading] = useState(false);
    const searchQuery = searchParams.get("q") || "";
    const [selectedItem, setSelectedItem] = useState<PVCInfo | null>(null);

    const filteredItems = pvcs.filter(item =>
        item.name.toLowerCase().includes(searchQuery.toLowerCase())
    );

    useEffect(() => {
        if (!selectedContext || !selectedNamespace) {
            setPvcs([]);
            return;
        }
        fetchPVCs();
    }, [selectedContext, selectedNamespace]);

    const fetchPVCs = async () => {
        setLoading(true);
        setPvcs([]);
        try {
            const res = await fetch(`${API_URL}/kube/pvcs?context=${selectedContext}&namespace=${selectedNamespace}`, { credentials: "include" });
            if (res.status === 401) {
                window.location.href = "/login";
                return;
            }
            if (res.ok) {
                const data = await res.json();
                setPvcs(data.pvcs || []);
            }
        } catch (error) {
            console.error("Failed to fetch PVCs:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleRefresh = () => {
        if (selectedContext && selectedNamespace) {
            fetchPVCs();
        }
    };

    const getStatusColor = (status: string) => {
        switch (status.toLowerCase()) {
            case 'bound': return 'bg-green-500/10 text-green-600';
            case 'pending': return 'bg-yellow-500/10 text-yellow-600';
            case 'lost': return 'bg-red-500/10 text-red-600';
            default: return 'bg-muted text-muted-foreground';
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
                                    <Database className="h-5 w-5 text-indigo-500" />
                                    {pvcs.length} PVCs Found
                                </CardTitle>
                                <CardDescription>
                                    {!selectedContext || !selectedNamespace ? "Select a namespace from the top bar to view PVCs" : "Manage storage requests"}
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
                                <Database className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    Select a cluster from the top bar to view PVCs.
                                </p>
                            </div>
                        ) : !selectedNamespace ? (
                            <div className="text-center py-12">
                                <Database className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    Select a namespace from the top bar to view PVCs.
                                </p>
                            </div>
                        ) : pvcs.length === 0 ? (
                            <div className="text-center py-12">
                                <Database className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    No PVCs found in this namespace.
                                </p>
                            </div>
                        ) : (
                            <div className="grid grid-cols-1 gap-4">
                                {filteredItems.map(item => (
                                    <div
                                        key={item.name}
                                        className="p-6 bg-muted/30 rounded-2xl border border-muted/20 hover:bg-muted/50 transition-colors cursor-pointer"
                                        onClick={() => setSelectedItem(item)}
                                    >
                                        <div className="flex flex-col lg:flex-row lg:items-center gap-4 justify-between">
                                            <div className="min-w-0 flex-1">
                                                <p className="font-semibold text-base truncate">{item.name}</p>
                                                <div className="flex items-center gap-2 mt-1 flex-wrap">
                                                    <span className={cn("text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full", getStatusColor(item.status))}>
                                                        {item.status}
                                                    </span>
                                                    <span className="text-xs text-muted-foreground flex items-center gap-1">
                                                        <HardDrive className="h-3 w-3" />
                                                        {item.capacity}
                                                    </span>
                                                    <span className="text-xs text-muted-foreground">
                                                        SC: {item.storage_class}
                                                    </span>
                                                </div>
                                            </div>

                                            <div className="flex flex-col gap-1 items-start lg:items-end min-w-[200px]">
                                                <p className="text-xs font-mono text-muted-foreground truncate w-full lg:text-right">
                                                    {item.volume}
                                                </p>
                                                <p className="text-[10px] text-muted-foreground uppercase tracking-widest">
                                                    {item.access_modes}
                                                </p>
                                            </div>

                                            <div
                                                className="flex flex-col items-end min-w-[120px]"
                                                onClick={(e) => e.stopPropagation()}
                                            >
                                                <NamespaceBadge namespace={item.namespace} />
                                            </div>

                                            <div className="flex flex-col items-end min-w-[80px]">
                                                <span className="text-xs text-muted-foreground">{formatAge(item.age)}</span>
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
                isOpen={!!selectedItem}
                onClose={() => setSelectedItem(null)}
                context={selectedContext}
                namespace={selectedItem?.namespace || ""}
                name={selectedItem?.name || ""}
                kind="PersistentVolumeClaim"
            />
        </div >
    );
}

export default function PVCPage() {
    return (
        <Suspense fallback={<div className="flex items-center justify-center min-h-screen"><RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" /></div>}>
            <PVCContent />
        </Suspense>
    );
}
