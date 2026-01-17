import { Sidebar } from "@/components/Sidebar";
import { useLocation, useNavigate } from "react-router-dom";
import { useState, useEffect } from "react";
import { Menu, LayoutDashboard, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ClusterContextSelector } from "@/components/ClusterContextSelector";
import { GlobalSearch } from "@/components/GlobalSearch";
import { NAVIGATION_CONFIG } from "@/config/navigation";

// Derive Page Configuration from Navigation Config
const PAGE_CONFIG = NAVIGATION_CONFIG.reduce((acc, item) => {
    acc[item.path] = item;
    return acc;
}, {} as Record<string, typeof NAVIGATION_CONFIG[0]>);

export function ClientLayout({ children }: { children: React.ReactNode }) {
    const location = useLocation();
    const pathname = location.pathname;
    const navigate = useNavigate();
    const [isSidebarOpen, setIsSidebarOpen] = useState(false);
    const [isAuthChecking, setIsAuthChecking] = useState(true);
    const isLoginPage = pathname === "/login";
    const isExecPage = pathname?.startsWith("/exec");

    // Auth check for protected pages
    useEffect(() => {
        // Skip auth check for login page
        if (isLoginPage) {
            setIsAuthChecking(false);
            return;
        }

        const checkAuth = async () => {
            try {
                const res = await fetch("/api/v1/me", { credentials: "include" });
                if (res.status === 401) {
                    navigate("/login");
                    return;
                }
            } catch (err) {
                console.error("Auth check failed:", err);
                navigate("/login");
                return;
            }
            setIsAuthChecking(false);
        };
        checkAuth();
    }, [pathname, isLoginPage, navigate]);

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

    // Helper to find configuration for current path
    const getCurrentPageConfig = (currentPath: string) => {
        // Try exact match first
        if (PAGE_CONFIG[currentPath]) {
            return PAGE_CONFIG[currentPath];
        }

        // Try to find a matching dynamic route
        // We look for navigation items with brackets like /kube-crds/[crd]
        return NAVIGATION_CONFIG.find(item => {
            if (!item.path.includes('[')) return false;

            // Convert pattern to regex
            // e.g. /kube-crds/[crd] -> /kube-crds/[^/]+
            const pattern = item.path.replace(/\[[^\]]+\]/g, '[^/]+');
            const regex = new RegExp(`^${pattern}$`);
            return regex.test(currentPath);
        });
    };

    const currentPage = getCurrentPageConfig(pathname);

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

                {/* Main Header with Global Context Selector (Visible if showHeader is true) */}
                {currentPage && currentPage.showHeader !== false && (
                    <header className="flex flex-col md:flex-row md:items-center justify-between gap-4 md:gap-0 px-4 py-3 md:px-8 md:py-4 border-b bg-background/50 backdrop-blur-sm z-20">
                        <div className="flex items-center gap-3 shrink-0">
                            {(() => {
                                const Icon = currentPage.icon;
                                return <Icon className="h-5 w-5 text-primary mt-1" />;
                            })()}
                            <div className="flex flex-col">
                                <h1 className="text-lg font-semibold tracking-tight leading-none">
                                    {currentPage.title}
                                </h1>
                                <p className="text-xs text-muted-foreground mt-1 hidden md:block">
                                    {currentPage.description}
                                </p>
                            </div>
                        </div>
                        <div className="flex flex-col md:flex-row items-stretch md:items-center justify-end gap-3 md:gap-4 shrink-0">
                            <ClusterContextSelector />
                            {currentPage.searchPlaceholder && (
                                <GlobalSearch placeholder={currentPage.searchPlaceholder} />
                            )}
                        </div>
                    </header>
                )}

                <main className={`flex-1 md:overflow-y-auto overflow-x-hidden transition-all duration-500 ${pathname?.startsWith('/exec') ? 'p-0' : 'p-6 md:p-8 lg:p-12'}`}>
                    {children}
                </main>
            </div>
        </div>
    );
}
