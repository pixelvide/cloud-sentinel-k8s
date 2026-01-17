"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Key, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn, formatAge } from "@/lib/utils";
import { api } from "@/lib/api";
import { NamespaceBadge } from "@/components/NamespaceBadge";
import { ResourceDetailsSheet } from "@/components/ResourceDetailsSheet";

interface LeaseInfo {
    name: string;
    namespace: string;
    age: string;
    holder: string;
}

function LeasesContent() {
    const [searchParams] = useSearchParams();
    const selectedContext = searchParams.get("context") || "";
    const selectedNamespace = searchParams.get("namespace") || "";

    const [resources, setResources] = useState<LeaseInfo[]>([]);
    const [loading, setLoading] = useState(false);
    const searchQuery = searchParams.get("q") || "";
    const [selectedResource, setSelectedResource] = useState<LeaseInfo | null>(null);

    const filteredResources = resources.filter((res) => res.name.toLowerCase().includes(searchQuery.toLowerCase()));

    useEffect(() => {
        if (!selectedContext || !selectedNamespace) {
            setResources([]);
            return;
        }

        fetchResources();
    }, [selectedContext, selectedNamespace]);

    const fetchResources = async () => {
        setLoading(true);
        setResources([]);
        try {
            const data = await api.get<any>(`/kube/leases?namespace=${selectedNamespace}`, {
                headers: { "x-kube-context": selectedContext || "" },
            });
            setResources(data.leases || []);
        } catch (error) {
            console.error("Failed to fetch Leases:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleRefresh = () => {
        if (selectedContext && selectedNamespace) {
            fetchResources();
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
                                    <Key className="h-5 w-5 text-yellow-500" />
                                    {resources.length} Leases Found
                                </CardTitle>
                                <CardDescription>
                                    {!selectedContext || !selectedNamespace
                                        ? "Select a namespace from the top bar to view Leases"
                                        : "Distributed coordination and locking"}
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
                                <Key className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">Select a cluster from the top bar.</p>
                            </div>
                        ) : !selectedNamespace ? (
                            <div className="text-center py-12">
                                <Key className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">Select a namespace from the top bar.</p>
                            </div>
                        ) : resources.length === 0 ? (
                            <div className="text-center py-12">
                                <Key className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">No Leases found.</p>
                            </div>
                        ) : (
                            <div className="grid grid-cols-1 gap-4">
                                {filteredResources.map((res) => (
                                    <div
                                        key={res.name}
                                        className="p-6 bg-muted/30 rounded-2xl border border-muted/20 hover:bg-muted/50 transition-colors cursor-pointer"
                                        onClick={() => setSelectedResource(res)}
                                    >
                                        <div className="flex flex-col lg:flex-row lg:items-center gap-4 justify-between">
                                            <div className="min-w-0 flex-1">
                                                <p className="font-semibold text-base truncate">{res.name}</p>
                                                <div className="flex items-center gap-2 mt-1">
                                                    <span className="text-xs text-muted-foreground">
                                                        Holder:{" "}
                                                        <span className="text-primary font-mono">
                                                            {res.holder || "None"}
                                                        </span>
                                                    </span>
                                                </div>
                                            </div>
                                            <div className="flex items-center gap-6">
                                                <NamespaceBadge namespace={res.namespace} />
                                                <span className="text-xs text-muted-foreground min-w-[80px] text-right">
                                                    {formatAge(res.age)}
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
                isOpen={!!selectedResource}
                onClose={() => setSelectedResource(null)}
                context={selectedContext}
                namespace={selectedResource?.namespace || ""}
                name={selectedResource?.name || ""}
                kind="Lease"
                onUpdate={handleRefresh}
            />
        </div>
    );
}

export default function LeasesPage() {
    return (
        <Suspense
            fallback={
                <div className="flex items-center justify-center min-h-screen">
                    <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
                </div>
            }
        >
            <LeasesContent />
        </Suspense>
    );
}
