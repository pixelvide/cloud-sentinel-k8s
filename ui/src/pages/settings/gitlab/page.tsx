"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import {
    Trash2,
    ShieldCheck,
    LayoutDashboard,
    Edit,
    CheckCircle2,
    AlertCircle,
    Plus,
    Globe,
    Key,
    Shield,
    ShieldAlert,
    Server,
    ChevronRight,
} from "lucide-react";
import { Link } from "react-router-dom";
import { cn } from "@/lib/utils";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
    DialogFooter,
} from "@/components/ui/dialog";
import { api } from "@/lib/api";

interface GitlabConfig {
    id: number;
    gitlab_host: string;
    token: string;
    is_https: boolean;
    is_validated: boolean;
}

interface GitlabK8sAgentConfig {
    id: number;
    gitlab_config_id: number;
    agent_id: string;
    agent_repo: string;
    is_configured: boolean;
    gitlab_config: GitlabConfig;
}

export default function SettingsPage() {
    const [configs, setConfigs] = useState<GitlabConfig[]>([]);
    const [agentConfigs, setAgentConfigs] = useState<GitlabK8sAgentConfig[]>([]);
    const [newHost, setNewHost] = useState("");
    const [newToken, setNewToken] = useState("");
    const [newIsHttps, setNewIsHttps] = useState(true);
    const [loading, setLoading] = useState(false);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [isAgentModalOpen, setIsAgentModalOpen] = useState(false);
    const [isValidatingId, setIsValidatingId] = useState<number | null>(null);
    const [configuringId, setConfiguringId] = useState<number | null>(null);
    const [validationStatuses, setValidationStatuses] = useState<Record<number, { success: boolean; message: string }>>(
        {}
    );
    const [agentConfigStatuses, setAgentConfigStatuses] = useState<
        Record<number, { success: boolean; message: string }>
    >({});

    const [selectedGitlabId, setSelectedGitlabId] = useState<string>("");
    const [agentId, setAgentId] = useState("");
    const [agentRepo, setAgentRepo] = useState("");
    const [isAgentLoading, setIsAgentLoading] = useState(false);

    useEffect(() => {
        fetchConfigs();
        fetchAgentConfigs();
    }, []);

    const fetchAgentConfigs = async () => {
        try {
            const data = await api.get<any>("/settings/gitlab/agents");
            setAgentConfigs(data.configs || []);
        } catch (error) {
            console.error(error);
        }
    };

    const fetchConfigs = async () => {
        try {
            const data = await api.get<any>("/settings/gitlab");
            setConfigs(data.configs || []);
        } catch (error) {
            console.error(error);
        }
    };

    const handleValidate = async (id: number) => {
        setIsValidatingId(id);
        try {
            const data = await api.post<any>(`/settings/gitlab/${id}/validate`);
            setValidationStatuses((prev) => ({
                ...prev,
                [id]: { success: true, message: data.message },
            }));
            fetchConfigs(); // Refresh to show updated validation status from DB
        } catch (error: any) {
            setValidationStatuses((prev) => ({
                ...prev,
                [id]: { success: false, message: error.message || "Validation failed" },
            }));
        } finally {
            setIsValidatingId(null);
        }
    };

    const handleAdd = async () => {
        if (!newHost || !newToken) return;
        setLoading(true);
        try {
            await api.post("/settings/gitlab", { gitlab_host: newHost, token: newToken, is_https: newIsHttps });
            setNewHost("");
            setNewToken("");
            setNewIsHttps(true);
            setIsModalOpen(false);
            fetchConfigs();
        } catch (error) {
            console.error(error);
        } finally {
            setLoading(false);
        }
    };

    const [editingId, setEditingId] = useState<number | null>(null);
    const [editToken, setEditToken] = useState("");

    const handleUpdate = async (id: number) => {
        if (!editToken) return;
        setLoading(true);
        try {
            await api.put(`/settings/gitlab/${id}`, { token: editToken });
            setEditingId(null);
            setEditToken("");
            fetchConfigs();
        } catch (error) {
            console.error(error);
        } finally {
            setLoading(false);
        }
    };

    const handleDelete = async (id: number) => {
        if (!confirm("Are you sure you want to delete this configuration?")) return;
        try {
            await api.del(`/settings/gitlab/${id}`);
            fetchConfigs();
        } catch (error) {
            console.error(error);
        }
    };

    const handleAddAgent = async () => {
        if (!selectedGitlabId || !agentId || !agentRepo) return;
        setIsAgentLoading(true);
        try {
            await api.post("/settings/gitlab/agents", {
                gitlab_config_id: parseInt(selectedGitlabId),
                agent_id: agentId,
                agent_repo: agentRepo,
            });
            setAgentId("");
            setAgentRepo("");
            setSelectedGitlabId("");
            setIsAgentModalOpen(false);
            fetchAgentConfigs();
        } catch (error) {
            console.error(error);
        } finally {
            setIsAgentLoading(false);
        }
    };

    const handleDeleteAgent = async (id: number) => {
        if (!confirm("Are you sure you want to delete this agent configuration?")) return;
        try {
            await api.del(`/settings/gitlab/agents/${id}`);
            fetchAgentConfigs();
        } catch (error) {
            console.error(error);
        }
    };

    const handleConfigureAgent = async (id: number) => {
        setConfiguringId(id);
        setAgentConfigStatuses((prev) => ({ ...prev, [id]: undefined as any })); // Clear previous status
        try {
            const data = await api.post<any>(`/settings/gitlab/agents/${id}/configure`);
            setAgentConfigStatuses((prev) => ({
                ...prev,
                [id]: { success: true, message: "Agent configured successfully! Use context: " + data.host },
            }));
            fetchAgentConfigs();
        } catch (error: any) {
            console.error(error);
            setAgentConfigStatuses((prev) => ({
                ...prev,
                [id]: { success: false, message: error.message || "An error occurred during configuration" },
            }));
        } finally {
            setConfiguringId(null);
        }
    };

    return (
        <div className="flex flex-col items-center w-full min-h-screen bg-background/50 p-4 md:p-0">
            <div className="w-full max-w-5xl space-y-8">
                <div className="flex flex-col md:flex-row md:items-end justify-between gap-6 mt-8">
                    <div className="flex flex-col gap-1">
                        <h1 className="text-3xl font-bold tracking-tight text-foreground">GitLab Settings</h1>
                        <p className="text-muted-foreground text-sm">
                            Configure your GitLab instances and credentials.
                        </p>
                    </div>

                    <Dialog open={isModalOpen} onOpenChange={setIsModalOpen}>
                        <DialogTrigger asChild>
                            <Button
                                size="lg"
                                className="h-14 px-8 rounded-2xl font-bold text-sm shadow-xl shadow-primary/20 transition-all active:scale-95 gap-3"
                            >
                                <Plus className="h-5 w-5" />
                                Add Instance
                            </Button>
                        </DialogTrigger>
                        <DialogContent className="sm:max-w-md border-none shadow-2xl bg-card rounded-[2.5rem] p-0 overflow-hidden">
                            <DialogHeader className="p-8 pb-4 bg-primary/5">
                                <DialogTitle className="text-xl font-bold tracking-tight">Connect GitLab</DialogTitle>
                                <DialogDescription className="text-sm">
                                    Link a new GitLab host to your private environment.
                                </DialogDescription>
                            </DialogHeader>
                            <div className="p-8 space-y-6">
                                <div className="space-y-4">
                                    <div className="space-y-2">
                                        <Label
                                            htmlFor="host"
                                            className="text-[10px] font-bold uppercase tracking-widest opacity-60 ml-1"
                                        >
                                            GitLab Host
                                        </Label>
                                        <div className="relative">
                                            <Globe className="absolute left-4 top-3.5 h-4 w-4 opacity-40" />
                                            <Input
                                                id="host"
                                                placeholder="gitlab.company.com"
                                                className="h-11 pl-11 border-muted bg-muted/30 focus:bg-background focus:ring-primary/20 transition-all rounded-xl"
                                                value={newHost}
                                                onChange={(e) => setNewHost(e.target.value)}
                                            />
                                        </div>
                                    </div>
                                    <div className="space-y-2">
                                        <Label
                                            htmlFor="token"
                                            className="text-[10px] font-bold uppercase tracking-widest opacity-60 ml-1"
                                        >
                                            Personal Access Token
                                        </Label>
                                        <div className="relative">
                                            <Key className="absolute left-4 top-3.5 h-4 w-4 opacity-40" />
                                            <Input
                                                id="token"
                                                type="password"
                                                placeholder="glpat-..."
                                                className="h-11 pl-11 border-muted bg-muted/30 focus:bg-background focus:ring-primary/20 transition-all rounded-xl"
                                                value={newToken}
                                                onChange={(e) => setNewToken(e.target.value)}
                                            />
                                        </div>
                                    </div>
                                    <div className="flex items-center justify-between p-4 bg-muted/20 rounded-2xl border border-muted/50">
                                        <div className="flex flex-col gap-0.5">
                                            <Label htmlFor="https" className="text-xs font-bold tracking-tight">
                                                SSL Security (HTTPS)
                                            </Label>
                                            <span className="text-[10px] text-muted-foreground opacity-60 italic">
                                                Always recommended
                                            </span>
                                        </div>
                                        <Checkbox
                                            id="https"
                                            checked={newIsHttps}
                                            className="h-5 w-5 rounded-lg border-2 shadow-sm"
                                            onCheckedChange={(c: boolean) => setNewIsHttps(c)}
                                        />
                                    </div>
                                </div>
                                <Button
                                    onClick={handleAdd}
                                    disabled={loading || !newHost || !newToken}
                                    className="w-full h-12 font-bold text-sm tracking-wide rounded-xl transition-all active:scale-95"
                                >
                                    {loading ? "Establishing Connection..." : "Authorize Instance"}
                                </Button>
                            </div>
                        </DialogContent>
                    </Dialog>
                </div>

                <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
                    <div className="lg:col-span-1 space-y-4">
                        <div className="p-6 rounded-2xl bg-primary/5 border border-primary/10 space-y-3 sticky top-8">
                            <h3 className="font-semibold text-primary flex items-center gap-2">
                                <ShieldCheck className="h-4 w-4" />
                                Global Status
                            </h3>
                            <div className="space-y-4 pt-1">
                                <div className="flex items-center justify-between">
                                    <span className="text-[11px] font-medium opacity-60 uppercase tracking-widest">
                                        Connected
                                    </span>
                                    <span className="text-sm font-bold font-mono">{configs.length}</span>
                                </div>
                                <div className="flex items-center justify-between">
                                    <span className="text-[11px] font-medium opacity-60 uppercase tracking-widest">
                                        Verified
                                    </span>
                                    <span className="text-sm font-bold font-mono text-green-500">
                                        {configs.filter((c) => c.is_validated).length}
                                    </span>
                                </div>
                                <div className="h-px bg-primary/10 w-full" />
                                <p className="text-[10px] leading-relaxed text-muted-foreground italic">
                                    Click "Validate" on any instance to verify connectivity and refresh its persistent
                                    status.
                                </p>
                            </div>
                        </div>
                    </div>

                    <div className="lg:col-span-3 space-y-6">
                        {configs.length === 0 && (
                            <div className="flex flex-col items-center justify-center p-20 border-2 border-dashed rounded-[3rem] border-muted/30 bg-card/10 text-center animate-in fade-in zoom-in duration-500">
                                <Shield className="h-12 w-12 text-muted-foreground/20 mb-4" />
                                <p className="text-muted-foreground font-medium italic mb-6">
                                    No GitLab instances found.
                                </p>
                                <Button
                                    variant="outline"
                                    onClick={() => setIsModalOpen(true)}
                                    className="rounded-xl font-bold uppercase tracking-widest text-[10px] border-2"
                                >
                                    Setup First Instance
                                </Button>
                            </div>
                        )}

                        <div className="grid grid-cols-1 gap-4">
                            {configs.map((config) => (
                                <div key={config.id} className="space-y-3">
                                    <div
                                        className={cn(
                                            "group relative flex flex-col md:flex-row md:items-center justify-between p-6 bg-card/60 backdrop-blur-md border rounded-[2rem] shadow-sm transition-all duration-300",
                                            editingId === config.id
                                                ? "border-primary ring-4 ring-primary/5"
                                                : "hover:shadow-xl hover:border-primary/30 hover:-translate-y-1"
                                        )}
                                    >
                                        <div className="flex items-center gap-5">
                                            <div
                                                className={cn(
                                                    "p-3.5 rounded-2xl transition-colors",
                                                    config.is_validated
                                                        ? "bg-green-500/10 text-green-500 shadow-sm shadow-green-500/10"
                                                        : "bg-muted/50 text-muted-foreground"
                                                )}
                                            >
                                                {config.is_validated ? (
                                                    <ShieldCheck className="h-6 w-6" />
                                                ) : (
                                                    <ShieldAlert className="h-6 w-6 opacity-40" />
                                                )}
                                            </div>
                                            <div className="min-w-0">
                                                <div className="flex items-center gap-3">
                                                    <span className="font-bold text-base tracking-tight truncate">
                                                        {config.gitlab_host}
                                                    </span>
                                                    <div className="flex items-center gap-1.5">
                                                        <span
                                                            className={cn(
                                                                "text-[9px] font-black uppercase tracking-widest px-2 py-0.5 rounded-md border shrink-0",
                                                                config.is_https
                                                                    ? "bg-green-500/5 border-green-500/20 text-green-600"
                                                                    : "bg-orange-500/5 border-orange-500/20 text-orange-600"
                                                            )}
                                                        >
                                                            {config.is_https ? "https" : "http"}
                                                        </span>
                                                        {config.is_validated && (
                                                            <span className="flex items-center gap-1 text-[9px] font-bold text-green-500 uppercase tracking-widest bg-green-500/5 px-2 py-0.5 rounded-md border border-green-500/20">
                                                                <CheckCircle2 className="h-2.5 w-2.5" />
                                                                Verified
                                                            </span>
                                                        )}
                                                    </div>
                                                </div>
                                                {editingId === config.id ? (
                                                    <div className="mt-3 flex flex-col sm:flex-row items-start sm:items-center gap-2">
                                                        <div className="relative w-full sm:w-80">
                                                            <Key className="absolute left-3 top-2.5 h-3.5 w-3.5 opacity-40" />
                                                            <Input
                                                                type="password"
                                                                placeholder="New Personal Access Token"
                                                                className="h-9 pl-9 text-xs rounded-lg bg-muted/50 border-primary/20"
                                                                value={editToken}
                                                                onChange={(e) => setEditToken(e.target.value)}
                                                                autoFocus
                                                            />
                                                        </div>
                                                        <div className="flex items-center gap-2 w-full sm:w-auto">
                                                            <Button
                                                                size="sm"
                                                                className="h-9 rounded-lg px-4 font-bold text-[10px] uppercase tracking-wider"
                                                                onClick={() => handleUpdate(config.id)}
                                                                disabled={loading || !editToken}
                                                            >
                                                                Update
                                                            </Button>
                                                            <Button
                                                                size="sm"
                                                                variant="ghost"
                                                                className="h-9 rounded-lg px-4 font-bold text-[10px] uppercase tracking-wider"
                                                                onClick={() => {
                                                                    setEditingId(null);
                                                                    setEditToken("");
                                                                }}
                                                            >
                                                                Cancel
                                                            </Button>
                                                        </div>
                                                    </div>
                                                ) : (
                                                    <div className="flex items-center gap-2 mt-1">
                                                        <p className="text-[11px] font-medium text-muted-foreground/50 tracking-widest font-mono">
                                                            TOKEN: ••••••••••••••••••
                                                        </p>
                                                    </div>
                                                )}
                                            </div>
                                        </div>

                                        <div className="flex items-center gap-2 mt-6 md:mt-0 md:opacity-0 md:group-hover:opacity-100 transition-all duration-300 md:translate-x-2 md:group-hover:translate-x-0">
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                className={cn(
                                                    "h-10 px-4 flex items-center gap-2 rounded-xl font-bold text-[10px] uppercase tracking-wider border-2 transition-all",
                                                    config.is_validated
                                                        ? "border-green-500/20 text-green-600 hover:bg-green-500/5"
                                                        : "border-primary/20 text-primary hover:bg-primary/5"
                                                )}
                                                onClick={() => handleValidate(config.id)}
                                                disabled={isValidatingId === config.id}
                                            >
                                                {isValidatingId === config.id
                                                    ? "Checking..."
                                                    : config.is_validated
                                                      ? "Re-validate"
                                                      : "Validate"}
                                            </Button>
                                            <div className="w-px h-6 bg-muted/60 mx-1" />
                                            <Button
                                                variant="ghost"
                                                size="icon"
                                                className="h-10 w-10 hover:bg-primary/10 hover:text-primary rounded-xl"
                                                onClick={() => {
                                                    setEditingId(config.id);
                                                    setEditToken("");
                                                }}
                                            >
                                                <Edit className="h-4 w-4" />
                                            </Button>
                                            <Button
                                                variant="ghost"
                                                size="icon"
                                                className="h-10 w-10 hover:bg-destructive/10 hover:text-destructive rounded-xl"
                                                onClick={() => handleDelete(config.id)}
                                            >
                                                <Trash2 className="h-4 w-4" />
                                            </Button>
                                        </div>
                                    </div>

                                    {validationStatuses[config.id] && (
                                        <div
                                            className={cn(
                                                "p-4 rounded-2xl border flex items-start gap-4 animate-in fade-in slide-in-from-top-4 duration-500 mx-4",
                                                validationStatuses[config.id].success
                                                    ? "bg-green-500/5 border-green-500/20 text-green-700"
                                                    : "bg-destructive/5 border-destructive/20 text-destructive"
                                            )}
                                        >
                                            {validationStatuses[config.id].success ? (
                                                <CheckCircle2 className="h-5 w-5 mt-1 shrink-0 animate-bounce-slow" />
                                            ) : (
                                                <AlertCircle className="h-5 w-5 mt-1 shrink-0" />
                                            )}
                                            <div className="space-y-1">
                                                <p className="text-[10px] font-black uppercase tracking-[0.2em]">
                                                    {validationStatuses[config.id].success
                                                        ? "Verified Successfully"
                                                        : "Validation Failed"}
                                                </p>
                                                <p className="text-[11px] leading-relaxed opacity-90 break-words font-mono font-medium">
                                                    {validationStatuses[config.id].message}
                                                </p>
                                            </div>
                                        </div>
                                    )}
                                </div>
                            ))}
                        </div>

                        {/* GitLab Agent Configuration Section */}
                        <div className="pt-12 space-y-6">
                            <div className="flex flex-col md:flex-row md:items-end justify-between gap-6">
                                <div className="flex flex-col gap-1">
                                    <h2 className="text-2xl font-bold tracking-tight text-foreground flex items-center gap-3">
                                        <div className="p-2 rounded-xl bg-orange-500/10 text-orange-500">
                                            <Server className="h-5 w-5" />
                                        </div>
                                        GitLab Agents
                                    </h2>
                                    <p className="text-muted-foreground text-sm pl-11">
                                        Connect your Kubernetes clusters using GitLab Agents.
                                    </p>
                                </div>

                                <Dialog open={isAgentModalOpen} onOpenChange={setIsAgentModalOpen}>
                                    <DialogTrigger asChild>
                                        <Button
                                            size="lg"
                                            className="h-12 px-6 rounded-xl font-bold text-sm shadow-lg shadow-orange-500/10 transition-all active:scale-95 gap-3 bg-orange-500 hover:bg-orange-600 text-white border-none"
                                        >
                                            <Plus className="h-4 w-4" />
                                            Register Agent
                                        </Button>
                                    </DialogTrigger>
                                    <DialogContent className="sm:max-w-xl border-none shadow-2xl bg-card rounded-[2.5rem] p-0 overflow-hidden">
                                        <DialogHeader className="p-8 pb-4 bg-orange-500/5">
                                            <DialogTitle className="text-xl font-bold tracking-tight text-orange-600">
                                                Register GitLab Agent
                                            </DialogTitle>
                                            <DialogDescription className="text-sm">
                                                Link a new Kubernetes agent from your GitLab projects.
                                            </DialogDescription>
                                        </DialogHeader>
                                        <div className="p-8 space-y-6">
                                            <div className="grid grid-cols-1 gap-6">
                                                <div className="space-y-2">
                                                    <Label className="text-[10px] font-bold uppercase tracking-widest opacity-60 ml-1">
                                                        GitLab Instance
                                                    </Label>
                                                    <Select
                                                        value={selectedGitlabId}
                                                        onValueChange={setSelectedGitlabId}
                                                    >
                                                        <SelectTrigger className="h-12 border-muted bg-muted/30 rounded-xl focus:ring-orange-500/20">
                                                            <SelectValue placeholder="Select instance" />
                                                        </SelectTrigger>
                                                        <SelectContent className="rounded-xl border-none shadow-2xl">
                                                            {configs
                                                                .filter((c) => c.is_validated)
                                                                .map((config) => (
                                                                    <SelectItem
                                                                        key={config.id}
                                                                        value={config.id.toString()}
                                                                        className="rounded-lg"
                                                                    >
                                                                        {config.gitlab_host}
                                                                    </SelectItem>
                                                                ))}
                                                            {configs.filter((c) => c.is_validated).length === 0 && (
                                                                <div className="p-4 text-[11px] text-muted-foreground italic">
                                                                    No verified instances available.
                                                                </div>
                                                            )}
                                                        </SelectContent>
                                                    </Select>
                                                </div>
                                                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                                    <div className="space-y-2">
                                                        <Label
                                                            htmlFor="agent_id"
                                                            className="text-[10px] font-bold uppercase tracking-widest opacity-60 ml-1"
                                                        >
                                                            Agent ID
                                                        </Label>
                                                        <Input
                                                            id="agent_id"
                                                            placeholder="e.g. 12345"
                                                            className="h-12 border-muted bg-muted/30 rounded-xl focus:ring-orange-500/20"
                                                            value={agentId}
                                                            onChange={(e) => setAgentId(e.target.value)}
                                                        />
                                                    </div>
                                                    <div className="space-y-2">
                                                        <Label
                                                            htmlFor="agent_repo"
                                                            className="text-[10px] font-bold uppercase tracking-widest opacity-60 ml-1"
                                                        >
                                                            Agent Repository
                                                        </Label>
                                                        <Input
                                                            id="agent_repo"
                                                            placeholder="path/to/agent-config"
                                                            className="h-12 border-muted bg-muted/30 rounded-xl focus:ring-orange-500/20"
                                                            value={agentRepo}
                                                            onChange={(e) => setAgentRepo(e.target.value)}
                                                        />
                                                    </div>
                                                </div>
                                            </div>
                                            <Button
                                                onClick={handleAddAgent}
                                                disabled={isAgentLoading || !selectedGitlabId || !agentId || !agentRepo}
                                                className="w-full h-12 font-bold text-sm tracking-wide rounded-xl transition-all active:scale-95 gap-3 bg-orange-500 hover:bg-orange-600 text-white"
                                            >
                                                {isAgentLoading ? (
                                                    "Connecting..."
                                                ) : (
                                                    <>
                                                        <Server className="h-4 w-4" />
                                                        Initialize Connection
                                                    </>
                                                )}
                                            </Button>
                                        </div>
                                    </DialogContent>
                                </Dialog>
                            </div>

                            <div className="grid grid-cols-1 gap-4">
                                {agentConfigs.map((agent) => (
                                    <div key={agent.id} className="space-y-4">
                                        <div className="group flex flex-col md:flex-row md:items-center justify-between gap-4 p-6 bg-card/40 backdrop-blur-md border border-muted/20 rounded-[2rem] hover:border-primary/20 transition-all duration-300">
                                            <div className="flex items-center gap-5">
                                                <div
                                                    className={cn(
                                                        "p-3 rounded-2xl transition-all duration-300",
                                                        agent.is_configured
                                                            ? "bg-green-500/10 text-green-500 shadow-sm shadow-green-500/5"
                                                            : "bg-primary/5 text-primary"
                                                    )}
                                                >
                                                    <Server className="h-5 w-5" />
                                                </div>
                                                <div className="space-y-1">
                                                    <div className="flex items-center gap-2 flex-wrap">
                                                        <span className="text-xs font-bold opacity-60 uppercase tracking-widest">
                                                            {agent.gitlab_config.gitlab_host}
                                                        </span>
                                                        <ChevronRight className="h-3 w-3 opacity-20" />
                                                        <span className="font-bold text-sm tracking-tight">
                                                            Agent Id: {agent.agent_id}
                                                        </span>
                                                        {agent.is_configured && (
                                                            <span className="flex items-center gap-1 text-[9px] font-black uppercase tracking-widest bg-green-500/5 text-green-600 px-2 py-0.5 rounded-md border border-green-500/10 ml-2">
                                                                <CheckCircle2 className="h-2.5 w-2.5" />
                                                                Configured
                                                            </span>
                                                        )}
                                                    </div>
                                                    <p className="text-[11px] font-medium text-muted-foreground/70 font-mono tracking-tight break-all">
                                                        {agent.agent_repo}
                                                    </p>
                                                </div>
                                            </div>
                                            <div className="flex items-center gap-2 md:opacity-0 md:group-hover:opacity-100 transition-all">
                                                <Button
                                                    variant="outline"
                                                    size="sm"
                                                    className={cn(
                                                        "h-10 px-4 rounded-xl font-bold text-[10px] uppercase tracking-wider border-2 transition-all",
                                                        agent.is_configured
                                                            ? "border-green-500/20 text-green-600 hover:bg-green-500/5"
                                                            : "border-primary/20 text-primary hover:bg-primary/5"
                                                    )}
                                                    onClick={() => handleConfigureAgent(agent.id)}
                                                    disabled={configuringId === agent.id}
                                                >
                                                    {configuringId === agent.id
                                                        ? "Configuring..."
                                                        : agent.is_configured
                                                          ? "Re-configure"
                                                          : "Configure"}
                                                </Button>
                                                <div className="w-px h-6 bg-muted/60 mx-1" />
                                                <Button
                                                    variant="ghost"
                                                    size="icon"
                                                    className="h-10 w-10 hover:bg-destructive/10 hover:text-destructive rounded-xl transition-all"
                                                    onClick={() => handleDeleteAgent(agent.id)}
                                                >
                                                    <Trash2 className="h-4 w-4" />
                                                </Button>
                                            </div>
                                        </div>

                                        {agentConfigStatuses[agent.id] && (
                                            <div
                                                className={cn(
                                                    "p-4 rounded-2xl border flex items-start gap-4 animate-in fade-in slide-in-from-top-4 duration-500 mx-4",
                                                    agentConfigStatuses[agent.id].success
                                                        ? "bg-green-500/5 border-green-500/20 text-green-700"
                                                        : "bg-destructive/5 border-destructive/20 text-destructive"
                                                )}
                                            >
                                                {agentConfigStatuses[agent.id].success ? (
                                                    <CheckCircle2 className="h-5 w-5 mt-1 shrink-0 animate-bounce-slow" />
                                                ) : (
                                                    <AlertCircle className="h-5 w-5 mt-1 shrink-0" />
                                                )}
                                                <div className="space-y-1">
                                                    <p className="text-[10px] font-black uppercase tracking-[0.2em]">
                                                        {agentConfigStatuses[agent.id].success
                                                            ? "Configuration Successful"
                                                            : "Configuration Failed"}
                                                    </p>
                                                    <p className="text-[11px] leading-relaxed opacity-90 break-words font-mono font-medium">
                                                        {agentConfigStatuses[agent.id].message}
                                                    </p>
                                                </div>
                                            </div>
                                        )}
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}
