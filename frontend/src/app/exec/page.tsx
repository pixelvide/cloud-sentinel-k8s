"use client";

import { useEffect, useRef, useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import "@xterm/xterm/css/xterm.css";
import { AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

function ExecContent() {
    const searchParams = useSearchParams();
    const context = searchParams.get("context");
    const namespace = searchParams.get("namespace");
    const pod = searchParams.get("pod");
    const container = searchParams.get("container");

    const terminalRef = useRef<HTMLDivElement>(null);
    const wsRef = useRef<WebSocket | null>(null);
    const xtermRef = useRef<Terminal | null>(null);
    const [status, setStatus] = useState<"connecting" | "connected" | "disconnected" | "error">("connecting");
    const [errorMsg, setErrorMsg] = useState("");

    useEffect(() => {
        if (!context || !namespace || !pod) {
            setStatus("error");
            setErrorMsg("Missing required parameters: context, namespace, or pod.");
            return;
        }

        const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
        const host = window.location.host;
        let url = `${proto}//${host}/api/v1/kube/exec?context=${context}&namespace=${namespace}&pod=${pod}`;
        if (container) {
            url += `&container=${container}`;
        }

        const term = new Terminal({
            cursorBlink: true,
            theme: {
                background: "#020817",
                foreground: "#f8fafc",
                cursor: "#ffffff",
                selectionBackground: "rgba(255, 255, 255, 0.3)",
            },
            fontFamily: 'Menlo, Monaco, "Courier New", monospace',
            fontSize: 13,
            lineHeight: 1.2,
        });
        const fitAddon = new FitAddon();
        term.loadAddon(fitAddon);

        if (terminalRef.current) {
            terminalRef.current.innerHTML = "";
            term.open(terminalRef.current);
            fitAddon.fit();
            xtermRef.current = term;

            const handleResize = () => fitAddon.fit();
            window.addEventListener("resize", handleResize);

            const ws = new WebSocket(url);
            ws.binaryType = "arraybuffer";
            wsRef.current = ws;

            ws.onopen = () => {
                setStatus("connected");
                term.write("\r\n\x1b[32mConnected to terminal.\x1b[0m\r\n");
                // Note: Backend does not currently support resize via JSON on stdin, 
                // so we do not send initial resize here to avoid it being treated as input.
            };

            ws.onmessage = async (event) => {
                if (typeof event.data === 'string') {
                    // Try to parse as JSON if the backend sends structured data mixed with stdout
                    // However, based on working page.tsx, backend sends raw.
                    // We'll write directly if string.
                    term.write(event.data);
                } else {
                    term.write(new Uint8Array(event.data));
                }
            };

            ws.onclose = () => {
                setStatus("disconnected");
                term.write("\r\n\x1b[31mConnection closed.\x1b[0m\r\n");
            };

            ws.onerror = (err) => {
                setStatus("error");
                console.error("WebSocket error:", err);
                term.write("\r\n\x1b[31mWebSocket error occurred.\x1b[0m\r\n");
            };

            // Send raw data directly to stdin
            term.onData((data) => {
                if (ws.readyState === WebSocket.OPEN) {
                    ws.send(data);
                }
            });

            return () => {
                window.removeEventListener("resize", handleResize);
                ws.close();
                term.dispose();
            };
        }
    }, [context, namespace, pod, container]);

    if (errorMsg) {
        return (
            <div className="flex items-center justify-center min-h-screen bg-[#020817] text-white">
                <Alert variant="destructive" className="max-w-md bg-red-900/20 border-red-900/50 text-red-200">
                    <AlertTriangle className="h-4 w-4" />
                    <AlertTitle>Connection Error</AlertTitle>
                    <AlertDescription>{errorMsg}</AlertDescription>
                </Alert>
            </div>
        );
    }

    return (
        <div className="flex flex-col h-screen w-screen bg-[#020817] text-white overflow-hidden font-mono antialiased">
            {/* Window Header (Dashboard Style) */}
            <div className="h-10 bg-black/40 border-b border-white/5 flex items-center justify-between px-4 shrink-0 backdrop-blur-md z-10 w-full">
                <div className="flex items-center gap-4">
                    {/* Traffic Lights */}
                    <div className="flex gap-1.5">
                        <div className="h-2.5 w-2.5 rounded-full bg-red-500/50 hover:bg-red-500 transition-colors shadow-sm" />
                        <div className="h-2.5 w-2.5 rounded-full bg-orange-500/50 hover:bg-orange-500 transition-colors shadow-sm" />
                        <div className="h-2.5 w-2.5 rounded-full bg-green-500/50 hover:bg-green-500 transition-colors shadow-sm" />
                    </div>
                    {/* Title */}
                    <div className="flex items-center gap-2 text-xs font-mono select-none">
                        <span className="text-zinc-300 font-bold">{pod}</span>
                        <span className="text-zinc-600">/</span>
                        <span className="text-zinc-400">{container || "default"}</span>
                        <span className="text-zinc-600">@</span>
                        <span className="text-zinc-500 max-w-[200px] truncate">{context}</span>
                    </div>
                </div>

                {/* Right Controls */}
                <div className="flex items-center gap-3">
                    <div className={`h-1.5 w-1.5 rounded-full transition-all duration-300 ${status === "connected" ? "bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.8)]" :
                        status === "connecting" ? "bg-amber-500 animate-pulse" :
                            "bg-red-500"
                        }`} />
                    <Button
                        variant="ghost"
                        size="sm"
                        className="h-6 text-[9px] font-bold uppercase tracking-widest px-2 text-zinc-500 hover:text-red-400 hover:bg-red-500/10 rounded-sm transition-colors"
                        onClick={() => {
                            if (wsRef.current) wsRef.current.close();
                            window.close();
                        }}
                    >
                        Terminate
                    </Button>
                </div>
            </div>

            {/* Terminal Area */}
            <div className="flex-1 bg-[#020817] p-0 relative">
                <div ref={terminalRef} className="h-full w-full p-4 overflow-hidden" />
            </div>
        </div>
    );
}

export default function ExecPage() {
    return (
        <Suspense fallback={
            <div className="flex items-center justify-center min-h-screen bg-[#020817] text-white">
                <div className="animate-pulse flex flex-col items-center gap-2">
                    <div className="h-2 w-24 bg-white/10 rounded"></div>
                    <span className="text-xs text-zinc-500 font-mono">INITIALIZING...</span>
                </div>
            </div>
        }>
            <ExecContent />
        </Suspense>
    );
}
