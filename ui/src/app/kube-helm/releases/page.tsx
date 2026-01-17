"use client";

import { useState, useEffect } from "react";
import { useSearchParams } from "react-router-dom";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { api } from "@/lib/api";


interface HelmRelease {
    name: string;
    namespace: string;
    revision: number;
    updated: string;
    status: string;
    chart_name: string;
    chart_version: string;
    app_version: string;
}

export default function HelmReleasesPage() {
    const [searchParams] = useSearchParams();
    const [releases, setReleases] = useState<HelmRelease[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState("");

    const fetchReleases = () => {
        const context = searchParams.get("context");
        const namespace = searchParams.get("namespace");

        if (!context) {
            setReleases([]);
            return;
        }

        setLoading(true);
        setError("");

        const params = new URLSearchParams();
        if (context) params.append("context", context);
        if (namespace) params.append("namespace", namespace);

        api.get<{ releases: HelmRelease[] }>(`/kube/helm/releases?${params.toString()}`)
            .then(data => setReleases(data.releases || []))
            .catch(err => setError(err.message))
            .finally(() => setLoading(false));
    };

    const [searchQuery, setSearchQuery] = useState("");

    useEffect(() => {
        fetchReleases();
    }, [searchParams]);

    const context = searchParams.get("context");

    const filteredReleases = releases.filter(r =>
        r.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        r.chart_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        r.status.toLowerCase().includes(searchQuery.toLowerCase())
    );

    return (
        <div className="flex flex-col h-full bg-background/50 backdrop-blur-sm">
            {/* PageHeader removed as per user request */}

            <div className="flex-1 p-6 min-h-0 overflow-hidden flex flex-col">
                <div className="rounded-xl border border-border bg-card shadow-sm flex flex-col min-h-0 overflow-hidden">
                    <div className="p-4 border-b border-border bg-muted/30 flex items-center justify-between gap-4">
                        <div className="flex items-center gap-2">
                            <span className="text-xs font-mono text-muted-foreground uppercase tracking-wider">
                                {filteredReleases.length} / {releases.length} Releases
                            </span>
                        </div>

                        <div className="flex items-center gap-2 flex-1 justify-end max-w-md">
                            <div className="relative flex-1">
                                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground">
                                    <circle cx="11" cy="11" r="8" />
                                    <path d="m21 21-4.3-4.3" />
                                </svg>
                                <input
                                    type="text"
                                    placeholder="Search releases..."
                                    value={searchQuery}
                                    onChange={(e) => setSearchQuery(e.target.value)}
                                    className="w-full h-8 pl-8 pr-4 text-xs bg-background border border-border rounded-md focus:outline-none focus:ring-1 focus:ring-primary/50"
                                />
                            </div>
                            <button
                                onClick={fetchReleases}
                                className="p-1.5 hover:bg-background rounded-md text-muted-foreground hover:text-foreground transition-colors border border-transparent hover:border-border"
                                title="Refresh Releases"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={`lucide lucide-rotate-cw ${loading ? "animate-spin" : ""}`}>
                                    <path d="M21 12a9 9 0 1 1-9-9c2.52 0 4.93 1 6.74 2.74L21 8" />
                                    <path d="M21 3v5h-5" />
                                </svg>
                            </button>
                        </div>
                    </div>

                    <div className="flex-1 overflow-auto">
                        {!context ? (
                            <div className="flex flex-col items-center justify-center h-full p-8 text-center text-muted-foreground">
                                <span className="font-mono text-sm mb-2">No Context Selected</span>
                                <span className="text-xs">Select a cluster from the top bar to view Helm releases.</span>
                            </div>
                        ) : loading ? (
                            <div className="p-8 text-center text-muted-foreground font-mono text-sm animate-pulse">
                                Loading releases...
                            </div>
                        ) : (
                            <>
                                {error && (
                                    <div className="p-8 text-center text-destructive font-mono text-sm">
                                        Error: {error}
                                    </div>
                                )}


                                {!loading && !error && releases.length === 0 && (
                                    <div className="p-8 text-center text-muted-foreground font-mono text-sm italic">
                                        No Helm releases found.
                                    </div>
                                )}

                                {!loading && !error && releases.length > 0 && filteredReleases.length === 0 && (
                                    <div className="p-8 text-center text-muted-foreground font-mono text-sm italic">
                                        No releases match your search.
                                    </div>
                                )}

                                {!loading && !error && filteredReleases.length > 0 && (
                                    <table className="w-full text-sm text-left font-mono">
                                        <thead className="bg-muted/50 text-muted-foreground border-b border-border sticky top-0 backdrop-blur-sm">
                                            <tr>
                                                <th className="p-4 font-medium w-[200px]">Name</th>
                                                <th className="p-4 font-medium w-[150px]">Namespace</th>
                                                <th className="p-4 font-medium w-[150px]">Chart</th>
                                                <th className="p-4 font-medium w-[100px]">Version</th>
                                                <th className="p-4 font-medium w-[100px]">App Version</th>
                                                <th className="p-4 font-medium w-[100px]">Status</th>
                                                <th className="p-4 font-medium w-[80px]">Revision</th>
                                                <th className="p-4 font-medium">Updated</th>
                                            </tr>
                                        </thead>
                                        <tbody className="divide-y divide-border">
                                            {filteredReleases.map((r, i) => (
                                                <tr key={i} className="hover:bg-muted/30 transition-colors group">
                                                    <td className="p-4 font-semibold text-foreground/90">{r.name}</td>
                                                    <td className="p-4 text-muted-foreground">{r.namespace}</td>
                                                    <td className="p-4 text-foreground/80">{r.chart_name}</td>
                                                    <td className="p-4 text-muted-foreground">{r.chart_version}</td>
                                                    <td className="p-4 text-muted-foreground">{r.app_version}</td>
                                                    <td className="p-4">
                                                        <Badge
                                                            variant="outline"
                                                            className={
                                                                r.status === "deployed"
                                                                    ? "text-green-500 border-green-500/20 bg-green-500/10"
                                                                    : r.status === "failed"
                                                                        ? "text-destructive border-destructive/20 bg-destructive/10"
                                                                        : "text-muted-foreground border-border bg-muted/30"
                                                            }
                                                        >
                                                            {r.status}
                                                        </Badge>
                                                    </td>
                                                    <td className="p-4 text-muted-foreground text-xs">
                                                        {r.revision}
                                                    </td>
                                                    <td className="p-4 text-muted-foreground text-xs whitespace-nowrap">
                                                        {new Date(r.updated).toLocaleString()}
                                                    </td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                )}
                            </>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
}
