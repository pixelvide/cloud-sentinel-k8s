"use client";

import Link from "next/link";
import { usePathname, useSearchParams } from "next/navigation";
import { useEffect, useState, Suspense } from "react";
import { Home, Settings, LogOut, LayoutDashboard, X, Cloud, ChevronDown, Layers, HardDrive, Box, Grid, Globe, Network, PlayCircle, Clock, Boxes, AlertCircle, Server, Database } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface UserProfile {
    id: number;
    email: string;
    name: string;
}

function SidebarContent({ isOpen, onClose }: { isOpen?: boolean, onClose?: () => void }) {
    const pathname = usePathname();
    const searchParams = useSearchParams();
    const currentContext = searchParams.get("context");

    const [user, setUser] = useState<UserProfile | null>(null);
    const [settingsOpen, setSettingsOpen] = useState(false);
    const [networkOpen, setNetworkOpen] = useState(false);
    const [workloadsOpen, setWorkloadsOpen] = useState(false);

    useEffect(() => {
        const fetchUser = async () => {
            try {
                const res = await fetch("/api/v1/me", { credentials: "include" });
                if (res.ok) {
                    const data = await res.json();
                    setUser(data);
                }
            } catch (err) {
                console.error("Failed to fetch user profile:", err);
            }
        };
        fetchUser();
    }, []);



    // Don't show sidebar on login or exec pages
    if (pathname === "/login" || pathname?.startsWith("/exec")) return null;

    const isActive = (path: string) => pathname === path;

    const getLinkHref = (path: string) => {
        const params = new URLSearchParams();
        if (currentContext) params.set("context", currentContext);

        const currentNamespace = searchParams.get("namespace");
        if (currentNamespace) params.set("namespace", currentNamespace);

        const paramString = params.toString();
        return paramString ? `${path}?${paramString}` : path;
    };

    const handleLogout = () => {
        window.location.href = "/api/v1/auth/logout";
    };

    const getInitials = (name: string) => {
        if (!name) return "??";
        return name.split(" ").map(n => n[0]).join("").toUpperCase().substring(0, 2);
    };

    return (
        <>
            {/* Mobile Overlay */}
            {isOpen && (
                <div
                    className="fixed inset-0 bg-background/80 backdrop-blur-sm z-40 lg:hidden"
                    onClick={onClose}
                />
            )}

            <div className={cn(
                "fixed inset-y-0 left-0 w-64 h-screen bg-sidebar text-sidebar-foreground flex flex-col shadow-xl z-50 transition-transform duration-300 lg:relative lg:translate-x-0 lg:z-20 lg:border-r",
                isOpen ? "translate-x-0" : "-translate-x-full"
            )}>
                <div className="p-8 pb-4 flex flex-col gap-4">
                    <div className="flex items-center justify-between gap-3">
                        <div className="flex items-center gap-3">
                            <div className="p-2 bg-primary rounded-xl shadow-lg shadow-primary/20">
                                <LayoutDashboard className="h-6 w-6 text-primary-foreground" />
                            </div>
                            <div className="flex flex-col">
                                <span className="font-bold text-lg tracking-tight leading-none text-sidebar-foreground">Cloud Sentinel</span>
                                <span className="text-sm font-medium opacity-60">K8s</span>
                            </div>
                        </div>
                        {/* Mobile Close Button */}
                        <Button variant="ghost" size="icon" className="lg:hidden text-sidebar-foreground/50" onClick={onClose}>
                            <X className="h-5 w-5" />
                        </Button>
                    </div>
                </div>

                <nav className="flex-1 px-3 space-y-1.5 overflow-y-auto">
                    <Link href={getLinkHref("/")} className="block" onClick={onClose}>
                        <Button
                            variant={isActive("/") ? "secondary" : "ghost"}
                            className={cn(
                                "w-full justify-start gap-3 h-11 px-4 transition-all duration-200",
                                isActive("/") ? "bg-sidebar-accent text-white shadow-sm" : "hover:bg-white/10 hover:text-white"
                            )}
                        >
                            <Home className={cn("h-4 w-4", isActive("/") ? "text-primary" : "opacity-60")} />
                            <span className="font-medium text-sm">Dashboard</span>
                        </Button>
                    </Link>
                    <Link href={getLinkHref("/namespaces")} className="block" onClick={onClose}>
                        <Button
                            variant={isActive("/namespaces") ? "secondary" : "ghost"}
                            className={cn(
                                "w-full justify-start gap-3 h-11 px-4 transition-all duration-200",
                                isActive("/namespaces") ? "bg-sidebar-accent text-white shadow-sm" : "hover:bg-white/10 hover:text-white"
                            )}
                        >
                            <Layers className={cn("h-4 w-4", isActive("/namespaces") ? "text-primary" : "opacity-60")} />
                            <span className="font-medium text-sm">Namespaces</span>
                        </Button>
                    </Link>
                    <Link href={getLinkHref("/nodes")} className="block" onClick={onClose}>
                        <Button
                            variant={isActive("/nodes") ? "secondary" : "ghost"}
                            className={cn(
                                "w-full justify-start gap-3 h-11 px-4 transition-all duration-200",
                                isActive("/nodes") ? "bg-sidebar-accent text-white shadow-sm" : "hover:bg-white/10 hover:text-white"
                            )}
                        >
                            <HardDrive className={cn("h-4 w-4", isActive("/nodes") ? "text-primary" : "opacity-60")} />
                            <span className="font-medium text-sm">Nodes</span>
                        </Button>
                    </Link>
                    <Link href={getLinkHref("/events")} className="block" onClick={onClose}>
                        <Button
                            variant={isActive("/events") ? "secondary" : "ghost"}
                            className={cn(
                                "w-full justify-start gap-3 h-11 px-4 transition-all duration-200",
                                isActive("/events") ? "bg-sidebar-accent text-white shadow-sm" : "hover:bg-white/10 hover:text-white"
                            )}
                        >
                            <AlertCircle className={cn("h-4 w-4", isActive("/events") ? "text-primary" : "opacity-60")} />
                            <span className="font-medium text-sm">Events</span>
                        </Button>
                    </Link>
                    <Button
                        variant="ghost"
                        className="w-full justify-between gap-3 h-11 px-4 mt-1 hover:bg-white/10 hover:text-white"
                        onClick={() => setWorkloadsOpen(!workloadsOpen)}
                    >
                        <div className="flex items-center gap-3">
                            <Boxes className="h-4 w-4 opacity-60" />
                            <span className="font-medium text-sm">Workloads</span>
                        </div>
                        <ChevronDown className={cn("h-4 w-4 opacity-40 transition-transform", workloadsOpen && "rotate-180")} />
                    </Button>
                    {workloadsOpen && (
                        <div className="ml-4 space-y-1">
                            <Link href={getLinkHref("/pods")} className="block" onClick={onClose}>
                                <Button
                                    variant={isActive("/pods") ? "secondary" : "ghost"}
                                    className={cn(
                                        "w-full justify-start gap-3 h-10 px-4 transition-all duration-200",
                                        isActive("/pods") ? "bg-sidebar-accent text-white shadow-sm" : "hover:bg-white/10 hover:text-white"
                                    )}
                                >
                                    <Box className={cn("h-4 w-4", isActive("/pods") ? "text-primary" : "opacity-60")} />
                                    <span className="font-medium text-sm">Pods</span>
                                </Button>
                            </Link>
                            <Link href={getLinkHref("/deployments")} className="block" onClick={onClose}>
                                <Button
                                    variant={isActive("/deployments") ? "secondary" : "ghost"}
                                    className={cn(
                                        "w-full justify-start gap-3 h-10 px-4 transition-all duration-200",
                                        isActive("/deployments") ? "bg-sidebar-accent text-white shadow-sm" : "hover:bg-white/10 hover:text-white"
                                    )}
                                >
                                    <Layers className={cn("h-4 w-4", isActive("/deployments") ? "text-primary" : "opacity-60")} />
                                    <span className="font-medium text-sm">Deployments</span>
                                </Button>
                            </Link>
                            <Link href={getLinkHref("/replicasets")} className="block" onClick={onClose}>
                                <Button
                                    variant={isActive("/replicasets") ? "secondary" : "ghost"}
                                    className={cn(
                                        "w-full justify-start gap-3 h-10 px-4 transition-all duration-200",
                                        isActive("/replicasets") ? "bg-sidebar-accent text-white shadow-sm" : "hover:bg-white/10 hover:text-white"
                                    )}
                                >
                                    <Layers className={cn("h-4 w-4", isActive("/replicasets") ? "text-primary" : "opacity-60")} />
                                    <span className="font-medium text-sm">ReplicaSets</span>
                                </Button>
                            </Link>
                            <Link href={getLinkHref("/jobs")} className="block" onClick={onClose}>
                                <Button
                                    variant={isActive("/jobs") ? "secondary" : "ghost"}
                                    className={cn(
                                        "w-full justify-start gap-3 h-10 px-4 transition-all duration-200",
                                        isActive("/jobs") ? "bg-sidebar-accent text-white shadow-sm" : "hover:bg-white/10 hover:text-white"
                                    )}
                                >
                                    <PlayCircle className={cn("h-4 w-4", isActive("/jobs") ? "text-primary" : "opacity-60")} />
                                    <span className="font-medium text-sm">Jobs</span>
                                </Button>
                            </Link>
                            <Link href={getLinkHref("/cronjobs")} className="block" onClick={onClose}>
                                <Button
                                    variant={isActive("/cronjobs") ? "secondary" : "ghost"}
                                    className={cn(
                                        "w-full justify-start gap-3 h-10 px-4 transition-all duration-200",
                                        isActive("/cronjobs") ? "bg-sidebar-accent text-white shadow-sm" : "hover:bg-white/10 hover:text-white"
                                    )}
                                >
                                    <Clock className={cn("h-4 w-4", isActive("/cronjobs") ? "text-primary" : "opacity-60")} />
                                    <span className="font-medium text-sm">CronJobs</span>
                                </Button>
                            </Link>
                            <Link href={getLinkHref("/daemonsets")} className="block" onClick={onClose}>
                                <Button
                                    variant={isActive("/daemonsets") ? "secondary" : "ghost"}
                                    className={cn(
                                        "w-full justify-start gap-3 h-10 px-4 transition-all duration-200",
                                        isActive("/daemonsets") ? "bg-sidebar-accent text-white shadow-sm" : "hover:bg-white/10 hover:text-white"
                                    )}
                                >
                                    <Server className={cn("h-4 w-4", isActive("/daemonsets") ? "text-primary" : "opacity-60")} />
                                    <span className="font-medium text-sm">DaemonSets</span>
                                </Button>
                            </Link>
                            <Link href={getLinkHref("/statefulsets")} className="block" onClick={onClose}>
                                <Button
                                    variant={isActive("/statefulsets") ? "secondary" : "ghost"}
                                    className={cn(
                                        "w-full justify-start gap-3 h-10 px-4 transition-all duration-200",
                                        isActive("/statefulsets") ? "bg-sidebar-accent text-white shadow-sm" : "hover:bg-white/10 hover:text-white"
                                    )}
                                >
                                    <Database className={cn("h-4 w-4", isActive("/statefulsets") ? "text-primary" : "opacity-60")} />
                                    <span className="font-medium text-sm">StatefulSets</span>
                                </Button>
                            </Link>
                        </div>

                    )}
                    <Button
                        variant="ghost"
                        className="w-full justify-between gap-3 h-11 px-4 mt-1 hover:bg-white/10 hover:text-white"
                        onClick={() => setNetworkOpen(!networkOpen)}
                    >
                        <div className="flex items-center gap-3">
                            <Network className="h-4 w-4 opacity-60" />
                            <span className="font-medium text-sm">Network</span>
                        </div>
                        <ChevronDown className={cn("h-4 w-4 opacity-40 transition-transform", networkOpen && "rotate-180")} />
                    </Button>
                    {networkOpen && (
                        <div className="ml-4 space-y-1">
                            <Link href={getLinkHref("/services")} className="block" onClick={onClose}>
                                <Button
                                    variant={isActive("/services") ? "secondary" : "ghost"}
                                    className={cn(
                                        "w-full justify-start gap-3 h-10 px-4 transition-all duration-200",
                                        isActive("/services") ? "bg-sidebar-accent text-white shadow-sm" : "hover:bg-white/10 hover:text-white"
                                    )}
                                >
                                    <Grid className={cn("h-4 w-4", isActive("/services") ? "text-primary" : "opacity-60")} />
                                    <span className="font-medium text-sm">Services</span>
                                </Button>
                            </Link>
                            <Link href={getLinkHref("/ingresses")} className="block" onClick={onClose}>
                                <Button
                                    variant={isActive("/ingresses") ? "secondary" : "ghost"}
                                    className={cn(
                                        "w-full justify-start gap-3 h-10 px-4 transition-all duration-200",
                                        isActive("/ingresses") ? "bg-sidebar-accent text-white shadow-sm" : "hover:bg-white/10 hover:text-white"
                                    )}
                                >
                                    <Globe className={cn("h-4 w-4", isActive("/ingresses") ? "text-primary" : "opacity-60")} />
                                    <span className="font-medium text-sm">Ingresses</span>
                                </Button>
                            </Link>
                        </div>
                    )}
                    <Button
                        variant="ghost"
                        className="w-full justify-between gap-3 h-11 px-4 mt-4 hover:bg-white/10 hover:text-white"
                        onClick={() => setSettingsOpen(!settingsOpen)}
                    >
                        <div className="flex items-center gap-3">
                            <Settings className="h-4 w-4 opacity-60" />
                            <span className="font-medium text-sm">Settings</span>
                        </div>
                        <ChevronDown className={cn("h-4 w-4 opacity-40 transition-transform", settingsOpen && "rotate-180")} />
                    </Button>
                    {settingsOpen && (
                        <div className="ml-4 space-y-1">
                            <Link href="/settings/gitlab" className="block" onClick={onClose}>
                                <Button
                                    variant={isActive("/settings/gitlab") ? "secondary" : "ghost"}
                                    className={cn(
                                        "w-full justify-start gap-3 h-10 px-4 transition-all duration-200",
                                        isActive("/settings/gitlab") ? "bg-sidebar-accent text-white shadow-sm" : "hover:bg-white/10 hover:text-white"
                                    )}
                                >
                                    <Settings className={cn("h-4 w-4", isActive("/settings/gitlab") ? "text-primary" : "opacity-60")} />
                                    <span className="font-medium text-sm">GitLab Settings</span>
                                </Button>
                            </Link>
                            <Link href="/settings/clusters" className="block" onClick={onClose}>
                                <Button
                                    variant={isActive("/settings/clusters") ? "secondary" : "ghost"}
                                    className={cn(
                                        "w-full justify-start gap-3 h-10 px-4 transition-all duration-200",
                                        isActive("/settings/clusters") ? "bg-sidebar-accent text-white shadow-sm" : "hover:bg-white/10 hover:text-white"
                                    )}
                                >
                                    <Cloud className={cn("h-4 w-4", isActive("/settings/clusters") ? "text-primary" : "opacity-60")} />
                                    <span className="font-medium text-sm">Cluster Settings</span>
                                </Button>
                            </Link>
                        </div>
                    )}
                </nav>

                <div className="p-4 mt-auto">
                    <div className="p-4 rounded-2xl bg-white/5 border border-white/5 flex flex-col gap-3">
                        <div className="flex items-center gap-3">
                            <div className="h-8 w-8 rounded-full bg-primary/20 flex items-center justify-center border border-primary/20">
                                <span className="text-xs font-bold text-primary">{user ? getInitials(user.name) : "..."}</span>
                            </div>
                            <div className="flex flex-col overflow-hidden">
                                <span className="text-xs font-semibold truncate">{user?.name || "Loading..."}</span>
                                <span className="text-[10px] opacity-50 truncate">{user?.email || "Account details"}</span>
                            </div>
                        </div>
                        <Button
                            variant="ghost"
                            className="w-full justify-start gap-3 h-10 px-3 text-destructive hover:text-white hover:bg-destructive/90 transition-all rounded-xl border border-transparent hover:border-destructive/20"
                            onClick={handleLogout}
                        >
                            <LogOut className="h-4 w-4" />
                            <span className="text-xs font-semibold uppercase tracking-wider">Logout</span>
                        </Button>
                    </div>
                </div>
            </div>
        </>
    );
}

export function Sidebar(props: { isOpen?: boolean, onClose?: () => void }) {
    return (
        <Suspense fallback={<div className="w-64 h-screen bg-sidebar sticky top-0" />}>
            <SidebarContent {...props} />
        </Suspense>
    );
}
