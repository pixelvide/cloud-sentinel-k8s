"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Link2, RefreshCw, ShieldCheck } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn, formatAge } from "@/lib/utils";
import { API_URL } from "@/lib/config";
import { ResourceDetailsSheet } from "@/components/ResourceDetailsSheet";

interface BindingInfo {
    name: string;
    role: string;
    age: string;
}

function ClusterRoleBindingsContent() {
    const searchParams = useSearchParams();
    const selectedContext = searchParams.get("context") || "";

    const [bindings, setBindings] = useState<BindingInfo[]>([]);
    const [loading, setLoading] = useState(false);
    const searchQuery = searchParams.get("q") || "";
    const [selectedItem, setSelectedItem] = useState<BindingInfo | null>(null);

    const filteredItems = bindings.filter(item =>
        item.name.toLowerCase().includes(searchQuery.toLowerCase())
    );

    useEffect(() => {
        if (!selectedContext) {
            setBindings([]);
            return;
        }
        fetchBindings();
    }, [selectedContext]);

    const fetchBindings = async () => {
        setLoading(true);
        setBindings([]);
        try {
            const res = await fetch(`${API_URL}/kube/clusterrolebindings?context=${selectedContext}`, { credentials: "include" });
            if (res.status === 401) {
                window.location.href = "/login";
                return;
            }
            if (res.ok) {
                const data = await res.json();
                setBindings(data.clusterrolebindings || []);
            }
        } catch (error) {
            console.error("Failed to fetch Cluster Role Bindings:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleRefresh = () => {
        if (selectedContext) {
            fetchBindings();
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
                                    <Link2 className="h-5 w-5 text-indigo-500" />
                                    {bindings.length} Cluster Role Bindings Found
                                </CardTitle>
                                <CardDescription>
                                    {!selectedContext ? "Select a cluster from the top bar to view bindings" : "Manage cluster-wide role assignments"}
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
                                <Link2 className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    Select a cluster from the top bar to view Cluster Role Bindings.
                                </p>
                            </div>
                        ) : bindings.length === 0 ? (
                            <div className="text-center py-12">
                                <Link2 className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    No Cluster Role Bindings found in this cluster.
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
                                                <div className="flex items-center gap-2 mt-1">
                                                    <ShieldCheck className="h-3 w-3 text-muted-foreground" />
                                                    <span className="text-xs text-muted-foreground truncate max-w-[400px]">
                                                        Role: {item.role}
                                                    </span>
                                                </div>
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
                kind="ClusterRoleBinding"
            />
        </div >
    );
}

export default function ClusterRoleBindingsPage() {
    return (
        <Suspense fallback={<div className="flex items-center justify-center min-h-screen"><RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" /></div>}>
            <ClusterRoleBindingsContent />
        </Suspense>
    );
}
