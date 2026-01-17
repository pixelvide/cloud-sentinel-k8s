"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Layers, RefreshCw } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { cn, formatAge } from "@/lib/utils";
import { api } from "@/lib/api";
import { ResourceDetailsSheet } from "@/components/ResourceDetailsSheet";

interface StorageClassInfo {
    name: string;
    provisioner: string;
    reclaim_policy: string;
    volume_binding_mode: string;
    is_default: boolean;
    age: string;
}

function StorageClassesContent() {
    const [searchParams] = useSearchParams();
    const selectedContext = searchParams.get("context") || "";

    const [classes, setClasses] = useState<StorageClassInfo[]>([]);
    const [loading, setLoading] = useState(false);
    const searchQuery = searchParams.get("q") || "";
    const [selectedItem, setSelectedItem] = useState<StorageClassInfo | null>(null);

    const filteredItems = classes.filter((item) => item.name.toLowerCase().includes(searchQuery.toLowerCase()));

    useEffect(() => {
        if (!selectedContext) {
            setClasses([]);
            return;
        }
        fetchStorageClasses();
    }, [selectedContext]);

    const fetchStorageClasses = async () => {
        setLoading(true);
        setClasses([]);
        try {
            const data = await api.get<any>(`/kube/storage-classes`, {
                headers: { "x-kube-context": selectedContext || "" },
            });
            setClasses(data.storageclasses || []);
        } catch (error) {
            console.error("Failed to fetch storage classes:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleRefresh = () => {
        if (selectedContext) {
            fetchStorageClasses();
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
                                    <Layers className="h-5 w-5 text-indigo-500" />
                                    {classes.length} Storage Classes Found
                                </CardTitle>
                                <CardDescription>
                                    {!selectedContext
                                        ? "Select a cluster from the top bar to view storage classes"
                                        : "Cluster-wide storage provisioning configs"}
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
                                <Layers className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    Select a cluster from the top bar to view storage classes.
                                </p>
                            </div>
                        ) : classes.length === 0 ? (
                            <div className="text-center py-12">
                                <Layers className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    No storage classes found in this cluster.
                                </p>
                            </div>
                        ) : (
                            <div className="grid grid-cols-1 gap-4">
                                {filteredItems.map((item) => (
                                    <div
                                        key={item.name}
                                        className="p-6 bg-muted/30 rounded-2xl border border-muted/20 hover:bg-muted/50 transition-colors cursor-pointer"
                                        onClick={() => setSelectedItem(item)}
                                    >
                                        <div className="flex flex-col lg:flex-row lg:items-center gap-4 justify-between">
                                            <div className="min-w-0 flex-1">
                                                <div className="flex items-center gap-2">
                                                    <p className="font-semibold text-base truncate">{item.name}</p>
                                                    {item.is_default && (
                                                        <Badge
                                                            variant="secondary"
                                                            className="bg-indigo-50 text-indigo-700 border-indigo-200 text-[10px] font-bold px-1.5 h-4"
                                                        >
                                                            default
                                                        </Badge>
                                                    )}
                                                </div>
                                                <p className="text-xs text-muted-foreground mt-1">
                                                    Provisioner: {item.provisioner}
                                                </p>
                                            </div>

                                            <div className="flex flex-col gap-1 items-start lg:items-end min-w-[200px]">
                                                <p className="text-xs text-muted-foreground">
                                                    Policy: {item.reclaim_policy}
                                                </p>
                                                <p className="text-[10px] text-muted-foreground uppercase tracking-widest mt-1">
                                                    Binding: {item.volume_binding_mode}
                                                </p>
                                            </div>

                                            <div className="flex flex-col items-end min-w-[80px]">
                                                <span className="text-xs text-muted-foreground">
                                                    {formatAge(item.age)}
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
                isOpen={!!selectedItem}
                onClose={() => setSelectedItem(null)}
                context={selectedContext}
                namespace=""
                name={selectedItem?.name || ""}
                kind="StorageClass"
                onUpdate={handleRefresh}
            />
        </div>
    );
}

export default function StorageClassesPage() {
    return (
        <Suspense
            fallback={
                <div className="flex items-center justify-center min-h-screen">
                    <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
                </div>
            }
        >
            <StorageClassesContent />
        </Suspense>
    );
}
