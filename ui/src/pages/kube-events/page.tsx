"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { AlertCircle, RefreshCw, CheckCircle, AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn, formatAge } from "@/lib/utils";
import { NamespaceBadge } from "@/components/NamespaceBadge";
import { ResourceDetailsSheet } from "@/components/ResourceDetailsSheet";
import { api } from "@/lib/api";

interface EventInfo {
    name: string;
    namespace: string;
    type: string;
    reason: string;
    message: string;
    object: string;
    count: number;
    first_seen: string;
    last_seen: string;
}

function EventsContent() {
    const [searchParams] = useSearchParams();
    const selectedContext = searchParams.get("context") || "";
    const selectedNamespace = searchParams.get("namespace") || "";

    const [events, setEvents] = useState<EventInfo[]>([]);
    const [loading, setLoading] = useState(false);
    const searchQuery = searchParams.get("q") || "";
    const [selectedEvent, setSelectedEvent] = useState<EventInfo | null>(null);

    const filteredEvents = events.filter(
        (e) =>
            e.message.toLowerCase().includes(searchQuery.toLowerCase()) ||
            e.object.toLowerCase().includes(searchQuery.toLowerCase()) ||
            e.reason.toLowerCase().includes(searchQuery.toLowerCase())
    );

    useEffect(() => {
        if (!selectedContext || !selectedNamespace) {
            setEvents([]);
            return;
        }
        fetchEvents();
    }, [selectedContext, selectedNamespace]);

    const fetchEvents = async () => {
        setLoading(true);
        setEvents([]);
        try {
            const data = await api.get<any>(`/kube/events?namespace=${selectedNamespace}`, {
                headers: { "x-kube-context": selectedContext || "" },
            });
            setEvents(data.events || []);
        } catch (error) {
            console.error("Failed to fetch events:", error);
        } finally {
            setLoading(false);
        }
    };

    const handleRefresh = () => {
        if (selectedContext && selectedNamespace) {
            fetchEvents();
        }
    };

    return (
        <div className="flex flex-col items-center w-full min-h-screen bg-background/50 p-4 md:p-0">
            <div className="w-full max-w-6xl space-y-8">
                <Card className="border-none shadow-2xl shadow-black/5 bg-card/50 backdrop-blur-sm overflow-hidden rounded-3xl">
                    <CardHeader className="border-b bg-card/50 px-8 py-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <CardTitle className="text-lg font-bold flex items-center gap-3">
                                    <AlertCircle className="h-5 w-5 text-yellow-500" />
                                    {events.length} Events Found
                                </CardTitle>
                                <CardDescription>
                                    {!selectedContext || !selectedNamespace
                                        ? "Select a namespace from the top bar to view events"
                                        : null}
                                </CardDescription>
                            </div>
                            <Button
                                variant="outline"
                                size="sm"
                                onClick={handleRefresh}
                                disabled={!selectedContext || !selectedNamespace || loading}
                                className="rounded-xl"
                            >
                                <RefreshCw className={cn("h-4 w-4 mr-2", loading && "animate-spin")} />
                                Refresh
                            </Button>
                        </div>
                    </CardHeader>
                    <CardContent className="p-8">
                        {loading ? (
                            <div className="flex items-center justify-center py-12">
                                <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
                            </div>
                        ) : !selectedContext ? (
                            <div className="text-center py-12">
                                <AlertCircle className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    Select a cluster from the top bar to view events.
                                </p>
                            </div>
                        ) : !selectedNamespace ? (
                            <div className="text-center py-12">
                                <AlertCircle className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                                <p className="text-muted-foreground text-sm">
                                    Select a namespace from the top bar to view events.
                                </p>
                            </div>
                        ) : events.length === 0 ? (
                            <div className="text-center py-12">
                                <CheckCircle className="h-12 w-12 mx-auto text-green-500/30 mb-4" />
                                <p className="text-muted-foreground text-sm">No events found in this namespace.</p>
                            </div>
                        ) : (
                            <div className="space-y-3">
                                {filteredEvents.map((e, idx) => (
                                    <div
                                        key={e.name + idx}
                                        className={cn(
                                            "p-4 rounded-xl border transition-colors cursor-pointer",
                                            e.type === "Warning"
                                                ? "bg-amber-500/5 border-amber-500/20 hover:bg-amber-500/10"
                                                : "bg-muted/30 border-muted/20 hover:bg-muted/50"
                                        )}
                                        onClick={() => setSelectedEvent(e)}
                                    >
                                        <div className="flex items-start gap-3 justify-between">
                                            <div className="flex items-start gap-3 flex-1 min-w-0">
                                                <div
                                                    className={cn(
                                                        "p-2 rounded-lg mt-0.5",
                                                        e.type === "Warning" ? "bg-amber-500/10" : "bg-blue-500/10"
                                                    )}
                                                >
                                                    {e.type === "Warning" ? (
                                                        <AlertTriangle className="h-4 w-4 text-amber-500" />
                                                    ) : (
                                                        <CheckCircle className="h-4 w-4 text-blue-500" />
                                                    )}
                                                </div>
                                                <div className="flex-1 min-w-0">
                                                    <div className="flex items-center gap-2 flex-wrap">
                                                        <span className="text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full bg-secondary text-secondary-foreground">
                                                            {e.reason}
                                                        </span>
                                                        <span className="text-[10px] font-medium text-muted-foreground">
                                                            {e.object}
                                                        </span>
                                                        {e.count > 1 && (
                                                            <span className="text-[10px] font-medium text-muted-foreground">
                                                                ×{e.count}
                                                            </span>
                                                        )}
                                                    </div>
                                                    <p className="text-sm mt-1 text-foreground/80">{e.message}</p>
                                                </div>
                                            </div>

                                            {/* Namespace */}
                                            <div
                                                className="flex flex-col items-end min-w-[120px]"
                                                onClick={(e) => e.stopPropagation()}
                                            >
                                                <NamespaceBadge namespace={e.namespace} />
                                            </div>

                                            {/* Age */}
                                            <div className="flex flex-col items-end min-w-[80px]">
                                                <span className="text-xs text-muted-foreground">
                                                    {formatAge(e.last_seen)}
                                                </span>
                                            </div>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </CardContent>
                </Card>
            </div>

            <ResourceDetailsSheet
                isOpen={!!selectedEvent}
                onClose={() => setSelectedEvent(null)}
                context={selectedContext}
                namespace={selectedEvent?.namespace || ""}
                name={selectedEvent?.name || ""}
                kind="Event"
            />
        </div>
    );
}

export default function EventsPage() {
    return (
        <Suspense
            fallback={
                <div className="flex items-center justify-center min-h-screen">
                    <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
                </div>
            }
        >
            <EventsContent />
        </Suspense>
    );
}
