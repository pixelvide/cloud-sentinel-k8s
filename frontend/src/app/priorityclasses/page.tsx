"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { ArrowUpCircle, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn, formatAge } from "@/lib/utils";
import { API_URL } from "@/lib/config";
import { ResourceDetailsSheet } from "@/components/ResourceDetailsSheet";

interface PriorityClassInfo {
    name: string;
    value: number;
    age: string;
}

function PriorityClassesContent() {
    const searchParams = useSearchParams();
    const selectedContext = searchParams.get("context") || "";

    const [resources, setResources] = useState<PriorityClassInfo[]>([]);
    const [loading, setLoading] = useState(false);
    const searchQuery = searchParams.get("q") || "";
    const [selectedResource, setSelectedResource] = useState<PriorityClassInfo | null>(null);

    const filteredResources = resources.filter(res =>
        res.name.toLowerCase().includes(searchQuery.toLowerCase())
    );

    useEffect(() => {
        if (!selectedContext) {
            setResources([]);
            return;
        }

        fetchResources();
    }, [selectedContext]);

    const fetchResources = async () => {
        setLoading(true);
        setResources([]);
        try {
            const res = await fetch(`${API_URL}/kube/priorityclasses?context=${selectedContext}`, { credentials: "include" });
            if (res.status === 401) {
                window.location.href = "/login";
                return;
            }
            if (res.ok) {
                const data = await res.json();
                setResources(data.priorityclasses || []);
            }
        } catch (error) {
            console.error("Failed to fetch PriorityClasses:", error);
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
                <Card className="border-none shadow-2xl shadow-black/5 bg-card/50 backdrop-blur-sm overflow-hidden rounded-3xl">
                    <CardHeader className="border-b bg-card/50 px-8 py-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <CardTitle className="text-lg font-bold flex items-center gap-3">
                                    <ArrowUpCircle className="h-5 w-5 text-purple-500" />
                                    {resources.length} Priority Classes Found
                                </CardTitle>
                                <CardDescription>
                                    {!selectedContext ? "Select a cluster from the top bar" : "Cluster-wide priority scheduling settings"}
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
                                <ArrowUpCircle className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">Select a cluster from the top bar.</p>
                            </div>
                        ) : resources.length === 0 ? (
                            <div className="text-center py-12">
                                <ArrowUpCircle className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">No Priority Classes found.</p>
                            </div>
                        ) : (
                            <div className="grid grid-cols-1 gap-4">
                                {filteredResources.map(res => (
                                    <div
                                        key={res.name}
                                        className="p-6 bg-muted/30 rounded-2xl border border-muted/20 hover:bg-muted/50 transition-colors cursor-pointer"
                                        onClick={() => setSelectedResource(res)}
                                    >
                                        <div className="flex flex-col lg:flex-row lg:items-center gap-4 justify-between">
                                            <div className="min-w-0 flex-1">
                                                <p className="font-semibold text-base truncate">{res.name}</p>
                                                <div className="flex items-center gap-2 mt-1">
                                                    <span className="text-xs text-muted-foreground">Priority Value: <span className="text-primary">{res.value}</span></span>
                                                </div>
                                            </div>
                                            <div className="flex items-center gap-6">
                                                <span className="text-xs text-muted-foreground min-w-[80px] text-right">{formatAge(res.age)}</span>
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
                namespace=""
                name={selectedResource?.name || ""}
                kind="PriorityClass"
            />
        </div>
    );
}

export default function PriorityClassesPage() {
    return (
        <Suspense fallback={<div className="flex items-center justify-center min-h-screen"><RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" /></div>}>
            <PriorityClassesContent />
        </Suspense>
    );
}
