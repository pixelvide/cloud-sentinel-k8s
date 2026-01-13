"use client";

import { Suspense } from "react";
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card";
import { Share2, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";

function PortForwardingContent() {
    return (
        <div className="flex flex-col items-center w-full min-h-screen bg-background/50 p-4 md:p-0">
            <div className="w-full max-w-6xl space-y-8">
                <Card className="border-none shadow-2xl shadow-black/5 bg-card/50 backdrop-blur-sm overflow-hidden rounded-3xl">
                    <CardHeader className="border-b bg-card/50 px-8 py-6">
                        <div className="flex items-center justify-between">
                            <div>
                                <CardTitle className="text-lg font-bold flex items-center gap-3">
                                    <Share2 className="h-5 w-5 text-indigo-500" />
                                    Port Forwarding
                                </CardTitle>
                                <CardDescription>
                                    Manage active port forwards to cluster resources
                                </CardDescription>
                            </div>
                            <Button
                                variant="outline"
                                size="sm"
                                disabled={true}
                                className="rounded-xl"
                            >
                                <RefreshCw className="h-4 w-4 mr-2" />
                                Refresh
                            </Button>
                        </div>
                    </CardHeader>
                    <CardContent className="p-8">
                        <div className="text-center py-12">
                            <Share2 className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                            <p className="text-muted-foreground text-sm max-w-md mx-auto">
                                Port forwarding management is coming soon. This page will eventually allow you to list, create, and terminate active port forwards to your cluster pods and services.
                            </p>
                        </div>
                    </CardContent>
                </Card>
            </div>
        </div >
    );
}

export default function PortForwardingPage() {
    return (
        <Suspense fallback={<div className="flex items-center justify-center min-h-screen"><RefreshCw className="h-8 w-8 animate-spin text-muted-foreground" /></div>}>
            <PortForwardingContent />
        </Suspense>
    );
}
