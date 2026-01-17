"use client";

import { useEffect, useState, Suspense } from "react";
import { useParams, useSearchParams } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Box, RefreshCw, Layers, ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn, formatAge } from "@/lib/utils";
import { NamespaceBadge } from "@/components/NamespaceBadge";
import { ResourceDetailsSheet } from "@/components/ResourceDetailsSheet";
import { api } from "@/lib/api";
import { Link } from "react-router-dom";
import { Badge } from "@/components/ui/badge";

interface CRResource {
    name: string;
    namespace: string;
    created: string;
    age: string;
    labels: Record<string, string>;
}

interface CRDMeta {
    scope: string;
    group: string;
    version: string;
    kind: string;
    resource: string;
}

function CustomResourcesContent() {
    const params = useParams();
    const crdName = params.crd as string;

    const [searchParams] = useSearchParams();
    const selectedContext = searchParams.get("context") || "";
    const selectedNamespace = searchParams.get("namespace") || "";
    const searchQuery = searchParams.get("q") || "";

    const [resources, setResources] = useState<CRResource[]>([]);
    const [crdMeta, setCrdMeta] = useState<CRDMeta | null>(null);
    const [loading, setLoading] = useState(false);
    const [selectedResource, setSelectedResource] = useState<CRResource | null>(null);

    const filteredResources = resources.filter((res) => res.name.toLowerCase().includes(searchQuery.toLowerCase()));

    useEffect(() => {
        if (!selectedContext) {
            setResources([]);
            return;
        }
        fetchResources();
    }, [selectedContext, selectedNamespace, crdName]);

    const fetchResources = async () => {
        setLoading(true);
        setResources([]);
        try {
            const url = `/kube/crds/${encodeURIComponent(crdName)}/resources?namespace=${encodeURIComponent(selectedNamespace)}`;
            const data = await api.get<{ items: CRResource[]; crd: CRDMeta }>(url, {
                headers: { "x-kube-context": selectedContext || "" },
            });
            setResources(data.items || []);
            setCrdMeta(data.crd);
        } catch (error) {
            console.error("Failed to fetch custom resources:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleRefresh = () => {
        if (selectedContext) {
            fetchResources();
        }
    };

    return (
        <div className="flex flex-col items-center w-full min-h-screen bg-background/50 p-4 md:p-0">
            <div className="w-full max-w-6xl space-y-8">
                {/* Header Back Button */}
                <div className="flex items-center gap-2"></div>

                <Card className="border-none shadow-2xl shadow-black/5 bg-card/50 backdrop-blur-sm overflow-hidden rounded-3xl">
                    <CardHeader className="border-b bg-card/50 px-8 py-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <CardTitle className="text-lg font-bold flex items-center gap-3">
                                    <Layers className="h-5 w-5 text-primary" />
                                    {crdMeta ? crdMeta.kind : crdName}
                                    {crdMeta && (
                                        <Badge variant="outline" className="text-xs font-normal">
                                            {crdMeta.group}/{crdMeta.version}
                                        </Badge>
                                    )}
                                </CardTitle>
                                <CardDescription>
                                    {(crdMeta?.scope === "Namespace" || crdMeta?.scope === "Namespaced") &&
                                    (!selectedNamespace || selectedNamespace === "__all__")
                                        ? "Select a namespace to view resources"
                                        : `${resources.length} items found`}
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
                                <Box className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">Select a cluster from the top bar.</p>
                            </div>
                        ) : resources.length === 0 ? (
                            <div className="text-center py-12">
                                <Box className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">No resources found.</p>
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
                                            {/* Name */}
                                            <div className="min-w-0 flex-1">
                                                <p className="font-semibold text-base truncate">{res.name}</p>
                                                {/* Labels */}
                                                <div className="flex items-center gap-2 mt-2 flex-wrap">
                                                    {res.labels &&
                                                        Object.entries(res.labels)
                                                            .slice(0, 3)
                                                            .map(([k, v]) => (
                                                                <span
                                                                    key={k}
                                                                    className="text-[10px] bg-background border px-2 py-0.5 rounded-full text-muted-foreground truncate max-w-[200px]"
                                                                >
                                                                    {k}={v}
                                                                </span>
                                                            ))}
                                                    {res.labels && Object.keys(res.labels).length > 3 && (
                                                        <span className="text-[10px] text-muted-foreground">
                                                            +{Object.keys(res.labels).length - 3} more
                                                        </span>
                                                    )}
                                                </div>
                                            </div>

                                            {/* Namespace */}
                                            <div
                                                className="flex flex-col items-end min-w-[120px]"
                                                onClick={(e) => e.stopPropagation()}
                                            >
                                                <NamespaceBadge namespace={res.namespace} />
                                            </div>

                                            {/* Age */}
                                            <div className="flex flex-col items-end min-w-[80px]">
                                                <span className="text-xs text-muted-foreground">
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
                kind={crdMeta?.kind || ""}
                crdName={crdName}
                onUpdate={handleRefresh}
            />
        </div>
    );
}

export default function CRDResourcesPage() {
    return (
        <Suspense
            fallback={
                <div className="flex items-center justify-center min-h-screen">
                    <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
                </div>
            }
        >
            <CustomResourcesContent />
        </Suspense>
    );
}
