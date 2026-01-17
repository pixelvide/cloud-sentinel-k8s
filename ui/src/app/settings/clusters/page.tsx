"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Cloud, ShieldCheck, Edit2, Check, X, Server, CheckCircle2 } from "lucide-react";
import { Link } from "react-router-dom";
import { cn } from "@/lib/utils";
import { api } from "@/lib/api";
import { KubeConfigUpload } from "@/components/settings/KubeConfigUpload";

interface ContextInfo {
    name: string;
    display_name: string;
}

export default function ClusterSettingsPage() {
    const [kubeConfig, setKubeConfig] = useState("");
    const [loading, setLoading] = useState(false);
    const [isValidating, setIsValidating] = useState(false);
    const [validationStatus, setValidationStatus] = useState<{ success: boolean; message: string } | null>(null);
    const [contexts, setContexts] = useState<ContextInfo[]>([]);
    const [selectedContext, setSelectedContext] = useState("");
    const [editingContext, setEditingContext] = useState<string | null>(null);
    const [editDisplayName, setEditDisplayName] = useState("");

    useEffect(() => {
        fetchKubeConfig();
        fetchContexts();
    }, []);

    const fetchKubeConfig = async () => {
        setLoading(true);
        try {
            const data = await api.get<any>("/settings/kube");
            setKubeConfig(data.config || "");
        } catch (error) {
            console.error(error);
        } finally {
            setLoading(false);
        }
    };

    const fetchContexts = async () => {
        try {
            const data = await api.get<any>("/kube/contexts");
            setContexts(data.contexts || []);
            if (data.current) setSelectedContext(data.current);
            else if (data.contexts && data.contexts.length > 0) setSelectedContext(data.contexts[0].name);
        } catch (error) {
            console.error(error);
        }
    };

    const handleContextChange = async (value: string) => {
        setSelectedContext(value);
        try {
            await api.post("/settings/kube/context", { context: value });
        } catch (error) {
            console.error("Failed to update context:", error);
        }
    };

    const handleValidate = async () => {
        setIsValidating(true);
        setValidationStatus(null);
        try {
            const data = await api.post<any>("/settings/kube/validate");
            setValidationStatus({
                success: data.valid,
                message: data.valid ? data.message : (data.error || "Validation failed")
            });
        } catch (error) {
            setValidationStatus({ success: false, message: "Network error during validation" });
        } finally {
            setIsValidating(false);
        }
    };

    const handleStartEdit = (ctx: ContextInfo) => {
        setEditingContext(ctx.name);
        setEditDisplayName(ctx.display_name);
    };

    const handleSaveMapping = async (contextName: string) => {
        try {
            await api.post("/settings/context-mappings", { context_name: contextName, display_name: editDisplayName });
            setEditingContext(null);
            fetchContexts();
        } catch (error) {
            console.error(error);
        }
    };

    const handleCancelEdit = () => {
        setEditingContext(null);
        setEditDisplayName("");
    };

    return (
        <div className="flex flex-col items-center w-full min-h-screen bg-background/50 p-4 md:p-0">
            <div className="w-full max-w-5xl space-y-8">
                <div className="flex flex-col gap-1 mt-8">
                    <h1 className="text-3xl font-bold tracking-tight text-foreground">Cluster Settings</h1>
                    <p className="text-muted-foreground text-sm">Configure your Kubernetes clusters and authentication.</p>
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                    <div className="lg:col-span-1 space-y-4">
                        <KubeConfigUpload />
                    </div>

                    <div className="lg:col-span-2 space-y-8">
                        {/* Active Cluster Selector */}
                        <Card className="border-none shadow-2xl shadow-black/5 bg-card/50 backdrop-blur-sm overflow-hidden rounded-3xl">
                            <CardHeader className="border-b bg-card/50 px-8 py-6">
                                <CardTitle className="text-lg font-bold flex items-center gap-3">
                                    <CheckCircle2 className="h-5 w-5 text-green-500" />
                                    Active Cluster
                                </CardTitle>
                                <CardDescription>Select the Kubernetes cluster context to use for all operations.</CardDescription>
                            </CardHeader>
                            <CardContent className="p-8">
                                {contexts.length === 0 ? (
                                    <p className="text-muted-foreground text-sm text-center py-4">No contexts available. Configure a GitLab Agent first.</p>
                                ) : (
                                    <div className="space-y-4">
                                        <div className="space-y-2">
                                            <Label htmlFor="active-cluster" className="text-xs font-bold uppercase tracking-wider opacity-60">Current Context</Label>
                                            <Select value={selectedContext} onValueChange={handleContextChange}>
                                                <SelectTrigger className="h-12 text-sm rounded-xl bg-muted/30 border-muted/20">
                                                    <SelectValue placeholder="Select a cluster context" />
                                                </SelectTrigger>
                                                <SelectContent className="rounded-xl">
                                                    {contexts.map(c => (
                                                        <SelectItem key={c.name} value={c.name} className="rounded-lg text-sm py-3">
                                                            <div className="flex flex-col">
                                                                <span className="font-medium">{c.display_name}</span>
                                                                <span className="text-xs opacity-50 font-mono">{c.name}</span>
                                                            </div>
                                                        </SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </Select>
                                        </div>
                                        <p className="text-xs text-muted-foreground">
                                            This context will be used for all Kubernetes operations including pod management and deployments.
                                        </p>
                                    </div>
                                )}
                            </CardContent>
                        </Card>

                        <Card className="border-none shadow-2xl shadow-black/5 bg-card/50 backdrop-blur-sm overflow-hidden rounded-3xl">
                            <CardHeader className="border-b bg-card/50 px-8 py-6">
                                <CardTitle className="text-lg font-bold">Kubeconfig Viewer</CardTitle>
                                <CardDescription>Your current Kubernetes configuration (read-only).</CardDescription>
                            </CardHeader>
                            <CardContent className="p-8">
                                {loading ? (
                                    <div className="h-64 flex items-center justify-center">
                                        <p className="text-muted-foreground animate-pulse font-mono text-sm uppercase tracking-widest">Retrieving configuration...</p>
                                    </div>
                                ) : (
                                    <div className="space-y-6">
                                        <div className="space-y-3">
                                            <Label htmlFor="kubeconfig" className="text-xs font-bold uppercase tracking-wider opacity-60 ml-1">Config YAML (Read Only)</Label>
                                            <textarea
                                                id="kubeconfig"
                                                className="w-full h-48 p-6 font-mono text-xs border rounded-2xl bg-black/5 transition-all resize-none outline-none leading-relaxed cursor-not-allowed opacity-80"
                                                placeholder="No kubeconfig configured. Use GitLab Agent to generate one."
                                                value={kubeConfig}
                                                readOnly
                                            />
                                        </div>

                                        {validationStatus && (
                                            <div className={cn(
                                                "p-4 rounded-2xl border flex items-start gap-3 animate-in fade-in slide-in-from-top-2 duration-300",
                                                validationStatus.success ? "bg-green-500/5 border-green-500/20 text-green-700" : "bg-destructive/5 border-destructive/20 text-destructive"
                                            )}>
                                                {validationStatus.success ? <ShieldCheck className="h-5 w-5 mt-0.5 shrink-0" /> : <ShieldCheck className="h-5 w-5 mt-0.5 shrink-0 opacity-50" />}
                                                <div className="space-y-1">
                                                    <p className="text-[10px] font-bold uppercase tracking-wider">{validationStatus.success ? "Validation Successful" : "Validation Failed"}</p>
                                                    <p className="text-[11px] leading-relaxed opacity-80 break-words font-mono">{validationStatus.message}</p>
                                                </div>
                                            </div>
                                        )}

                                        <Button
                                            variant="outline"
                                            onClick={handleValidate}
                                            disabled={isValidating || !kubeConfig}
                                            className="w-full h-14 font-bold text-sm tracking-wide rounded-2xl transition-all active:scale-95 border-2"
                                        >
                                            {isValidating ? "Verifying..." : "Validate Connection"}
                                        </Button>
                                    </div>
                                )}
                            </CardContent>
                        </Card>

                        {/* Context Mapping Section */}
                        <Card className="border-none shadow-2xl shadow-black/5 bg-card/50 backdrop-blur-sm overflow-hidden rounded-3xl">
                            <CardHeader className="border-b bg-card/50 px-8 py-6">
                                <CardTitle className="text-lg font-bold flex items-center gap-3">
                                    <Server className="h-5 w-5 text-orange-500" />
                                    Context Display Names
                                </CardTitle>
                                <CardDescription>Customize how cluster contexts appear in dropdowns.</CardDescription>
                            </CardHeader>
                            <CardContent className="p-8">
                                {contexts.length === 0 ? (
                                    <p className="text-muted-foreground text-sm text-center py-8">No contexts available. Configure a GitLab Agent first.</p>
                                ) : (
                                    <div className="space-y-3">
                                        {contexts.map((ctx) => (
                                            <div key={ctx.name} className="flex items-center justify-between p-4 bg-muted/30 rounded-2xl border border-muted/20">
                                                <div className="flex-1">
                                                    <p className="text-[10px] font-bold uppercase tracking-widest opacity-50 mb-1">Original Context</p>
                                                    <p className="font-mono text-xs opacity-70">{ctx.name}</p>
                                                </div>
                                                {editingContext === ctx.name ? (
                                                    <div className="flex items-center gap-2 flex-1">
                                                        <Input
                                                            value={editDisplayName}
                                                            onChange={(e) => setEditDisplayName(e.target.value)}
                                                            className="h-10 text-sm rounded-xl"
                                                            placeholder="Enter display name"
                                                            autoFocus
                                                        />
                                                        <Button size="icon" variant="ghost" className="h-10 w-10 text-green-600 hover:bg-green-500/10 rounded-xl" onClick={() => handleSaveMapping(ctx.name)}>
                                                            <Check className="h-4 w-4" />
                                                        </Button>
                                                        <Button size="icon" variant="ghost" className="h-10 w-10 text-destructive hover:bg-destructive/10 rounded-xl" onClick={handleCancelEdit}>
                                                            <X className="h-4 w-4" />
                                                        </Button>
                                                    </div>
                                                ) : (
                                                    <div className="flex items-center gap-3 flex-1 justify-end">
                                                        <div className="text-right">
                                                            <p className="text-[10px] font-bold uppercase tracking-widest opacity-50 mb-1">Display Name</p>
                                                            <p className="font-bold text-sm">{ctx.display_name}</p>
                                                        </div>
                                                        <Button size="icon" variant="ghost" className="h-10 w-10 hover:bg-primary/10 rounded-xl" onClick={() => handleStartEdit(ctx)}>
                                                            <Edit2 className="h-4 w-4" />
                                                        </Button>
                                                    </div>
                                                )}
                                            </div>
                                        ))}
                                    </div>
                                )}
                            </CardContent>
                        </Card>
                    </div>
                </div>
            </div>
        </div>
    );
}
