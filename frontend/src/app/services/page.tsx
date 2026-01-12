"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Grid, RefreshCw, Radio, Link as LinkIcon, ExternalLink, Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { cn, formatAge } from "@/lib/utils";
import { API_URL } from "@/lib/config";
import { NamespaceBadge } from "@/components/NamespaceBadge";
import { ResourceDetailsSheet } from "@/components/ResourceDetailsSheet";

interface ServiceInfo {
    name: string;
    type: string;
    cluster_ip: string;
    external_ips: string[];
    ports: string[];
    age: string;
    namespace: string;
}

function ServicesContent() {
    const searchParams = useSearchParams();
    const selectedContext = searchParams.get("context") || "";
    const selectedNamespace = searchParams.get("namespace") || "";

    const [services, setServices] = useState<ServiceInfo[]>([]);
    const [servicesLoading, setServicesLoading] = useState(false);
    const [searchQuery, setSearchQuery] = useState("");
    const [selectedService, setSelectedService] = useState<ServiceInfo | null>(null);

    const filteredServices = services.filter(svc =>
        svc.name.toLowerCase().includes(searchQuery.toLowerCase())
    );

    useEffect(() => {
        if (!selectedContext || !selectedNamespace) {
            setServices([]);
            return;
        }
        fetchServices();
    }, [selectedContext, selectedNamespace]);

    const fetchServices = async () => {
        setServicesLoading(true);
        setServices([]);
        try {
            const res = await fetch(`${API_URL}/kube/services?context=${selectedContext}&namespace=${selectedNamespace}`, { credentials: "include" });
            if (res.status === 401) {
                window.location.href = "/login";
                return;
            }
            if (res.ok) {
                const data = await res.json();
                setServices(data.services || []);
            }
        } catch (error) {
            console.error("Failed to fetch services:", error);
        } finally {
            setServicesLoading(false);
        }
    };

    const handleRefresh = () => {
        if (selectedContext && selectedNamespace) {
            fetchServices();
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
                                    <Grid className="h-5 w-5 text-indigo-500" />
                                    {services.length} Services Found
                                </CardTitle>
                                <CardDescription>
                                    {!selectedContext || !selectedNamespace ? "Select a namespace from the top bar to view services" : null}
                                </CardDescription>
                            </div>
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={handleRefresh}
                                disabled={!selectedContext || !selectedNamespace || servicesLoading}
                                className="rounded-xl"
                            >
                                <RefreshCw className={cn("h-4 w-4 mr-2", servicesLoading && "animate-spin")} />
                                Refresh
                            </Button>
                        </div>
                        {services.length > 0 && (
                            <div className="relative mt-4">
                                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                                <Input
                                    placeholder="Search services by name..."
                                    value={searchQuery}
                                    onChange={(e) => setSearchQuery(e.target.value)}
                                    className="pl-10"
                                />
                            </div>
                        )}
                    </CardHeader>
                    <CardContent className="p-8">
                        {servicesLoading ? (
                            <div className="flex items-center justify-center py-12">
                                <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
                            </div>
                        ) : !selectedContext ? (
                            <div className="text-center py-12">
                                <Grid className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    Select a cluster from the top bar to view services.
                                </p>
                            </div>
                        ) : !selectedNamespace ? (
                            <div className="text-center py-12">
                                <Grid className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    Select a namespace from the top bar to view services.
                                </p>
                            </div>
                        ) : services.length === 0 ? (
                            <div className="text-center py-12">
                                <Grid className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    No services found in this namespace.
                                </p>
                            </div>
                        ) : (
                            <div className="grid grid-cols-1 gap-4">
                                {filteredServices.map(svc => (
                                    <div
                                        key={svc.name}
                                        className="p-6 bg-muted/30 rounded-2xl border border-muted/20 hover:bg-muted/50 transition-colors cursor-pointer"
                                        onClick={() => setSelectedService(svc)}
                                    >
                                        <div className="flex flex-col lg:flex-row lg:items-center gap-4 justify-between">
                                            <div className="min-w-0 flex-1">
                                                <p className="font-semibold text-base truncate">{svc.name}</p>
                                                <div className="flex items-center gap-2 mt-1 flex-wrap">
                                                    <span className="text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full bg-blue-500/10 text-blue-600">
                                                        {svc.type}
                                                    </span>
                                                    <span className="text-xs text-muted-foreground flex items-center gap-1">
                                                        <Radio className="h-3 w-3" />
                                                        {svc.cluster_ip}
                                                    </span>
                                                    {svc.external_ips && svc.external_ips.length > 0 && (
                                                        <span className="text-xs text-muted-foreground flex items-center gap-1">
                                                            <ExternalLink className="h-3 w-3" />
                                                            {svc.external_ips.join(", ")}
                                                        </span>
                                                    )}
                                                </div>
                                            </div>

                                            <div className="flex flex-col gap-1 items-start lg:items-end min-w-[200px]">
                                                <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
                                                    <LinkIcon className="h-3 w-3" />
                                                    <span>Ports</span>
                                                </div>
                                                <div className="flex flex-wrap gap-2 justify-end">
                                                    {svc.ports.map(p => (
                                                        <span key={p} className="text-xs bg-muted border rounded-md px-2 py-1 font-mono">
                                                            {p}
                                                        </span>
                                                    ))}
                                                </div>
                                            </div>

                                            {/* Namespace */}
                                            <div
                                                className="flex flex-col items-end min-w-[120px]"
                                                onClick={(e) => e.stopPropagation()}
                                            >
                                                <NamespaceBadge namespace={svc.namespace} />
                                            </div>

                                            {/* Age */}
                                            <div className="flex flex-col items-end min-w-[80px]">
                                                <span className="text-xs text-muted-foreground">{formatAge(svc.age)}</span>
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
                isOpen={!!selectedService}
                onClose={() => setSelectedService(null)}
                context={selectedContext}
                namespace={selectedService?.namespace || ""}
                name={selectedService?.name || ""}
                kind="Service"
            />
        </div >

    );
}

export default function ServicesPage() {
    return (
        <Suspense fallback={<div className="flex items-center justify-center min-h-screen"><RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" /></div>}>
            <ServicesContent />
        </Suspense>
    );
}
