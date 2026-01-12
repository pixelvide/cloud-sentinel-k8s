"use client";

import { useEffect, useState, Suspense } from "react";
import { useRouter, usePathname, useSearchParams } from "next/navigation";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { MultiSelect } from "@/components/ui/multi-select";
import { Server, Layers } from "lucide-react";

interface ContextInfo {
    name: string;
    display_name: string;
}

function ClusterContextSelectorContent() {
    const router = useRouter();
    const pathname = usePathname();
    const searchParams = useSearchParams();

    const [contexts, setContexts] = useState<ContextInfo[]>([]);
    const [namespaces, setNamespaces] = useState<string[]>([]);
    const [defaultContext, setDefaultContext] = useState("");
    const [loading, setLoading] = useState(false);
    const [nsLoading, setNsLoading] = useState(false);

    // Get context and namespace from URL or default
    const currentContext = searchParams.get("context") || "";
    // Parse comma-separated namespaces
    const currentNamespaces = searchParams.get("namespace") ? searchParams.get("namespace")!.split(",") : [];

    useEffect(() => {
        const fetchContexts = async () => {
            setLoading(true);
            try {
                const res = await fetch("/api/v1/kube/contexts", { credentials: "include" });
                if (res.ok) {
                    const data = await res.json();
                    setContexts(data.contexts || []);
                    if (data.current) {
                        setDefaultContext(data.current);
                    } else if (data.contexts && data.contexts.length > 0) {
                        setDefaultContext(data.contexts[0].name);
                    }
                }
            } catch (error) {
                console.error("Failed to fetch contexts:", error);
            } finally {
                setLoading(false);
            }
        };
        fetchContexts();
    }, []);

    useEffect(() => {
        // Enforce default context if none selected and we have a default
        if (!loading && contexts.length > 0 && !currentContext && defaultContext) {
            updateContext(defaultContext);
        }
    }, [loading, contexts, currentContext, defaultContext]);

    // Fetch namespaces when context changes
    useEffect(() => {
        if (!currentContext) {
            setNamespaces([]);
            return;
        }

        const fetchNamespaces = async () => {
            setNsLoading(true);
            try {
                const res = await fetch(`/api/v1/kube/namespaces?context=${currentContext}`, { credentials: "include" });
                if (res.ok) {
                    const data = await res.json();
                    const nsList = data.namespaces || [];
                    setNamespaces(nsList);

                    // Logic: Default to 'All Namespaces' if not set
                    if (!searchParams.get("namespace")) {
                        const params = new URLSearchParams(searchParams.toString());
                        params.set("namespace", "__all__");
                        // use replace from router, not redirect
                        router.replace(`${pathname}?${params.toString()}`);
                    }
                }
            } catch (error) {
                console.error("Failed to fetch namespaces:", error);
                setNamespaces([]);
            } finally {
                setNsLoading(false);
            }
        };

        fetchNamespaces();
    }, [currentContext]);


    const updateContext = (ctx: string) => {
        const params = new URLSearchParams(searchParams.toString());
        params.set("context", ctx);
        // Default to All Namespaces when context changes
        params.set("namespace", "__all__");
        router.replace(`${pathname}?${params.toString()}`);
    };

    const updateNamespaces = (selected: string[]) => {
        const params = new URLSearchParams(searchParams.toString());
        if (selected.length > 0) {
            params.set("namespace", selected.join(","));
        } else {
            params.delete("namespace");
        }
        router.replace(`${pathname}?${params.toString()}`);
    };

    if (contexts.length === 0 && !loading) return null;

    return (
        <div className="flex flex-col md:flex-row items-stretch md:items-center gap-3 md:gap-4 w-full md:w-auto">
            <div className="flex items-center gap-2">
                <Server className="h-4 w-4 text-muted-foreground hidden md:block" />
                <Select value={currentContext} onValueChange={updateContext}>
                    <SelectTrigger className="w-full md:w-[250px] h-9 text-sm">
                        <SelectValue placeholder="Select Cluster" />
                    </SelectTrigger>
                    <SelectContent>
                        {contexts.map((c) => (
                            <SelectItem key={c.name} value={c.name} className="text-sm">
                                {c.display_name}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
            </div>

            {currentContext && (
                <div className="flex items-center gap-2 w-full md:w-[300px]">
                    <Layers className="h-4 w-4 text-muted-foreground shrink-0 hidden md:block" />
                    <MultiSelect
                        options={namespaces.map(ns => ({ label: ns, value: ns }))}
                        selected={currentNamespaces}
                        onChange={updateNamespaces}
                        placeholder={nsLoading ? "Loading..." : "Select Namespaces"}
                        loading={nsLoading}
                        allOption={{ label: "All Namespaces", value: "__all__" }}
                    />
                </div>
            )}
        </div>
    );
}

export function ClusterContextSelector() {
    return (
        <Suspense fallback={<div className="flex gap-4"><div className="w-[200px] h-9 bg-muted/20 animate-pulse rounded-md" /><div className="w-[200px] h-9 bg-muted/20 animate-pulse rounded-md" /></div>}>
            <ClusterContextSelectorContent />
        </Suspense>
    );
}
