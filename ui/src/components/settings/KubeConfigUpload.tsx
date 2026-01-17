"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { api } from "@/lib/api";
import { Trash, Upload, FileCode } from "lucide-react";
import { formatDistanceToNow } from "date-fns";

interface KubeConfig {
    id: number;
    name: string;
    is_default: boolean;
    created_at: string;
}

export function KubeConfigUpload() {
    const [configs, setConfigs] = useState<KubeConfig[]>([]);
    const [uploading, setUploading] = useState(false);
    const [dragging, setDragging] = useState(false);

    useEffect(() => {
        fetchConfigs();
    }, []);

    const fetchConfigs = async () => {
        try {
            const data = await api.get<{ configs: KubeConfig[] }>("/settings/kube/configs");
            setConfigs(data.configs || []);
        } catch (error) {
            console.error("Failed to fetch configs:", error);
        }
    };

    const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
        const files = e.target.files;
        if (!files || files.length === 0) return;

        await uploadFile(files[0]);
    };

    const uploadFile = async (file: File) => {
        setUploading(true);
        const formData = new FormData();
        formData.append("file", file);

        try {
            // Using api.fetch directly to avoid JSON.stringify and explicit Content-Type
            const response = await api.fetch("/settings/kube/configs", {
                method: "POST",
                body: formData,
            });

            if (!response.ok) {
                throw new Error("Failed to upload");
            }

            fetchConfigs();
        } catch (error) {
            console.error("Failed to upload config:", error);
        } finally {
            setUploading(false);
        }
    };

    const handleDelete = async (id: number) => {
        if (!confirm("Are you sure you want to delete this configuration?")) return;

        try {
            await api.del(`/settings/kube/configs/${id}`);
            fetchConfigs();
        } catch (error) {
            console.error("Failed to delete config:", error);
        }
    };

    const handleDrop = async (e: React.DragEvent) => {
        e.preventDefault();
        setDragging(false);

        if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
            await uploadFile(e.dataTransfer.files[0]);
        }
    };

    return (
        <Card className="border-none shadow-2xl shadow-black/5 bg-card/50 backdrop-blur-sm overflow-hidden rounded-3xl">
            <CardHeader className="border-b bg-card/50 px-8 py-6">
                <CardTitle className="text-lg font-bold flex items-center gap-3">
                    <FileCode className="h-5 w-5 text-blue-500" />
                    Kubeconfig Files
                </CardTitle>
                <CardDescription>Upload additional kubeconfig files to merge into your environment.</CardDescription>
            </CardHeader>
            <CardContent className="p-8 space-y-6">
                {/* Upload Area */}
                <div
                    className={`
                        border-2 border-dashed rounded-2xl p-8 transition-all duration-200 text-center
                        ${dragging ? "border-primary bg-primary/5" : "border-muted-foreground/20 hover:border-primary/50 hover:bg-muted/30"}
                    `}
                    onDragOver={(e) => { e.preventDefault(); setDragging(true); }}
                    onDragLeave={() => setDragging(false)}
                    onDrop={handleDrop}
                >
                    <div className="flex flex-col items-center gap-3">
                        <div className="h-12 w-12 rounded-full bg-primary/10 flex items-center justify-center">
                            <Upload className="h-6 w-6 text-primary" />
                        </div>
                        <div className="space-y-1">
                            <p className="font-bold text-sm">Click to upload or drag and drop</p>
                            <p className="text-xs text-muted-foreground">Any valid kubeconfig file</p>
                        </div>
                        <input
                            type="file"
                            className="hidden"
                            id="kubeconfig-upload"
                            onChange={handleFileUpload}
                            disabled={uploading}
                        />
                        <Button
                            variant="outline"
                            size="sm"
                            className="mt-2 rounded-xl"
                            disabled={uploading}
                            asChild
                        >
                            <label htmlFor="kubeconfig-upload" className="cursor-pointer">
                                {uploading ? "Uploading..." : "Select File"}
                            </label>
                        </Button>
                    </div>
                </div>

                {/* File List */}
                <div className="space-y-3">
                    <Label className="text-xs font-bold uppercase tracking-wider opacity-60 ml-1">Uploaded Files</Label>

                    {configs.length === 0 ? (
                        <p className="text-sm text-muted-foreground italic pl-1">No custom configurations uploaded.</p>
                    ) : (
                        <div className="grid gap-3">
                            {configs.map((config) => (
                                <div
                                    key={config.id}
                                    className="flex items-center justify-between p-4 bg-muted/30 rounded-2xl border border-muted/20 group hover:border-primary/20 transition-all"
                                >
                                    <div className="flex items-center gap-4">
                                        <div className={`h-10 w-10 rounded-xl flex items-center justify-center ${config.is_default ? "bg-amber-500/10 text-amber-500" : "bg-blue-500/10 text-blue-500"}`}>
                                            <FileCode className="h-5 w-5" />
                                        </div>
                                        <div>
                                            <div className="flex items-center gap-2">
                                                <p className="font-bold text-sm">{config.name}</p>
                                                {config.is_default && (
                                                    <span className="px-2 py-0.5 rounded-full bg-amber-500/10 text-amber-600 text-[10px] font-bold uppercase tracking-wide">
                                                        Default
                                                    </span>
                                                )}
                                            </div>
                                            <p className="text-xs text-muted-foreground">
                                                Uploaded {formatDistanceToNow(new Date(config.created_at))} ago
                                            </p>
                                        </div>
                                    </div>

                                    {!config.is_default && (
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            className="h-10 w-10 text-muted-foreground hover:text-destructive hover:bg-destructive/10 rounded-xl opacity-0 group-hover:opacity-100 transition-all"
                                            onClick={() => handleDelete(config.id)}
                                        >
                                            <Trash className="h-4 w-4" />
                                        </Button>
                                    )}
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </CardContent>
        </Card>
    );
}
