"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { History as HistoryIcon, RefreshCw, ChevronLeft, ChevronRight, Info, Search, Filter, ArrowRight, Clock, ShieldCheck, Activity, Database, Key, Server, LayoutDashboard, Boxes, GitBranch } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn, formatAge } from "@/lib/utils";
import { api } from "@/lib/api";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";

interface AuditLog {
    id: number;
    app_name: string;
    action: string;
    actor: string;
    ip_address: string;
    user_agent: string;
    payload: string;
    created_at: string;
}

function AuditLogsContent() {
    const [logs, setLogs] = useState<AuditLog[]>([]);
    const [loading, setLoading] = useState(false);
    const [page, setPage] = useState(1);
    const [total, setTotal] = useState(0);
    const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null);
    const pageSize = 20;

    const fetchLogs = async (p: number) => {
        setLoading(true);
        try {
            const data = await api.get<any>(`/settings/audit-logs?page=${p}&pageSize=${pageSize}`);
            setLogs(data.logs || []);
            setTotal(data.total || 0);
            setPage(data.page || 1);
        } catch (error) {
            console.error("Failed to fetch audit logs:", error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchLogs(page);
    }, [page]);

    const totalPages = Math.ceil(total / pageSize);

    const getActionColor = (action: string) => {
        if (action.includes("LOGIN")) return "bg-green-500/10 text-green-500 border-green-500/20";
        if (action.includes("DELETE")) return "bg-red-500/10 text-red-500 border-red-500/20";
        if (action.includes("UPDATE") || action.includes("SET") || action.includes("UPSERT")) return "bg-blue-500/10 text-blue-500 border-blue-500/20";
        if (action.includes("EXEC")) return "bg-amber-500/10 text-amber-500 border-amber-500/20";
        return "bg-slate-500/10 text-slate-500 border-slate-500/20";
    };

    return (
        <div className="flex flex-col items-center w-full min-h-screen bg-transparent p-4 md:p-8">
            <div className="w-full max-w-6xl space-y-6">
                <div className="flex flex-col gap-1">
                    <h1 className="text-2xl font-bold tracking-tight">Audit Logs</h1>
                    <p className="text-muted-foreground">Monitor and track all security and configuration actions performed by your account.</p>
                </div>

                <Card className="border-none shadow-2xl shadow-black/5 bg-card/50 backdrop-blur-sm overflow-hidden rounded-3xl">
                    <CardHeader className="border-b bg-card/50 px-8 py-6">
                        <div className="flex items-center justify-between">
                            <div className="flex items-center gap-3">
                                <HistoryIcon className="h-5 w-5 text-primary" />
                                <div>
                                    <CardTitle className="text-lg font-bold">Activity History</CardTitle>
                                    <CardDescription>Total {total} logs found</CardDescription>
                                </div>
                            </div>
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={() => fetchLogs(page)}
                                disabled={loading}
                                className="rounded-xl"
                            >
                                <RefreshCw className={cn("h-4 w-4 mr-2", loading && "animate-spin")} />
                                Refresh
                            </Button>
                        </div>
                    </CardHeader>
                    <CardContent className="p-0">
                        <Table>
                            <TableHeader>
                                <TableRow className="hover:bg-transparent border-muted/20">
                                    <TableHead className="px-8 py-4 font-semibold">Action</TableHead>
                                    <TableHead className="font-semibold">IP Address</TableHead>
                                    <TableHead className="font-semibold">Details</TableHead>
                                    <TableHead className="font-semibold text-right pr-8">Time</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {loading && logs.length === 0 ? (
                                    <TableRow>
                                        <TableCell colSpan={4} className="h-32 text-center">
                                            <RefreshCw className="h-8 w-8 animate-spin mx-auto text-muted-foreground" />
                                        </TableCell>
                                    </TableRow>
                                ) : logs.length === 0 ? (
                                    <TableRow>
                                        <TableCell colSpan={4} className="h-32 text-center text-muted-foreground">
                                            No audit logs found.
                                        </TableCell>
                                    </TableRow>
                                ) : (
                                    logs.map((log) => (
                                        <TableRow
                                            key={log.id}
                                            className="group border-muted/10 hover:bg-muted/30 transition-colors cursor-pointer"
                                            onClick={() => setSelectedLog(log)}
                                        >
                                            <TableCell className="px-8 py-4">
                                                <Badge variant="outline" className={cn("rounded-lg font-bold text-[10px] uppercase tracking-wider px-2 py-0.5", getActionColor(log.action))}>
                                                    {log.action.replace(/_/g, " ")}
                                                </Badge>
                                            </TableCell>
                                            <TableCell className="font-mono text-xs opacity-60">
                                                {log.ip_address}
                                            </TableCell>
                                            <TableCell>
                                                {log.payload && log.payload !== "{}" ? (
                                                    <div className="flex items-center gap-2 group-hover:text-primary transition-colors">
                                                        <Info className="h-3.5 w-3.5 opacity-40" />
                                                        <span className="text-xs truncate max-w-[200px] opacity-60">
                                                            {log.payload}
                                                        </span>
                                                    </div>
                                                ) : (
                                                    <span className="text-xs opacity-30 italic">No extra data</span>
                                                )}
                                            </TableCell>
                                            <TableCell className="text-right pr-8">
                                                <div className="flex flex-col items-end">
                                                    <span className="text-xs font-medium">{formatAge(log.created_at)}</span>
                                                    <span className="text-[10px] opacity-40">{new Date(log.created_at).toLocaleString()}</span>
                                                </div>
                                            </TableCell>
                                        </TableRow>
                                    ))
                                )}
                            </TableBody>
                        </Table>

                        {/* Pagination */}
                        {totalPages > 1 && (
                            <div className="flex items-center justify-between px-8 py-4 border-t border-muted/20 bg-muted/10">
                                <div className="text-[10px] uppercase tracking-widest font-bold opacity-40">
                                    Page {page} of {totalPages}
                                </div>
                                <div className="flex items-center gap-2">
                                    <Button
                                        variant="outline"
                                        size="icon"
                                        className="h-8 w-8 rounded-lg"
                                        onClick={() => setPage(p => Math.max(1, p - 1))}
                                        disabled={page === 1 || loading}
                                    >
                                        <ChevronLeft className="h-4 w-4" />
                                    </Button>
                                    <Button
                                        variant="outline"
                                        size="icon"
                                        className="h-8 w-8 rounded-lg"
                                        onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                                        disabled={page === totalPages || loading}
                                    >
                                        <ChevronRight className="h-4 w-4" />
                                    </Button>
                                </div>
                            </div>
                        )}
                    </CardContent>
                </Card>
            </div>

            {/* Detail Modal */}
            <Dialog open={!!selectedLog} onOpenChange={(open) => !open && setSelectedLog(null)}>
                <DialogContent className="max-w-2xl rounded-3xl border-none shadow-2xl bg-card/95 backdrop-blur-xl">
                    <DialogHeader>
                        <DialogTitle className="flex items-center gap-3 text-xl">
                            <HistoryIcon className="h-5 w-5 text-primary" />
                            Audit Log Details
                        </DialogTitle>
                        <DialogDescription>
                            Full trace of the recorded action and its context.
                        </DialogDescription>
                    </DialogHeader>

                    {selectedLog && (
                        <div className="grid gap-6 py-4">
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-1">
                                    <span className="text-[10px] uppercase font-bold opacity-40">Action</span>
                                    <div>
                                        <Badge variant="outline" className={cn("rounded-lg font-bold text-[10px] uppercase tracking-wider px-2 py-0.5", getActionColor(selectedLog.action))}>
                                            {selectedLog.action.replace(/_/g, " ")}
                                        </Badge>
                                    </div>
                                </div>
                                <div className="space-y-1">
                                    <span className="text-[10px] uppercase font-bold opacity-40">Time</span>
                                    <div className="text-sm font-medium">{new Date(selectedLog.created_at).toLocaleString()}</div>
                                </div>
                                <div className="space-y-1">
                                    <span className="text-[10px] uppercase font-bold opacity-40">IP Address</span>
                                    <div className="text-sm font-mono">{selectedLog.ip_address}</div>
                                </div>
                                <div className="space-y-1">
                                    <span className="text-[10px] uppercase font-bold opacity-40">App Name</span>
                                    <div className="text-sm font-medium">{selectedLog.app_name}</div>
                                </div>
                            </div>

                            <div className="space-y-1">
                                <span className="text-[10px] uppercase font-bold opacity-40">User Agent</span>
                                <div className="text-xs bg-muted/30 p-3 rounded-xl border border-muted/20 break-all leading-relaxed opacity-60">
                                    {selectedLog.user_agent}
                                </div>
                            </div>

                            <div className="space-y-2">
                                <span className="text-[10px] uppercase font-bold opacity-40">Payload Data</span>
                                <div className="relative group">
                                    <div className="absolute inset-0 bg-primary/5 rounded-2xl blur-xl opacity-0 group-hover:opacity-100 transition-opacity" />
                                    <pre className="relative p-4 rounded-2xl bg-black/20 font-mono text-[10px] overflow-auto max-h-[400px] border border-white/5 shadow-inner whitespace-pre-wrap break-all leading-relaxed">
                                        {selectedLog.payload && selectedLog.payload !== "{}"
                                            ? JSON.stringify(JSON.parse(selectedLog.payload), null, 2)
                                            : "No additional data recorded."
                                        }
                                    </pre>
                                </div>
                            </div>
                        </div>
                    )}
                </DialogContent>
            </Dialog>
        </div>
    );
}

export default function AuditLogsPage() {
    return (
        <Suspense fallback={<div className="flex items-center justify-center min-h-screen bg-transparent"><RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" /></div>}>
            <AuditLogsContent />
        </Suspense>
    );
}
