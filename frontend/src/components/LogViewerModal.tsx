"use client";

import { useEffect, useRef, useState } from "react";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import "@xterm/xterm/css/xterm.css";
import { API_URL } from "@/lib/config";
import { Terminal as TerminalIcon, Loader2, WrapText, Clock, Tag } from "lucide-react";

export function LogViewerModal({
    isOpen,
    onClose,
    context,
    namespace,
    pod,
    containers,
    initContainers = [],
}: {
    isOpen: boolean;
    onClose: () => void;
    context: string;
    namespace: string;
    pod: string;
    containers: string[];
    initContainers?: string[];
}) {
    const allContainers = [...initContainers, ...containers];
    const terminalRef = useRef<HTMLDivElement>(null);
    const wsRef = useRef<WebSocket | null>(null);
    const xtermRef = useRef<Terminal | null>(null);
    const fitAddonRef = useRef<FitAddon | null>(null);
    const [status, setStatus] = useState<"connecting" | "connected" | "disconnected" | "error">("connecting");
    const [isWrapEnabled, setIsWrapEnabled] = useState(false);
    const [showTimestamps, setShowTimestamps] = useState(true);
    const [showPrefix, setShowPrefix] = useState(allContainers.length > 1);
    const [selectedContainers, setSelectedContainers] = useState<string[]>(["__all__"]);
    const [isDropdownOpen, setIsDropdownOpen] = useState(false);

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

            if (selectedContainers.length === 0) {
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
            connectWebSocket(term, containerParam);
        }, 100);

        function connectWebSocket(term: Terminal, containerParam: string) {
            // ... (rest of function) ...
            // Close existing socket if any
            if (wsRef.current) {
                wsRef.current.close();
                wsRef.current = null;
            }

            // Connect WebSocket
            const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
            let wsHost = window.location.host;
            if (API_URL.startsWith("http")) {
                const url = new URL(API_URL);
                wsHost = url.host;
            }

            const wsUrl = `${protocol}//${wsHost}/api/v1/kube/logs?context=${context}&namespace=${namespace}&pod=${pod}&container=${containerParam}&timestamps=${showTimestamps}&prefix=${showPrefix}`;

            setStatus("connecting");
            term.writeln(`\x1b[33mConnecting to logs for ${pod} [${containerParam === "__all__" ? "All Containers" : containerParam}]...\x1b[0m`);

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

    }, [isOpen, context, namespace, pod, selectedContainers, showTimestamps, showPrefix]); // Re-run when container changes

    const toggleContainer = (c: string) => {
        if (c === "__all__") {
            if (selectedContainers.includes("__all__")) {
                setSelectedContainers([]);
            } else {
                setSelectedContainers(["__all__"]);
            }
        } else {
            // If "All" is currently selected, clicking a specific container 
            // should Deselect All and Select ONLY the clicked one (Focus Mode).
            if (selectedContainers.includes("__all__")) {
                setSelectedContainers([c]);
                return;
            }

            let newSelection = [...selectedContainers];
            if (newSelection.includes(c)) {
                newSelection = newSelection.filter(x => x !== c);
            } else {
                newSelection.push(c);
            }

            // If all individual containers are selected, switch back to "__all__"
            if (newSelection.length === allContainers.length) {
                newSelection = ["__all__"];
            }

            setSelectedContainers(newSelection);
        }
    };

    const isContainerSelected = (c: string) => {
        return selectedContainers.includes("__all__") || selectedContainers.includes(c);
    };

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
                            {namespace}
                        </DialogDescription>
                    </div>

                    <div className="flex items-center gap-2">
                        <div className="relative">
                            <button
                                onClick={() => setIsDropdownOpen(!isDropdownOpen)}
                                className="h-8 pl-3 pr-2 text-xs font-mono bg-zinc-800 border-zinc-700 text-zinc-200 rounded-md hover:bg-zinc-700/50 transition-colors flex items-center gap-2"
                            >
                                {selectedContainers.includes("__all__") ? "All Containers" : `${selectedContainers.length} Container${selectedContainers.length > 1 ? 's' : ''}`}
                                <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={`transition-transform ${isDropdownOpen ? 'rotate-180' : ''}`}>
                                    <path d="m6 9 6 6 6-6" />
                                </svg>
                            </button>

                            {isDropdownOpen && (
                                <>
                                    <div className="fixed inset-0 z-40" onClick={() => setIsDropdownOpen(false)} />
                                    <div className="absolute right-0 top-full mt-1 w-56 bg-zinc-900 border border-zinc-700 rounded-lg shadow-xl z-50 p-1 flex flex-col gap-0.5">
                                        <button
                                            onClick={() => toggleContainer("__all__")}
                                            className={`flex items-center gap-2 w-full px-2 py-1.5 text-xs font-mono transition-colors ${selectedContainers.includes("__all__") ? 'bg-primary/20 text-primary' : 'text-zinc-400 hover:bg-white/5'}`}
                                        >
                                            <div className={`w-3.5 h-3.5 border rounded-sm flex items-center justify-center ${selectedContainers.includes("__all__") ? 'border-primary bg-primary' : 'border-zinc-600'}`}>
                                                {selectedContainers.includes("__all__") && <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="4" strokeLinecap="round" strokeLinejoin="round" className="text-black"><polyline points="20 6 9 17 4 12" /></svg>}
                                            </div>
                                            All Containers
                                        </button>

                                        {/* Standard Containers Section */}
                                        {containers.length > 0 && (
                                            <>
                                                <div className="px-2 py-1 mt-1 text-[10px] font-bold text-zinc-500 uppercase tracking-wider">Containers</div>
                                                {containers.map(c => {
                                                    const isSelected = isContainerSelected(c);
                                                    return (
                                                        <button
                                                            key={c}
                                                            onClick={() => toggleContainer(c)}
                                                            className={`flex items-center gap-2 w-full px-2 py-1.5 text-xs font-mono transition-colors ${isSelected ? 'bg-primary/20 text-primary' : 'text-zinc-400 hover:bg-white/5'}`}
                                                        >
                                                            <div className={`w-3.5 h-3.5 border rounded-sm flex items-center justify-center ${isSelected ? 'border-primary bg-primary' : 'border-zinc-600'}`}>
                                                                {isSelected && <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="4" strokeLinecap="round" strokeLinejoin="round" className="text-black"><polyline points="20 6 9 17 4 12" /></svg>}
                                                            </div>
                                                            {c}
                                                        </button>
                                                    );
                                                })}
                                            </>
                                        )}

                                        {/* Init Containers Section */}
                                        {initContainers && initContainers.length > 0 && (
                                            <>
                                                <div className="h-px bg-zinc-800 my-1 mt-2" />
                                                <div className="px-2 py-1 text-[10px] font-bold text-zinc-500 uppercase tracking-wider">Init Containers</div>
                                                {initContainers.map(c => {
                                                    const isSelected = isContainerSelected(c);
                                                    return (
                                                        <button
                                                            key={c}
                                                            onClick={() => toggleContainer(c)}
                                                            className={`flex items-center gap-2 w-full px-2 py-1.5 text-xs font-mono transition-colors ${isSelected ? 'bg-primary/20 text-primary' : 'text-zinc-400 hover:bg-white/5'}`}
                                                        >
                                                            <div className={`w-3.5 h-3.5 border rounded-sm flex items-center justify-center ${isSelected ? 'border-primary bg-primary' : 'border-zinc-600'}`}>
                                                                {isSelected && <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="4" strokeLinecap="round" strokeLinejoin="round" className="text-black"><polyline points="20 6 9 17 4 12" /></svg>}
                                                            </div>
                                                            {c}
                                                        </button>
                                                    );
                                                })}
                                            </>
                                        )}
                                    </div>
                                </>
                            )}
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
                    {selectedContainers.length === 0 && (
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
