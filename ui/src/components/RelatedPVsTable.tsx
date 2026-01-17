"use client";

import { useEffect, useState } from "react";
import { Badge } from "@/components/ui/badge";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { api } from "@/lib/api";
import { formatAge } from "@/lib/utils";
import { cn } from "@/lib/utils";
import { Loader2 } from "lucide-react";

interface RelatedPVsTableProps {
    resource: any;
    context: string;
}

interface PV {
    name: string;
    capacity: string;
    access_modes: string;
    reclaim_policy: string;
    status: string;
    claim: string;
    storage_class: string;
    age: string;
}

export function RelatedPVsTable({ resource, context }: RelatedPVsTableProps) {
    const [pvs, setPvs] = useState<PV[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState("");

    const { kind, metadata } = resource;

    useEffect(() => {
        const fetchPVs = async () => {
            if (kind !== "StorageClass") return;

            setLoading(true);
            setError("");
            setPvs([]);

            try {
                const data = await api.get<{ pvs: PV[] }>(`/kube/pvs?storageClass=${metadata.name}`, {
                    headers: { "x-kube-context": context || "" }
                });
                setPvs(data.pvs || []);
            } catch (err: any) {
                console.error("Failed to fetch related PVs:", err);
                setError(err.message || "Failed to fetch PVs");
            } finally {
                setLoading(false);
            }
        };

        fetchPVs();
    }, [resource, context, kind, metadata.name]);

    if (kind !== "StorageClass") return null;
    if (!loading && pvs.length === 0 && !error) return null;

    const getStatusColor = (status: string) => {
        switch (status) {
            case "Bound":
                return "bg-emerald-50 text-emerald-700 border-emerald-200";
            case "Available":
                return "bg-blue-50 text-blue-700 border-blue-200";
            case "Released":
                return "bg-amber-50 text-amber-700 border-amber-200";
            case "Failed":
                return "bg-red-50 text-red-700 border-red-200";
            default:
                return "bg-muted text-muted-foreground border-border";
        }
    };

    return (
        <div className="space-y-3">
            <div className="flex items-center justify-between">
                <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wider">
                    Persistent Volumes
                </h3>
                {pvs.length > 0 && (
                    <Badge variant="secondary" className="bg-muted text-muted-foreground border-none text-[10px] font-bold px-1.5 h-4">
                        {pvs.length}
                    </Badge>
                )}
            </div>

            <div className="rounded-md border border-border bg-card overflow-hidden shadow-sm overflow-x-auto">
                {loading ? (
                    <div className="p-4 flex items-center justify-center text-muted-foreground text-xs gap-2">
                        <Loader2 className="h-3 w-3 animate-spin" />
                        Fetching related volumes...
                    </div>
                ) : error ? (
                    <div className="p-4 text-destructive text-xs">
                        {error}
                    </div>
                ) : pvs.length > 0 ? (
                    <Table>
                        <TableHeader className="bg-muted text-muted-foreground">
                            <TableRow>
                                <TableHead className="h-8 text-[11px] font-medium whitespace-nowrap">Name</TableHead>
                                <TableHead className="h-8 text-[11px] font-medium whitespace-nowrap">Capacity</TableHead>
                                <TableHead className="h-8 text-[11px] font-medium whitespace-nowrap">Status</TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {pvs.map((pv) => (
                                <TableRow key={pv.name} className="hover:bg-muted/50 border-border">
                                    <TableCell className="py-2 text-xs font-mono font-medium whitespace-nowrap">{pv.name}</TableCell>
                                    <TableCell className="py-2 text-xs text-muted-foreground whitespace-nowrap">{pv.capacity}</TableCell>
                                    <TableCell className="py-2 whitespace-nowrap">
                                        <Badge
                                            variant="outline"
                                            className={cn("text-[9px] font-bold px-1.5 py-0 h-4 border", getStatusColor(pv.status))}
                                        >
                                            {pv.status}
                                        </Badge>
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                ) : (
                    <div className="p-4 text-center text-xs text-muted-foreground italic">
                        No persistent volumes found.
                    </div>
                )}
            </div>
        </div>
    );
}
