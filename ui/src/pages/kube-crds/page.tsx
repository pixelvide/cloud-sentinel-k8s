"use client";

import { useState, useEffect } from "react";
import { useSearchParams } from "react-router-dom";
import { api } from "@/lib/api";
import { Search, FileCode } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Link } from "react-router-dom";
import { Badge } from "@/components/ui/badge";

interface CRDInfo {
    name: string;
    group: string;
    version: string;
    scope: string;
    kind: string;
    resource: string;
    short_names?: string[];
    categories?: string[];
}

export default function CRDsPage() {
    const [searchParams] = useSearchParams();
    const currentContext = searchParams.get("context") || "";

    const [crds, setCrds] = useState<CRDInfo[]>([]);
    const [loading, setLoading] = useState(true);
    const [search, setSearch] = useState("");

    useEffect(() => {
        const fetchCRDs = async () => {
            setLoading(true);
            try {
                const res = await api.get<{ items: CRDInfo[] }>(`/kube/crds`, {
                    headers: { "x-kube-context": currentContext || "" },
                });
                setCrds(res.items || []);
            } catch (err) {
                console.error(err);
            } finally {
                setLoading(false);
            }
        };
        fetchCRDs();
    }, [currentContext]);

    const filtered = crds.filter(
        (c) =>
            c.name.includes(search.toLowerCase()) ||
            c.group.includes(search.toLowerCase()) ||
            c.kind.toLowerCase().includes(search.toLowerCase())
    );

    useEffect(() => {
        const query = searchParams.get("q") || "";
        setSearch(query);
    }, [searchParams]);

    return (
        <div className="h-full flex flex-col">
            <div className="flex-1 overflow-auto p-4">
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                    {filtered.map((crd) => (
                        <Link
                            key={crd.name}
                            to={`/kube-crds/${crd.name}?context=${currentContext}`}
                            className="group block p-4 rounded-xl border bg-card hover:bg-accent/5 transition-all hover:border-primary/20 hover:shadow-lg"
                        >
                            <div className="flex items-start justify-between mb-3">
                                <div className="p-2 rounded-lg bg-primary/5 text-primary group-hover:bg-primary/10 transition-colors">
                                    <FileCode className="h-5 w-5" />
                                </div>
                                <Badge
                                    variant={crd.scope === "Cluster" ? "secondary" : "outline"}
                                    className="text-[10px] h-5"
                                >
                                    {crd.scope}
                                </Badge>
                            </div>

                            <h3 className="font-semibold truncate mb-1" title={crd.kind}>
                                {crd.kind}
                            </h3>
                            <p className="text-xs text-muted-foreground mb-4 font-mono truncate" title={crd.name}>
                                {crd.name}
                            </p>

                            <div className="flex flex-wrap gap-1.5 mt-auto">
                                <Badge
                                    variant="secondary"
                                    className="text-[10px] h-5 bg-background/50 border font-mono"
                                >
                                    {crd.group}/{crd.version}
                                </Badge>
                                {crd.categories?.map((cat) => (
                                    <Badge key={cat} variant="outline" className="text-[10px] h-5">
                                        {cat}
                                    </Badge>
                                ))}
                            </div>
                        </Link>
                    ))}

                    {!loading && filtered.length === 0 && (
                        <div className="col-span-full flex flex-col items-center justify-center py-12 text-center">
                            <FileCode className="h-12 w-12 text-muted-foreground/20 mb-4" />
                            <h3 className="text-lg font-medium">No CRDs found</h3>
                            <p className="text-sm text-muted-foreground">
                                {search
                                    ? "Try adjusting your search terms"
                                    : "No Custom Resource Definitions found in this cluster"}
                            </p>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}
