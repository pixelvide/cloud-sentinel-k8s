"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Server, Globe, RefreshCw, Link as LinkIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn, formatAge } from "@/lib/utils";
import { api } from "@/lib/api";
import { NamespaceBadge } from "@/components/NamespaceBadge";
import { ResourceDetailsSheet } from "@/components/ResourceDetailsSheet";

interface IngressInfo {
    name: string;
    hosts: string[];
    ips: string[];
    namespace: string;
    age: string;
}

function IngressesContent() {
    const [searchParams] = useSearchParams();
    const selectedContext = searchParams.get("context") || "";
    const selectedNamespace = searchParams.get("namespace") || "";

    const [ingresses, setIngresses] = useState<IngressInfo[]>([]);
    const [ingLoading, setIngLoading] = useState(false);
    const searchQuery = searchParams.get("q") || "";
    const [selectedIngress, setSelectedIngress] = useState<IngressInfo | null>(null);

    const filteredIngresses = ingresses.filter((ing) => ing.name.toLowerCase().includes(searchQuery.toLowerCase()));

    useEffect(() => {
        if (!selectedContext || !selectedNamespace) {
            setIngresses([]);
            return;
        }
        fetchIngresses();
    }, [selectedContext, selectedNamespace]);

    const fetchIngresses = async () => {
        setIngLoading(true);
        setIngresses([]);
        try {
            const data = await api.get<any>(`/kube/ingresses?namespace=${selectedNamespace}`, {
                headers: { "x-kube-context": selectedContext || "" },
            });
            setIngresses(data.ingresses || []);
        } catch (error) {
            console.error("Failed to fetch ingresses:", error);
        } finally {
            setIngLoading(false);
        }
    };

    const handleRefresh = () => {
        if (selectedContext && selectedNamespace) {
            fetchIngresses();
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
                                    <Globe className="h-5 w-5 text-pink-500" />
                                    {ingresses.length} Ingresses Found
                                </CardTitle>
                                <CardDescription>
                                    {!selectedContext || !selectedNamespace
                                        ? "Select a namespace from the top bar to view ingresses"
                                        : null}
                                </CardDescription>
                            </div>
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={handleRefresh}
                                disabled={!selectedContext || !selectedNamespace || ingLoading}
                                className="rounded-xl"
                            >
                                <RefreshCw className={cn("h-4 w-4 mr-2", ingLoading && "animate-spin")} />
                                Refresh
                            </Button>
                        </div>
                    </CardHeader>
                    <CardContent className="p-8">
                        {ingLoading ? (
                            <div className="flex items-center justify-center py-12">
                                <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
                            </div>
                        ) : !selectedContext ? (
                            <div className="text-center py-12">
                                <Globe className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    Select a cluster from the top bar to view ingresses.
                                </p>
                            </div>
                        ) : !selectedNamespace ? (
                            <div className="text-center py-12">
                                <Globe className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    Select a namespace from the top bar to view ingresses.
                                </p>
                            </div>
                        ) : ingresses.length === 0 ? (
                            <div className="text-center py-12">
                                <Globe className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">No ingresses found in this namespace.</p>
                            </div>
                        ) : (
                            <div className="grid grid-cols-1 gap-4">
                                {filteredIngresses.map((ing) => (
                                    <div
                                        key={ing.name}
                                        className="p-6 bg-muted/30 rounded-2xl border border-muted/20 hover:bg-muted/50 transition-colors cursor-pointer"
                                        onClick={() => setSelectedIngress(ing)}
                                    >
                                        <div className="flex flex-col lg:flex-row lg:items-center gap-4 justify-between">
                                            <div className="min-w-0 flex-1">
                                                <p className="font-semibold text-base truncate">{ing.name}</p>
                                                <div className="flex flex-col gap-1 mt-1">
                                                    {(ing.hosts || []).map((host) => (
                                                        <a
                                                            key={host}
                                                            href={`http://${host}`}
                                                            target="_blank"
                                                            rel="noreferrer"
                                                            className="text-xs text-blue-500 hover:underline flex items-center gap-1"
                                                            onClick={(e) => e.stopPropagation()}
                                                        >
                                                            <LinkIcon className="h-3 w-3" />
                                                            {host}
                                                        </a>
                                                    ))}
                                                </div>
                                            </div>

                                            <div className="flex flex-col gap-1 items-start lg:items-end min-w-[200px]">
                                                {(ing.ips?.length || 0) > 0 && (
                                                    <>
                                                        <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
                                                            <Server className="h-3 w-3" />
                                                            <span>Load Balancer</span>
                                                        </div>
                                                        <div className="flex flex-wrap gap-2 justify-end">
                                                            {ing.ips.map((ip) => (
                                                                <span
                                                                    key={ip}
                                                                    className="text-xs bg-muted border rounded-md px-2 py-1 font-mono"
                                                                >
                                                                    {ip}
                                                                </span>
                                                            ))}
                                                        </div>
                                                    </>
                                                )}
                                            </div>

                                            {/* Namespace */}
                                            <div
                                                className="flex flex-col items-end min-w-[120px]"
                                                onClick={(e) => e.stopPropagation()}
                                            >
                                                <NamespaceBadge namespace={ing.namespace} />
                                            </div>

                                            {/* Age */}
                                            <div className="flex flex-col items-end min-w-[80px]">
                                                <span className="text-xs text-muted-foreground">
                                                    {formatAge(ing.age)}
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
                isOpen={!!selectedIngress}
                onClose={() => setSelectedIngress(null)}
                context={selectedContext}
                namespace={selectedIngress?.namespace || ""}
                name={selectedIngress?.name || ""}
                kind="Ingress"
                onUpdate={handleRefresh}
            />
        </div>
    );
}

export default function IngressesPage() {
    return (
        <Suspense
            fallback={
                <div className="flex items-center justify-center min-h-screen">
                    <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
                </div>
            }
        >
            <IngressesContent />
        </Suspense>
    );
}
