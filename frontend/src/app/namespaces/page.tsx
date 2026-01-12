"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Layers, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

import { API_URL } from "@/lib/config";

interface ContextInfo {
    name: string;
    display_name: string;
}

function NamespacesContent() {
    const searchParams = useSearchParams();
    const selectedContext = searchParams.get("context") || "";

    const [contexts, setContexts] = useState<ContextInfo[]>([]);
    const [namespaces, setNamespaces] = useState<string[]>([]);
    const [namespacesLoading, setNamespacesLoading] = useState(false);

    // Load contexts on mount just for names
    useEffect(() => {
        const fetchContexts = async () => {
            try {
                const res = await fetch(`${API_URL}/kube/contexts`, { credentials: "include" });
                if (res.ok) {
                    const data = await res.json();
                    setContexts(data.contexts || []);
                }
            } catch (error) {
                console.error("Failed to fetch contexts:", error);
            }
        };
        fetchContexts();
    }, []);

    // Load namespaces when context changes
    useEffect(() => {
        if (!selectedContext) return;

        const fetchNamespaces = async () => {
            setNamespacesLoading(true);
            setNamespaces([]);
            try {
                const res = await fetch(`${API_URL}/kube/namespaces?context=${selectedContext}`, { credentials: "include" });
                if (res.status === 401) {
                    window.location.href = "/login";
                    return;
                }
                if (res.ok) {
                    const data = await res.json();
                    setNamespaces(data.namespaces || []);
                }
            } catch (error) {
                console.error("Failed to fetch namespaces:", error);
            } finally {
                setNamespacesLoading(false);
            }
        };
        fetchNamespaces();
    }, [selectedContext]);

    const handleRefresh = () => {
        if (selectedContext) {
            // Re-trigger namespace fetch
            const fetchNamespaces = async () => {
                setNamespacesLoading(true);
                try {
                    const res = await fetch(`${API_URL}/kube/namespaces?context=${selectedContext}`, { credentials: "include" });
                    if (res.ok) {
                        const data = await res.json();
                        setNamespaces(data.namespaces || []);
                    }
                } catch (error) {
                    console.error("Failed to fetch namespaces:", error);
                } finally {
                    setNamespacesLoading(false);
                }
            };
            fetchNamespaces();
        }
    };

    const getDisplayName = (contextName: string) => {
        const ctx = contexts.find(c => c.name === contextName);
        return ctx?.display_name || contextName;
    };

    return (
        <div className="flex flex-col items-center w-full min-h-screen bg-background/50 p-4 md:p-0">
            <div className="w-full max-w-5xl space-y-8">


                {/* Namespaces Grid */}
                <Card className="border-none shadow-2xl shadow-black/5 bg-card/50 backdrop-blur-sm overflow-hidden rounded-3xl">
                    <CardHeader className="border-b bg-card/50 px-8 py-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <CardTitle className="text-lg font-bold flex items-center gap-3">
                                    <Layers className="h-5 w-5 text-green-500" />
                                    Namespaces
                                    {namespaces.length > 0 && (
                                        <span className="text-xs font-normal text-muted-foreground">
                                            ({namespaces.length} found)
                                        </span>
                                    )}
                                </CardTitle>
                            </div>
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={handleRefresh}
                                disabled={!selectedContext || namespacesLoading}
                                className="rounded-xl"
                            >
                                <RefreshCw className={cn("h-4 w-4 mr-2", namespacesLoading && "animate-spin")} />
                                Refresh
                            </Button>
                        </div>
                    </CardHeader>
                    <CardContent className="p-8">
                        {namespacesLoading ? (
                            <div className="flex items-center justify-center py-12">
                                <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
                            </div>
                        ) : !selectedContext ? (
                            <div className="text-center py-12">
                                <Layers className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    Select a cluster from the top bar to view namespaces.
                                </p>
                            </div>
                        ) : namespaces.length === 0 ? (
                            <div className="text-center py-12">
                                <Layers className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    No namespaces found in this cluster.
                                </p>
                            </div>
                        ) : (
                            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                                {namespaces.map(ns => (
                                    <div
                                        key={ns}
                                        className="p-4 bg-muted/30 rounded-2xl border border-muted/20 hover:bg-muted/50 transition-colors cursor-default"
                                    >
                                        <div className="flex items-center gap-3">
                                            <div className="p-2 bg-primary/10 rounded-lg">
                                                <Layers className="h-4 w-4 text-primary" />
                                            </div>
                                            <div className="overflow-hidden">
                                                <p className="font-semibold text-sm truncate">{ns}</p>
                                                <p className="text-[10px] text-muted-foreground uppercase tracking-wider">namespace</p>
                                            </div>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </CardContent>
                </Card>
            </div>
        </div>
    );
}

export default function NamespacesPage() {
    return (
        <Suspense fallback={<div className="flex items-center justify-center min-h-screen"><RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" /></div>}>
            <NamespacesContent />
        </Suspense>
    );
}
