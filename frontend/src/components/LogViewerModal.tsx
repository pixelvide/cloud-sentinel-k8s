"use client";

import { useEffect, useRef, useState } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import "@xterm/xterm/css/xterm.css";
import { API_URL } from "@/lib/config";
import { Terminal as TerminalIcon, Loader2, WrapText, Clock } from "lucide-react";

interface LogViewerModalProps {
    isOpen: boolean;
    onClose: () => void;
    context: string;
    namespace: string;
    pod: string;
    container: string;
}

export function LogViewerModal({
    isOpen,
    onClose,
    context,
    namespace,
    pod,
    container,
}: LogViewerModalProps) {
    const terminalRef = useRef<HTMLDivElement>(null);
    const wsRef = useRef<WebSocket | null>(null);
    const xtermRef = useRef<Terminal | null>(null);
    const fitAddonRef = useRef<FitAddon | null>(null);
    const [status, setStatus] = useState<"connecting" | "connected" | "disconnected" | "error">("connecting");
    const [isWrapEnabled, setIsWrapEnabled] = useState(false);
    const [showTimestamps, setShowTimestamps] = useState(true);

    // Handle Resize Logic
    const handleResize = () => {
        if (!fitAddonRef.current || !xtermRef.current) return;

        if (isWrapEnabled) {
            // Fit to container width (standard wrapping)
            fitAddonRef.current.fit();
        } else {
            // "No Wrap" mode: Set massive column width, restrict rows to container height
            const dims = fitAddonRef.current.proposeDimensions();
            if (dims) {
                // 1000 cols is usually enough to prevent wrap for reasonable logs. 
                // We keep rows dynamic to fill vertical space.
                xtermRef.current.resize(1000, dims.rows);
            }
        }
    };

    // Re-run resize when wrap state toggles
    useEffect(() => {
        handleResize();
    }, [isWrapEnabled]);

    useEffect(() => {
        if (!isOpen) {
            // Cleanup on close
            if (wsRef.current) {
                wsRef.current.close();
                wsRef.current = null;
            }
            if (xtermRef.current) {
                xtermRef.current.dispose();
                xtermRef.current = null;
            }
            return;
        }

        // Initialize xterm inside timeout
        const initTimeout = setTimeout(() => {
            if (!terminalRef.current) return;

            const term = new Terminal({
                cursorBlink: true,
                fontSize: 12,
                fontFamily: 'Menlo, Monaco, "Courier New", monospace',
                theme: {
                    background: '#09090b',
                    foreground: '#f4f4f5',
                },
                disableStdin: true,
                convertEol: true, // Help with line endings if needed
            });

            const fitAddon = new FitAddon();
            term.loadAddon(fitAddon);

            term.open(terminalRef.current);

            xtermRef.current = term;
            fitAddonRef.current = fitAddon;

            // Initial Resize
            // Wait a frame for layout
            requestAnimationFrame(() => {
                handleResize();
            });

            connectWebSocket(term);
        }, 100);

        function connectWebSocket(term: Terminal) {
            // Connect WebSocket
            const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
            let wsHost = window.location.host;
            if (API_URL.startsWith("http")) {
                const url = new URL(API_URL);
                wsHost = url.host;
            }

            const wsUrl = `${protocol}//${wsHost}/api/v1/kube/logs?context=${context}&namespace=${namespace}&pod=${pod}&container=${container}&timestamps=${showTimestamps}`;

            setStatus("connecting");
            term.writeln(`\x1b[33mConnecting to logs for ${pod}...\x1b[0m`);

            const ws = new WebSocket(wsUrl);
            wsRef.current = ws;

            ws.onopen = () => {
                setStatus("connected");
                term.writeln(`\x1b[32mConnected.\x1b[0m`);
            };

            ws.onmessage = (event) => {
                // Ensure data is string
                if (typeof event.data === 'string') {
                    term.write(event.data);
                } else {
                    // If blob, read it (rare for textmessage, but safe to handle)
                    const reader = new FileReader();
                    reader.onload = () => {
                        term.write(reader.result as string);
                    };
                    reader.readAsText(event.data);
                }
            };

            ws.onerror = (error) => {
                setStatus("error");
                console.error("WS Error:", error);
                term.writeln(`\r\n\x1b[31mConnection Error\x1b[0m`);
            };

            ws.onclose = () => {
                if (status !== "error") {
                    setStatus("disconnected");
                    term.writeln(`\r\n\x1b[33mConnection Closed\x1b[0m`);
                }
            };
        }

        // Resize observer
        const resizeObserver = new ResizeObserver(() => {
            handleResize();
        });
        if (terminalRef.current) {
            resizeObserver.observe(terminalRef.current!);
        }

        return () => {
            resizeObserver.disconnect();
            clearTimeout(initTimeout);
            if (wsRef.current) {
                wsRef.current.close();
            }
            if (xtermRef.current) {
                xtermRef.current.dispose();
            }
        };

    }, [isOpen, context, namespace, pod, container, showTimestamps]); // Intentionally exclude isWrapEnabled from re-init

    return (
        <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
            <DialogContent className="sm:max-w-4xl h-[80vh] bg-[#09090b] border-zinc-800 p-0 flex flex-col gap-0 overflow-hidden">
                <DialogHeader className="p-4 border-b border-zinc-800 bg-zinc-900/50 flex flex-row items-center justify-between space-y-0 text-left">
                    <div className="flex flex-col gap-1">
                        <DialogTitle className="flex items-center gap-2 text-sm font-mono text-zinc-200">
                            <span>Logs: {pod}</span>
                            {status === "connecting" && <Loader2 className="h-3 w-3 animate-spin text-yellow-500" />}
                            {status === "connected" && <div className="h-2 w-2 rounded-full bg-green-500 animate-pulse" />}
                            {status === "disconnected" && <div className="h-2 w-2 rounded-full bg-zinc-500" />}
                            {status === "error" && <div className="h-2 w-2 rounded-full bg-red-500" />}
                        </DialogTitle>
                        <DialogDescription className="text-xs font-mono text-zinc-500">
                            {namespace} / {container}
                        </DialogDescription>
                    </div>

                    <div className="flex items-center gap-1">
                        <button
                            onClick={() => setShowTimestamps(!showTimestamps)}
                            className={`p-2 rounded-md transition-colors hover:bg-white/10 ${showTimestamps ? 'text-primary' : 'text-zinc-500'}`}
                            title={showTimestamps ? "Hide Timestamps" : "Show Timestamps"}
                        >
                            <Clock className="h-4 w-4" />
                        </button>
                        <button
                            onClick={() => setIsWrapEnabled(!isWrapEnabled)}
                            className={`p-2 rounded-md transition-colors hover:bg-white/10 ${isWrapEnabled ? 'text-primary' : 'text-zinc-500'}`}
                            title={isWrapEnabled ? "Disable Wrap" : "Enable Wrap"}
                        >
                            <WrapText className="h-4 w-4" />
                        </button>
                    </div>
                </DialogHeader>
                <div className="flex-1 w-full relative bg-[#09090b] p-2 overflow-hidden">
                    {/* 
                         Container for xterm. overflow-x-auto allows scrolling when nowrap implies wide terminal.
                         We override standard xterm overflow behavior if needed contextually, but usually
                         xterm manages its own scrollbar for rows. 
                         For horizontal scroll, we rely on this parent div if the canvas is huge.
                     */}
                    <div ref={terminalRef} className="absolute inset-2 overflow-x-auto" />
                </div>
            </DialogContent>
        </Dialog>
    );
}
