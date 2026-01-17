"use client";

import { useEffect, useState, Suspense } from "react";
import { useSearchParams } from "react-router-dom";
import { Link } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  LayoutDashboard, HardDrive, Layers, Box, Grid, Globe,
  PlayCircle, Clock, AlertTriangle, CheckCircle, RefreshCw
} from "lucide-react";
import { cn, formatAge } from "@/lib/utils";
import { api } from "@/lib/api";

interface DashboardSummary {
  nodes: number;
  namespaces: number;
  pods: number;
  deployments: number;
  services: number;
  ingresses: number;
  jobs: number;
  cronjobs: number;
}

interface EventInfo {
  name: string;
  namespace: string;
  type: string;
  reason: string;
  message: string;
  object: string;
  count: number;
  last_seen: string;
}

function DashboardContent() {
  const [searchParams] = useSearchParams();
  const selectedContext = searchParams.get("context") || "";
  const selectedNamespace = searchParams.get("namespace") || "";

  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [events, setEvents] = useState<EventInfo[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!selectedContext) {
      setSummary(null);
      setEvents([]);
      return;
    }
    fetchDashboard();
  }, [selectedContext, selectedNamespace]);

  const fetchDashboard = async () => {
    setLoading(true);
    try {
      // Fetch summary
      const data = await api.get<DashboardSummary>(`/kube/dashboard`, {
        headers: { "x-kube-context": selectedContext }
      });
      setSummary(data);

      // Fetch events (if namespace selected)
      if (selectedNamespace) {
        const eventsData = await api.get<any>(`/kube/events?namespace=${selectedNamespace}&limit=10`, {
          headers: { "x-kube-context": selectedContext }
        });
        setEvents(eventsData.events || []);
      } else {
        setEvents([]);
      }
    } catch (error) {
      console.error("Failed to fetch dashboard:", error);
    } finally {
      setLoading(false);
    }
  };

  const getLinkHref = (path: string) => {
    const params = new URLSearchParams();
    if (selectedContext) params.set("context", selectedContext);
    if (selectedNamespace) params.set("namespace", selectedNamespace);
    const paramString = params.toString();
    return paramString ? `${path}?${paramString}` : path;
  };

  const statCards = [
    { label: "Nodes", value: summary?.nodes ?? "-", icon: HardDrive, color: "text-emerald-500", bg: "bg-emerald-500/10", href: "/kube-nodes" },
    { label: "Namespaces", value: summary?.namespaces ?? "-", icon: Layers, color: "text-purple-500", bg: "bg-purple-500/10", href: "/kube-namespaces" },
    { label: "Pods", value: summary?.pods ?? "-", icon: Box, color: "text-blue-500", bg: "bg-blue-500/10", href: "/kube-workload/pods" },
    { label: "Deployments", value: summary?.deployments ?? "-", icon: Layers, color: "text-indigo-500", bg: "bg-indigo-500/10", href: "/kube-workload/deployments" },
    { label: "Services", value: summary?.services ?? "-", icon: Grid, color: "text-cyan-500", bg: "bg-cyan-500/10", href: "/kube-network/services" },
    { label: "Ingresses", value: summary?.ingresses ?? "-", icon: Globe, color: "text-pink-500", bg: "bg-pink-500/10", href: "/kube-network/ingresses" },
    { label: "Jobs", value: summary?.jobs ?? "-", icon: PlayCircle, color: "text-orange-500", bg: "bg-orange-500/10", href: "/kube-workload/jobs" },
    { label: "CronJobs", value: summary?.cronjobs ?? "-", icon: Clock, color: "text-teal-500", bg: "bg-teal-500/10", href: "/kube-workload/cron-jobs" },
  ];

  return (
    <div className="flex flex-col items-center w-full min-h-screen bg-transparent p-4 md:p-0">
      <div className="w-full max-w-6xl space-y-8">

        {/* Stats Grid */}
        {loading ? (
          <div className="flex items-center justify-center py-12">
            <RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        ) : !selectedContext ? (
          <Card className="border-none shadow-xl bg-card/50 backdrop-blur-sm rounded-3xl">
            <CardContent className="p-12 text-center">
              <LayoutDashboard className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
              <p className="text-muted-foreground text-sm">
                Select a cluster from the top bar to view the dashboard.
              </p>
            </CardContent>
          </Card>
        ) : (
          <>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              {statCards.map((stat) => (
                <Link key={stat.label} to={getLinkHref(stat.href)}>
                  <Card className="border-none shadow-lg bg-card/50 backdrop-blur-sm rounded-2xl hover:bg-card/80 transition-all cursor-pointer group">
                    <CardContent className="p-5">
                      <div className="flex items-center gap-3">
                        <div className={cn("p-2.5 rounded-xl", stat.bg)}>
                          <stat.icon className={cn("h-5 w-5", stat.color)} />
                        </div>
                        <div>
                          <p className="text-2xl font-bold">{stat.value}</p>
                          <p className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">{stat.label}</p>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </Link>
              ))}
            </div>

            {/* Events Section */}
            <Card className="border-none shadow-xl bg-card/50 backdrop-blur-sm rounded-3xl overflow-hidden">
              <CardHeader className="border-b bg-card/50 px-6 py-4">
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base font-bold flex items-center gap-2">
                    <AlertTriangle className="h-4 w-4 text-yellow-500" />
                    Recent Events
                  </CardTitle>
                  {selectedNamespace && (
                    <Link to={getLinkHref("/kube-events")} className="text-xs text-primary hover:underline">
                      View all →
                    </Link>
                  )}
                </div>
              </CardHeader>
              <CardContent className="p-4">
                {!selectedNamespace ? (
                  <p className="text-sm text-muted-foreground text-center py-6">
                    Select a namespace to view events
                  </p>
                ) : events.length === 0 ? (
                  <div className="text-center py-6">
                    <CheckCircle className="h-8 w-8 mx-auto text-green-500/50 mb-2" />
                    <p className="text-sm text-muted-foreground">No recent events</p>
                  </div>
                ) : (
                  <div className="space-y-2">
                    {events.map((e, idx) => (
                      <div
                        key={e.name + idx}
                        className={cn(
                          "p-3 rounded-xl border flex items-start gap-3",
                          e.type === "Warning"
                            ? "bg-amber-500/5 border-amber-500/20"
                            : "bg-muted/30 border-muted/20"
                        )}
                      >
                        {e.type === "Warning" ? (
                          <AlertTriangle className="h-4 w-4 text-amber-500 mt-0.5 shrink-0" />
                        ) : (
                          <CheckCircle className="h-4 w-4 text-blue-500 mt-0.5 shrink-0" />
                        )}
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2 flex-wrap">
                            <span className="text-[10px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded bg-secondary text-secondary-foreground">
                              {e.reason}
                            </span>
                            <span className="text-[10px] text-muted-foreground">{e.object}</span>
                            <span className="text-[10px] text-muted-foreground">{formatAge(e.last_seen)}</span>
                          </div>
                          <p className="text-xs mt-1 text-foreground/80 line-clamp-1">{e.message}</p>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </>
        )}
      </div>
    </div>
  );
}

export default function Home() {
  return (
    <Suspense fallback={<div className="flex items-center justify-center min-h-screen"><RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" /></div>}>
      <DashboardContent />
    </Suspense>
  );
}
