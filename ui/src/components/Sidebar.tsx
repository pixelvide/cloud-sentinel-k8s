import { Link, useLocation, useSearchParams } from "react-router-dom";
import { useEffect, useState } from "react";
import { X, LayoutDashboard, ChevronDown, LogOut } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { NAVIGATION_CONFIG, NavigationItem } from "@/config/navigation";
import { VersionDisplay } from "@/components/VersionDisplay";
import { api } from "@/lib/api";

interface UserProfile {
    id: number;
    email: string;
    name: string;
}

export function Sidebar({ isOpen, onClose }: { isOpen?: boolean, onClose?: () => void }) {
    const location = useLocation();
    const pathname = location.pathname;
    const [searchParams] = useSearchParams();
    const currentContext = searchParams.get("context");

    const [user, setUser] = useState<UserProfile | null>(null);
    const [openCategories, setOpenCategories] = useState<Record<string, boolean>>({});
    const [crds, setCrds] = useState<any[]>([]);
    const [openGroups, setOpenGroups] = useState<Record<string, boolean>>({});

    useEffect(() => {
        if (!pathname) return;

        // Expand active category
        const activeItem = NAVIGATION_CONFIG.find(item =>
            item.path === pathname || (item.path !== "/" && pathname.startsWith(item.path))
        );
        if (activeItem?.category) {
            setOpenCategories(prev => ({ ...prev, [activeItem.category!]: true }));
        }

        // Special handling for CRDs
        if (pathname.startsWith("/kube-crds")) {
            setOpenCategories(prev => ({ ...prev, 'Custom Resources': true }));

            // If it's a specific CRD, expand its group
            const crdNameMatch = pathname.match(/\/kube-crds\/([^\/\?]+)/);
            if (crdNameMatch && crds.length > 0) {
                const crdName = crdNameMatch[1];
                const crd = crds.find(c => c.name === crdName);
                if (crd?.group) {
                    setOpenGroups(prev => ({ ...prev, [crd.group]: true }));
                }
            }
        }
    }, [pathname, crds]);

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

        // Fetch CRDs for sidebar
        const fetchCRDs = async () => {
            if (!currentContext) return;
            try {
                const data = await api.get<any>(`/kube/crds`, {
                    headers: { "x-kube-context": currentContext || "" }
                });
                setCrds(data.items || []);
            } catch (err) {
                console.error("Failed to fetch CRDs:", err);
            }
        };
        fetchCRDs();
    }, [currentContext]);

    // Group CRDs
    const groupedCrds = crds.reduce((acc, crd) => {
        const group = crd.group || 'Other';
        if (!acc[group]) acc[group] = [];
        acc[group].push(crd);
        return acc;
    }, {} as Record<string, any[]>);



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

    const toggleCategory = (category: string) => {
        setOpenCategories(prev => ({ ...prev, [category]: !prev[category] }));
    };

    const toggleGroup = (group: string) => {
        setOpenGroups(prev => ({ ...prev, [group]: !prev[group] }));
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
                    {(() => {
                        // Group items by category
                        const uncategorizedItems = NAVIGATION_CONFIG.filter(item => !item.category);
                        const categories = Array.from(new Set(NAVIGATION_CONFIG.map(item => item.category).filter(Boolean))) as string[];

                        // Define custom order for top-level and categories
                        // We want: Dashboard, Nodes, Workloads, Config, Network, Storage, Namespaces, Events, Access Control, Settings
                        const order = [
                            { type: 'path', value: '/' },
                            { type: 'path', value: '/kube-nodes' },
                            { type: 'category', value: 'Workloads' },
                            { type: 'category', value: 'Config' },
                            { type: 'category', value: 'Network' },
                            { type: 'category', value: 'Storage' },
                            { type: 'path', value: '/kube-namespaces' },
                            { type: 'path', value: '/kube-events' },
                            { type: 'category', value: 'Helm' },
                            { type: 'category', value: 'Access Control' },
                            { type: 'category', value: 'Custom Resources' },
                            { type: 'category', value: 'Settings' }
                        ];

                        return order.map((orderItem) => {
                            if (orderItem.type === 'path') {
                                const item = uncategorizedItems.find(i => i.path === orderItem.value);
                                if (!item) return null;
                                const Icon = item.icon;
                                return (
                                    <Link key={item.path} to={getLinkHref(item.path)} className="block" onClick={onClose}>
                                        <Button
                                            variant={isActive(item.path) ? "secondary" : "ghost"}
                                            className={cn(
                                                "w-full justify-start gap-3 h-11 px-4 transition-all duration-200 mt-1",
                                                isActive(item.path) ? "bg-sidebar-accent text-white shadow-sm" : "hover:bg-white/10 hover:text-white"
                                            )}
                                        >
                                            <Icon className={cn("h-4 w-4", isActive(item.path) ? "text-primary" : "opacity-60")} />
                                            <span className="font-medium text-sm">{item.title}</span>
                                        </Button>
                                    </Link>
                                );
                            } else {
                                const categoryName = orderItem.value;
                                const items = NAVIGATION_CONFIG.filter(i => i.category === categoryName && !i.hideFromSidebar);
                                if (items.length === 0) return null;

                                // Find an icon for the category (use first item's icon)
                                const CategoryIcon = items[0].icon;
                                const isOpen = openCategories[categoryName];

                                return (
                                    <div key={categoryName} className="space-y-1">
                                        <Button
                                            variant="ghost"
                                            className="w-full justify-between gap-3 h-11 px-4 mt-1 hover:bg-white/10 hover:text-white"
                                            onClick={() => toggleCategory(categoryName)}
                                        >
                                            <div className="flex items-center gap-3">
                                                <CategoryIcon className="h-4 w-4 opacity-60" />
                                                <span className="font-medium text-sm">{categoryName}</span>
                                            </div>
                                            <ChevronDown className={cn("h-4 w-4 opacity-40 transition-transform", isOpen && "rotate-180")} />
                                        </Button>
                                        {isOpen && (
                                            <div className="ml-4 space-y-1">
                                                {items.map((item) => {
                                                    const Icon = item.icon;
                                                    return (
                                                        <Link key={item.path} to={getLinkHref(item.path)} className="block" onClick={onClose}>
                                                            <Button
                                                                variant={isActive(item.path) ? "secondary" : "ghost"}
                                                                className={cn(
                                                                    "w-full justify-start gap-3 h-10 px-4 transition-all duration-200",
                                                                    isActive(item.path) ? "bg-sidebar-accent text-white shadow-sm" : "hover:bg-white/10 hover:text-white"
                                                                )}
                                                            >
                                                                <Icon className={cn("h-4 w-4", isActive(item.path) ? "text-primary" : "opacity-60")} />
                                                                <span className="font-medium text-sm">{item.title}</span>
                                                            </Button>
                                                        </Link>
                                                    );
                                                })}
                                                {/* Dynamic CRD Groups */}
                                                {categoryName === 'Custom Resources' && (Object.entries(groupedCrds) as [string, any[]][]).sort().map(([group, groupCrds]) => (
                                                    <div key={group} className="space-y-1">
                                                        <Button
                                                            variant="ghost"
                                                            className="w-full justify-between gap-3 h-10 px-4 hover:bg-white/10 hover:text-white"
                                                            onClick={() => toggleGroup(group)}
                                                        >
                                                            <div className="flex items-center gap-3 pl-2">
                                                                <div className="h-1.5 w-1.5 rounded-full bg-muted-foreground/50" />
                                                                <span className="font-medium text-xs truncate max-w-[140px]" title={group}>{group}</span>
                                                            </div>
                                                            <ChevronDown className={cn("h-3 w-3 opacity-40 transition-transform", openGroups[group] && "rotate-180")} />
                                                        </Button>
                                                        {openGroups[group] && (
                                                            <div className="ml-4 space-y-1 border-l border-white/5 pl-2">
                                                                {groupCrds.map((crd: any) => (
                                                                    <Link key={crd.name} to={getLinkHref(`/kube-crds/${crd.name}`)} className="block" onClick={onClose}>
                                                                        <Button
                                                                            variant={isActive(`/kube-crds/${crd.name}`) ? "secondary" : "ghost"}
                                                                            className={cn(
                                                                                "w-full justify-start gap-3 h-9 px-3 transition-all duration-200",
                                                                                isActive(`/kube-crds/${crd.name}`) ? "bg-sidebar-accent text-white shadow-sm" : "hover:bg-white/10 hover:text-white"
                                                                            )}
                                                                        >
                                                                            <span className="font-medium text-xs truncate" title={crd.kind}>{crd.kind}</span>
                                                                        </Button>
                                                                    </Link>
                                                                ))}
                                                            </div>
                                                        )}
                                                    </div>
                                                ))}
                                            </div>
                                        )}
                                    </div>
                                );
                            }
                        });
                    })()}
                </nav>

                <div className="p-4 mt-auto space-y-2">
                    <VersionDisplay />
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


