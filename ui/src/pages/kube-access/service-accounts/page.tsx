"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { UserCheck, RefreshCw, Box } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn, formatAge } from "@/lib/utils";
import { api } from "@/lib/api";
import { NamespaceBadge } from "@/components/NamespaceBadge";
import { ResourceDetailsSheet } from "@/components/ResourceDetailsSheet";

interface SAInfo {
    name: string;
    namespace: string;
    secrets: number;
    age: string;
}

function SAContent() {
    const [searchParams] = useSearchParams();
    const selectedContext = searchParams.get("context") || "";
    const selectedNamespace = searchParams.get("namespace") || "";

    const [sas, setSas] = useState<SAInfo[]>([]);
    const [loading, setLoading] = useState(false);
    const searchQuery = searchParams.get("q") || "";
    const [selectedItem, setSelectedItem] = useState<SAInfo | null>(null);

    const filteredItems = sas.filter((item) => item.name.toLowerCase().includes(searchQuery.toLowerCase()));

    useEffect(() => {
        if (!selectedContext || !selectedNamespace) {
            setSas([]);
            return;
        }
        fetchSAs();
    }, [selectedContext, selectedNamespace]);

    const fetchSAs = async () => {
        setLoading(true);
        setSas([]);
        try {
            const data = await api.get<any>(`/kube/service-accounts?namespace=${selectedNamespace}`, {
                headers: { "x-kube-context": selectedContext || "" },
            });
            setSas(data.serviceaccounts || []);
        } catch (error) {
            console.error("Failed to fetch Service Accounts:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleRefresh = () => {
        if (selectedContext && selectedNamespace) {
            fetchSAs();
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
                                    <UserCheck className="h-5 w-5 text-indigo-500" />
                                    {sas.length} Service Accounts Found
                                </CardTitle>
                                <CardDescription>
                                    {!selectedContext || !selectedNamespace
                                        ? "Select a namespace from the top bar to view Service Accounts"
                                        : "Manage identity for processes"}
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
                                <UserCheck className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    Select a cluster from the top bar to view Service Accounts.
                                </p>
                            </div>
                        ) : !selectedNamespace ? (
                            <div className="text-center py-12">
                                <UserCheck className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    Select a namespace from the top bar to view Service Accounts.
                                </p>
                            </div>
                        ) : sas.length === 0 ? (
                            <div className="text-center py-12">
                                <UserCheck className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    No Service Accounts found in this namespace.
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
                                                <p className="font-semibold text-base truncate">{item.name}</p>
                                                <div className="flex items-center gap-3 mt-1">
                                                    <span className="text-xs text-muted-foreground flex items-center gap-1">
                                                        <Box className="h-3 w-3" />
                                                        {item.secrets} Secrets
                                                    </span>
                                                </div>
                                            </div>

                                            <div
                                                className="flex flex-col lg:items-end min-w-[120px]"
                                                onClick={(e) => e.stopPropagation()}
                                            >
                                                <NamespaceBadge namespace={item.namespace} />
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
                namespace={selectedItem?.namespace || ""}
                name={selectedItem?.name || ""}
                kind="ServiceAccount"
                onUpdate={handleRefresh}
            />
        </div>
    );
}

export default function SAPage() {
    return (
        <Suspense
            fallback={
                <div className="flex items-center justify-center min-h-screen">
                    <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
                </div>
            }
        >
            <SAContent />
        </Suspense>
    );
}
