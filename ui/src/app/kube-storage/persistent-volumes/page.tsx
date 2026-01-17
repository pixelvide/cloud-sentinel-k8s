"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { HardDrive, RefreshCw, Link as LinkIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn, formatAge } from "@/lib/utils";
import { api } from "@/lib/api";
import { ResourceDetailsSheet } from "@/components/ResourceDetailsSheet";

interface PVInfo {
    name: string;
    capacity: string;
    access_modes: string;
    reclaim_policy: string;
    status: string;
    claim: string;
    storage_class: string;
    reason: string;
    age: string;
}

function PVContent() {
    const [searchParams] = useSearchParams();
    const selectedContext = searchParams.get("context") || "";

    const [pvs, setPvs] = useState<PVInfo[]>([]);
    const [loading, setLoading] = useState(false);
    const searchQuery = searchParams.get("q") || "";
    const [selectedItem, setSelectedItem] = useState<PVInfo | null>(null);

    const filteredItems = pvs.filter(item =>
        item.name.toLowerCase().includes(searchQuery.toLowerCase())
    );

    useEffect(() => {
        if (!selectedContext) {
            setPvs([]);
            return;
        }
        fetchPVs();
    }, [selectedContext]);

    const fetchPVs = async () => {
        setLoading(true);
        setPvs([]);
        try {
            const data = await api.get<any>(`/kube/pvs`, {
                headers: { "x-kube-context": selectedContext || "" }
            });
            setPvs(data.pvs || []);
        } catch (error) {
            console.error("Failed to fetch PVs:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleRefresh = () => {
        if (selectedContext) {
            fetchPVs();
        }
    };

    const getStatusColor = (status: string) => {
        switch (status.toLowerCase()) {
            case 'available': return 'bg-blue-500/10 text-blue-600';
            case 'bound': return 'bg-green-500/10 text-green-600';
            case 'released': return 'bg-yellow-500/10 text-yellow-600';
            case 'failed': return 'bg-red-500/10 text-red-600';
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
                                    <HardDrive className="h-5 w-5 text-indigo-500" />
                                    {pvs.length} PVs Found
                                </CardTitle>
                                <CardDescription>
                                    {!selectedContext ? "Select a cluster from the top bar to view PVs" : "Cluster-wide storage volumes"}
                                </CardDescription>
                            </div>
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={handleRefresh}
                                disabled={!selectedContext || loading}
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
                                <HardDrive className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    Select a cluster from the top bar to view PVs.
                                </p>
                            </div>
                        ) : pvs.length === 0 ? (
                            <div className="text-center py-12">
                                <HardDrive className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    No PVs found in this cluster.
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
                                                        {item.capacity}
                                                    </span>
                                                    <span className="text-xs text-muted-foreground">
                                                        {item.reclaim_policy}
                                                    </span>
                                                </div>
                                            </div>

                                            <div className="flex flex-col gap-1 items-start lg:items-end min-w-[250px]">
                                                {item.claim && (
                                                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                                                        <LinkIcon className="h-3 w-3" />
                                                        <span className="truncate max-w-[200px]">{item.claim}</span>
                                                    </div>
                                                )}
                                                <p className="text-[10px] text-muted-foreground uppercase tracking-widest mt-1">
                                                    {item.access_modes}
                                                </p>
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
                namespace=""
                name={selectedItem?.name || ""}
                kind="PersistentVolume"
                onUpdate={handleRefresh}
            />
        </div >
    );
}

export default function PVPage() {
    return (
        <Suspense fallback={<div className="flex items-center justify-center min-h-screen"><RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" /></div>}>
            <PVContent />
        </Suspense>
    );
}
