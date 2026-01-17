"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { HardDrive, RefreshCw, Cpu, MemoryStick, CheckCircle2, XCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn, formatAge } from "@/lib/utils";
import { ResourceDetailsSheet } from "@/components/ResourceDetailsSheet";
import { api } from "@/lib/api";

interface ContextInfo {
    name: string;
    display_name: string;
}

interface NodeInfo {
    name: string;
    status: string;
    roles: string[];
    kubelet_version: string;
    os: string;
    architecture: string;
    cpu_capacity: string;
    memory_capacity: string;
    cpu_allocatable: string;
    memory_allocatable: string;
    age: string;
}

function NodesContent() {
    const [searchParams] = useSearchParams();
    const selectedContext = searchParams.get("context") || "";

    const [contexts, setContexts] = useState<ContextInfo[]>([]);
    const [nodes, setNodes] = useState<NodeInfo[]>([]);
    const [nodesLoading, setNodesLoading] = useState(false);
    const searchQuery = searchParams.get("q") || "";
    const [selectedNode, setSelectedNode] = useState<NodeInfo | null>(null);

    const filteredNodes = nodes.filter((node) => node.name.toLowerCase().includes(searchQuery.toLowerCase()));

    // Load contexts for display names
    useEffect(() => {
        const fetchContexts = async () => {
            try {
                const data = await api.get<any>("/kube/contexts");
                setContexts(data.contexts || []);
            } catch (error) {
                console.error("Failed to fetch contexts:", error);
            }
        };
        fetchContexts();
    }, []);

    // Load nodes when context changes
    useEffect(() => {
        if (!selectedContext) return;

        fetchNodes();
    }, [selectedContext]);

    const fetchNodes = async () => {
        setNodesLoading(true);
        setNodes([]);
        try {
            const data = await api.get<any>(`/kube/nodes`, {
                headers: { "x-kube-context": selectedContext || "" },
            });
            setNodes(data.nodes || []);
        } catch (error) {
            console.error("Failed to fetch nodes:", error);
        } finally {
            setNodesLoading(false);
        }
    };

    const handleRefresh = () => {
        if (selectedContext) {
            fetchNodes();
        }
    };

    const getDisplayName = (contextName: string) => {
        const ctx = contexts.find((c) => c.name === contextName);
        return ctx?.display_name || contextName;
    };

    const formatMemory = (memory: string) => {
        // Convert Ki to GB for display
        const match = memory.match(/^(\d+)Ki$/);
        if (match) {
            const ki = parseInt(match[1]);
            const gb = (ki / 1024 / 1024).toFixed(1);
            return `${gb} GB`;
        }
        return memory;
    };

    return (
        <div className="flex flex-col items-center w-full min-h-screen bg-background/50 p-4 md:p-0">
            <div className="w-full max-w-6xl space-y-8">
                {/* Nodes List */}
                <Card className="border-none shadow-2xl shadow-black/5 bg-card/50 backdrop-blur-sm overflow-hidden rounded-3xl">
                    <CardHeader className="border-b bg-card/50 px-8 py-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <CardTitle className="text-lg font-bold flex items-center gap-3">
                                    <HardDrive className="h-5 w-5 text-green-500" />
                                    Cluster Nodes
                                    {nodes.length > 0 && (
                                        <span className="text-xs font-normal text-muted-foreground">
                                            ({nodes.length} node{nodes.length !== 1 ? "s" : ""})
                                        </span>
                                    )}
                                </CardTitle>
                            </div>
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={handleRefresh}
                                disabled={!selectedContext || nodesLoading}
                                className="rounded-xl"
                            >
                                <RefreshCw className={cn("h-4 w-4 mr-2", nodesLoading && "animate-spin")} />
                                Refresh
                            </Button>
                        </div>
                    </CardHeader>
                    <CardContent className="p-8">
                        {nodesLoading ? (
                            <div className="flex items-center justify-center py-12">
                                <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
                            </div>
                        ) : !selectedContext ? (
                            <div className="text-center py-12">
                                <HardDrive className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    Select a cluster from the top bar to view nodes.
                                </p>
                            </div>
                        ) : nodes.length === 0 ? (
                            <div className="text-center py-12">
                                <HardDrive className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">No nodes found in this cluster.</p>
                            </div>
                        ) : (
                            <div className="space-y-4">
                                {filteredNodes.map((node) => (
                                    <div
                                        key={node.name}
                                        className="p-6 bg-muted/30 rounded-2xl border border-muted/20 hover:bg-muted/50 transition-colors cursor-pointer"
                                        onClick={() => setSelectedNode(node)}
                                    >
                                        <div className="flex flex-col lg:flex-row lg:items-center gap-4">
                                            {/* Node Name and Status */}
                                            <div className="flex items-center gap-4 flex-1 min-w-0">
                                                <div
                                                    className={cn(
                                                        "p-3 rounded-xl",
                                                        node.status === "Ready" ? "bg-green-500/10" : "bg-red-500/10"
                                                    )}
                                                >
                                                    {node.status === "Ready" ? (
                                                        <CheckCircle2 className="h-6 w-6 text-green-500" />
                                                    ) : (
                                                        <XCircle className="h-6 w-6 text-red-500" />
                                                    )}
                                                </div>
                                                <div className="min-w-0">
                                                    <p className="font-semibold text-base truncate">{node.name}</p>
                                                    <div className="flex items-center gap-2 mt-1 flex-wrap">
                                                        {node.roles.map((role) => (
                                                            <span
                                                                key={role}
                                                                className="text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 bg-primary/10 text-primary rounded-full"
                                                            >
                                                                {role}
                                                            </span>
                                                        ))}
                                                        <span className="text-xs text-muted-foreground">
                                                            {node.kubelet_version}
                                                        </span>
                                                    </div>
                                                </div>
                                            </div>

                                            {/* Resource Info */}
                                            <div className="flex items-center gap-6 flex-wrap">
                                                <div className="flex items-center gap-2">
                                                    <Cpu className="h-4 w-4 text-blue-500" />
                                                    <div className="text-sm">
                                                        <span className="font-medium">{node.cpu_allocatable}</span>
                                                        <span className="text-muted-foreground text-xs">
                                                            {" "}
                                                            / {node.cpu_capacity} CPU
                                                        </span>
                                                    </div>
                                                </div>
                                                <div className="flex items-center gap-2">
                                                    <MemoryStick className="h-4 w-4 text-purple-500" />
                                                    <div className="text-sm">
                                                        <span className="font-medium">
                                                            {formatMemory(node.memory_allocatable)}
                                                        </span>
                                                        <span className="text-muted-foreground text-xs">
                                                            {" "}
                                                            / {formatMemory(node.memory_capacity)}
                                                        </span>
                                                    </div>
                                                </div>
                                                <div className="text-xs text-muted-foreground">
                                                    {node.os}/{node.architecture}
                                                </div>
                                            </div>

                                            {/* Age */}
                                            <div className="flex flex-col items-end min-w-[80px]">
                                                <span className="text-xs text-muted-foreground">
                                                    {formatAge(node.age)}
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
                isOpen={!!selectedNode}
                onClose={() => setSelectedNode(null)}
                context={selectedContext}
                namespace=""
                name={selectedNode?.name || ""}
                kind="Node"
                onUpdate={handleRefresh}
            />
        </div>
    );
}

export default function NodesPage() {
    return (
        <Suspense
            fallback={
                <div className="flex items-center justify-center min-h-screen">
                    <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
                </div>
            }
        >
            <NodesContent />
        </Suspense>
    );
}
