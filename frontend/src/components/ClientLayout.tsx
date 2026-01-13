"use client";

import { Sidebar } from "@/components/Sidebar";
import { usePathname, useRouter } from "next/navigation";
import { useState, useEffect } from "react";
import { Menu, LayoutDashboard, Box, Grid, Globe, HardDrive, Layers, PlayCircle, Clock, Boxes, AlertCircle, RefreshCw, Server, Database } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ClusterContextSelector } from "@/components/ClusterContextSelector";

// Page Configuration
const PAGE_CONFIG: Record<string, { title: string; description: string; icon: any }> = {
    "/": {
        title: "Dashboard",
        description: "Cluster overview and health status",
        icon: LayoutDashboard
    },
    "/pods": {
        title: "Pods",
        description: "Manage workload instances",
        icon: Box
    },
    "/deployments": {
        title: "Deployments",
        description: "Manage application deployments",
        icon: Boxes
    },
    "/jobs": {
        title: "Jobs",
        description: "Manage batch jobs",
        icon: PlayCircle
    },
    "/cronjobs": {
        title: "CronJobs",
        description: "Manage scheduled jobs",
        icon: Clock
    },
    "/services": {
        title: "Services",
        description: "Manage networking endpoints",
        icon: Grid
    },
    "/ingresses": {
        title: "Ingresses",
        description: "Manage external access",
        icon: Globe
    },
    "/nodes": {
        title: "Nodes",
        description: "Cluster nodes and capacity",
        icon: HardDrive
    },
    "/namespaces": {
        title: "Namespaces",
        description: "Manage cluster namespaces",
        icon: Layers
    },
    "/events": {
        title: "Events",
        description: "Cluster events and alerts",
        icon: AlertCircle
    },
    "/daemonsets": {
        title: "DaemonSets",
        description: "Manage daemon set workloads",
        icon: Server
    },
    "/statefulsets": {
        title: "StatefulSets",
        description: "Manage stateful applications",
        icon: Database
    },
    "/replicasets": {
        title: "ReplicaSets",
        description: "Manage replica set workloads",
        icon: Boxes
    },
};

export function ClientLayout({ children }: { children: React.ReactNode }) {
    const pathname = usePathname();
    const router = useRouter();
    const [isSidebarOpen, setIsSidebarOpen] = useState(false);
    const [isAuthChecking, setIsAuthChecking] = useState(true);
    const isLoginPage = pathname === "/login";
    const isExecPage = pathname?.startsWith("/exec");

    // Auth check for protected pages
    useEffect(() => {
        // Skip auth check for login and exec pages
        if (isLoginPage || isExecPage) {
            setIsAuthChecking(false);
            return;
        }

        const checkAuth = async () => {
            try {
                const res = await fetch("/api/v1/me", { credentials: "include" });
                if (res.status === 401) {
                    router.push("/login");
                    return;
                }
            } catch (err) {
                console.error("Auth check failed:", err);
                router.push("/login");
                return;
            }
            setIsAuthChecking(false);
        };
        checkAuth();
    }, [pathname, isLoginPage, isExecPage, router]);

    // Close sidebar on path change
    useEffect(() => {
        setIsSidebarOpen(false);
    }, [pathname]);

    if (isLoginPage) {
        return <main className="min-h-screen w-full">{children}</main>;
    }

    // Show loading while checking auth
    if (isAuthChecking) {
        return (
            <div className="flex items-center justify-center min-h-screen">
                <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
            </div>
        );
    }

    const isContextPage = Object.keys(PAGE_CONFIG).includes(pathname);
    const currentPage = PAGE_CONFIG[pathname];

    return (
        <div className="flex min-h-screen md:h-screen md:overflow-hidden bg-background">
            {!isExecPage && <Sidebar isOpen={isSidebarOpen} onClose={() => setIsSidebarOpen(false)} />}

            <div className="flex-1 flex flex-col min-w-0">
                {/* Mobile Sidebar Toggle Header */}
                {!isExecPage && (
                    <header className="lg:hidden flex items-center justify-between p-4 border-b bg-sidebar text-sidebar-foreground z-30">
                        <div className="flex items-center gap-2">
                            <LayoutDashboard className="h-5 w-5 text-primary" />
                            <span className="font-bold text-sm tracking-tight">Cloud K8s</span>
                        </div>
                        <Button variant="ghost" size="icon" onClick={() => setIsSidebarOpen(true)}>
                            <Menu className="h-5 w-5" />
                        </Button>
                    </header>
                )}

                {/* Main Header with Global Context Selector (Visible on Context Pages) */}
                {isContextPage && currentPage && (
                    <header className="flex flex-col md:flex-row md:items-center justify-between gap-4 md:gap-0 px-4 py-3 md:px-8 md:py-4 border-b bg-background/50 backdrop-blur-sm z-20">
                        <div className="flex items-center gap-3">
                            <currentPage.icon className="h-5 w-5 text-primary mt-1" />
                            <div className="flex flex-col">
                                <h1 className="text-lg font-semibold tracking-tight leading-none">
                                    {currentPage.title}
                                </h1>
                                <p className="text-xs text-muted-foreground mt-1 hidden md:block">
                                    {currentPage.description}
                                </p>
                            </div>
                        </div>
                        <ClusterContextSelector />
                    </header>
                )}

                <main className={`flex-1 md:overflow-y-auto overflow-x-hidden transition-all duration-500 ${pathname?.startsWith('/exec') ? 'p-0' : 'p-6 md:p-8 lg:p-12'}`}>
                    {children}
                </main>
            </div>
        </div>
    );
}
