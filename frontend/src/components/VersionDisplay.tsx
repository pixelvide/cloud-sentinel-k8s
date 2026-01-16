"use client";

import { useEffect, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { RefreshCcw } from "lucide-react";

interface VersionData {
    current_version: string;
    latest_version: string;
    update_available: boolean;
}

export function VersionDisplay() {
    const [version, setVersion] = useState<VersionData | null>(null);
    const [loading, setLoading] = useState(true);

    const fetchVersion = async () => {
        try {
            const res = await fetch("/api/v1/settings/version", { credentials: "include" });
            if (res.ok) {
                const data = await res.json();
                setVersion(data);
            }
        } catch (err) {
            console.error("Failed to fetch version:", err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchVersion();
    }, []);

    if (loading) return (
        <div className="px-3 py-2 flex items-center gap-2 opacity-30">
            <RefreshCcw className="h-3 w-3 animate-spin" />
            <span className="text-[10px] font-medium uppercase tracking-wider">v...</span>
        </div>
    );

    if (!version) return null;

    const vString = version.current_version === "dev" ? "dev" : `v${version.current_version}`;

    return (
        <div className="px-3 py-2 flex flex-col gap-1.5">
            <div className="flex items-center justify-between">
                <div className="flex items-center gap-2 group cursor-default">
                    <span className="text-[10px] font-bold text-sidebar-foreground/40 uppercase tracking-widest transition-colors group-hover:text-sidebar-foreground/60">
                        {vString}
                    </span>
                    {version.update_available && (
                        <div className="relative flex h-2 w-2">
                            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary opacity-75"></span>
                            <span className="relative inline-flex rounded-full h-2 w-2 bg-primary"></span>
                        </div>
                    )}
                </div>
            </div>
            {version.update_available && (
                <a
                    href={`https://github.com/pixelvide/cloud-sentinel-k8s/releases/tag/${version.latest_version}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="group"
                >
                    <Badge variant="outline" className="text-[10px] h-5 border-primary/20 bg-primary/5 text-primary hover:bg-primary/10 transition-colors cursor-pointer w-fit py-0 px-2 font-semibold">
                        Update available: {version.latest_version}
                    </Badge>
                </a>
            )}
        </div>
    );
}
