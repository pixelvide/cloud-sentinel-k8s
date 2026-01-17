"use client";

import { useEffect, useRef, useState } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import "@xterm/xterm/css/xterm.css";
import { api } from "@/lib/api";
import { Terminal as TerminalIcon, Loader2, WrapText, Clock, Tag } from "lucide-react";
import { MultiSelect } from "@/components/ui/multi-select";

export function LogViewerModal({
    isOpen,
    onClose,
    context,
    namespace,
    containers,
    initContainers = [],
    selector,
    pods = [],
    showPodSelector,
    title,
}: {
    isOpen: boolean;
    onClose: () => void;
    context: string;
    namespace: string;
    containers: string[];
    initContainers?: string[];
    selector?: string;
    pods?: Array<{ name: string, status: string }>;
    showPodSelector?: boolean;
    title?: string;
}) {
    const allContainers = [...(initContainers || []), ...(containers || [])];
    const terminalRef = useRef<HTMLDivElement>(null);
    const wsRef = useRef<WebSocket | null>(null);
    const xtermRef = useRef<Terminal | null>(null);
    const fitAddonRef = useRef<FitAddon | null>(null);
    const [status, setStatus] = useState<"connecting" | "connected" | "disconnected" | "error">("connecting");
    const [isWrapEnabled, setIsWrapEnabled] = useState(false);
    const [showTimestamps, setShowTimestamps] = useState(false);
    // Show prefix by default if multiple containers OR if pod selector is shown (multi-pod view)
    const [showPrefix, setShowPrefix] = useState(allContainers.length > 1 || (showPodSelector ?? false));
    const [selectedContainers, setSelectedContainers] = useState<string[]>(["__all__"]);
    // Smart default: if pod selector is hidden and we have exactly one pod, use that pod name
    // Otherwise default to __all__ for multi-pod views
    const [selectedPods, setSelectedPods] = useState<string[]>(() => {
        if (showPodSelector === false && pods && pods.length === 1) {
            return [pods[0].name];
        }
        return ["__all__"];
    });

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

            // Dispose existing terminal if we are re-initializing (e.g. container change)
            if (xtermRef.current) {
                xtermRef.current.dispose();
            }

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


            if (selectedContainers.length === 0 || selectedPods.length === 0) {
                if (wsRef.current) {
                    wsRef.current.close();
                    wsRef.current = null;
                }
                setStatus("disconnected");
                if (xtermRef.current) {
                    xtermRef.current.clear();
                }
                return;
            }

            const containerParam = selectedContainers.join(",");
            const podParam = selectedPods.includes("__all__") ? "__all__" : selectedPods.join(",");
            connectWebSocket(term, podParam, containerParam);
        }, 100);

        function connectWebSocket(term: Terminal, podParam: string, containerParam: string) {
            // ... (rest of function) ...
            // Close existing socket if any
            if (wsRef.current) {
                wsRef.current.close();
                wsRef.current = null;
            }

            const selectorParam = selector ? `&selector=${encodeURIComponent(selector)}` : "";
            const wsUrl = api.getWsUrl(`/api/v1/kube/logs?context=${context}&namespace=${namespace}&pod=${podParam}&container=${containerParam}&timestamps=${showTimestamps}&prefix=${showPrefix}${selectorParam}`);

            setStatus("connecting");
            const podDisplay = podParam === "__all__" ? "All Pods" : (selectedPods.length > 1 ? `${selectedPods.length} Pods` : podParam);
            const containerDisplay = containerParam === "__all__" ? "All Containers" : (selectedContainers.length > 1 ? `${selectedContainers.length} Containers` : containerParam);
            term.writeln(`\x1b[33mConnecting to logs for ${podDisplay} [${containerDisplay}]...\x1b[0m`);

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
            // Logic to dispose xterm is handled at start of effect or on modal close
        };

    }, [isOpen, context, namespace, selectedContainers, selectedPods, showTimestamps, showPrefix, selector]); // Re-run when container changes


    // Pod selection handlers
    const handlePodSelection = (newSelection: string[]) => {
        setSelectedPods(newSelection);

        // Auto-enable prefix when multiple pods are selected
        if (newSelection.includes("__all__") || newSelection.length > 1) {
            setShowPrefix(true);
        }
    };

    return (
        <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
            <DialogContent className="sm:max-w-6xl h-[90vh] bg-[#09090b] border-zinc-800 p-0 flex flex-col gap-0 overflow-hidden dark">
                <DialogHeader className="p-4 border-b border-zinc-800 bg-zinc-900/50 flex flex-row items-center justify-between space-y-0 text-left">
                    <div className="flex flex-col gap-1">
                        <DialogTitle className="flex items-center gap-2 text-sm font-mono text-zinc-200">
                            <span>Logs: {title || "Unknown Resource"}</span>
                            {status === "connecting" && <Loader2 className="h-3 w-3 animate-spin text-yellow-500" />}
                            {status === "connected" && <div className="h-2 w-2 rounded-full bg-green-500 animate-pulse" />}
                            {status === "disconnected" && <div className="h-2 w-2 rounded-full bg-zinc-500" />}
                            {status === "error" && <div className="h-2 w-2 rounded-full bg-red-500" />}
                        </DialogTitle>
                        <DialogDescription className="text-xs font-mono text-zinc-500">
                            {namespace}
                        </DialogDescription>
                    </div>

                    <div className="flex items-center gap-2">
                        {/* Pod Selector - Only show if multiple pods available */}
                        {(showPodSelector ?? (pods && pods.length > 0)) && (
                            <div className="w-48 overflow-hidden">
                                <MultiSelect
                                    options={(pods || []).map(p => ({ value: p.name, label: p.name }))}
                                    selected={selectedPods}
                                    onChange={handlePodSelection}
                                    placeholder="Select Pods"
                                    showSearch={false}
                                    allOption={{ label: "All Pods", value: "__all__" }}
                                    popoverClassName="dark"
                                />
                            </div>
                        )}

                        {/* Container Selector */}
                        <div className="w-48 overflow-hidden">
                            <MultiSelect
                                groups={[
                                    {
                                        label: "Containers",
                                        options: containers.map(c => ({ value: c, label: c }))
                                    },
                                    {
                                        label: "Init Containers",
                                        options: (initContainers || []).map(c => ({ value: c, label: c }))
                                    }
                                ].filter(g => g.options.length > 0)}
                                selected={selectedContainers}
                                onChange={setSelectedContainers}
                                placeholder="Select Containers"
                                showSearch={false}
                                allOption={{ label: "All Containers", value: "__all__" }}
                                popoverClassName="dark"
                            />
                        </div>

                        <div className="h-4 w-px bg-zinc-800 mx-1" />

                        <button
                            onClick={() => setShowPrefix(!showPrefix)}
                            className={`p-2 rounded-md transition-colors hover:bg-white/10 ${showPrefix ? 'text-primary' : 'text-zinc-500'}`}
                            title={showPrefix ? "Hide Container Names" : "Show Container Names"}
                        >
                            <Tag className="h-4 w-4" />
                        </button>

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
                    {selectedPods.length === 0 && (
                        <div className="absolute inset-0 z-10 flex flex-col items-center justify-center bg-[#09090b]/80 backdrop-blur-sm text-zinc-400">
                            <TerminalIcon className="h-12 w-12 mb-4 opacity-20" />
                            <p className="text-sm font-mono">Select at least one pod</p>
                            <p className="text-xs text-zinc-600 mt-2">Use the pod selector above to view logs</p>
                        </div>
                    )}
                    {selectedContainers.length === 0 && selectedPods.length > 0 && (
                        <div className="absolute inset-0 z-10 flex flex-col items-center justify-center bg-[#09090b]/80 backdrop-blur-sm text-zinc-400">
                            <Tag className="h-12 w-12 mb-4 opacity-20" />
                            <p className="text-sm font-mono">Select at least one container</p>
                            <p className="text-xs text-zinc-600 mt-2">Use the dropdown above to view logs</p>
                        </div>
                    )}
                    <div ref={terminalRef} className="absolute inset-2 overflow-x-auto" />
                </div>
            </DialogContent>
        </Dialog>
    );
}
