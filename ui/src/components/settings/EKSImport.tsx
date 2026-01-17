"use client";

import React, { useState, useEffect } from "react";
import { eksApi, EKSCluster } from "@/lib/eksApi";

import { Check as CheckIcon, Plus as PlusIcon, X as XMarkIcon, Loader2 } from "lucide-react";

import { awsApi, AWSConfig } from "@/lib/awsApi";

export default function EKSImport() {
    const [accounts, setAccounts] = useState<AWSConfig[]>([]);
    const [selectedAccount, setSelectedAccount] = useState<string>("");
    const [regionOverride, setRegionOverride] = useState("");
    const [clusters, setClusters] = useState<EKSCluster[]>([]);
    const [selectedClusters, setSelectedClusters] = useState<Set<string>>(new Set());
    const [loading, setLoading] = useState(false);
    const [importing, setImporting] = useState(false);
    const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);

    useEffect(() => {
        loadAccounts();
    }, []);

    const loadAccounts = async () => {
        try {
            const res = await awsApi.listConfigs();
            setAccounts(res.configs);
            if (res.configs.length > 0) {
                setSelectedAccount(res.configs[0].id.toString());
            }
        } catch (err) {
            console.error("Failed to load AWS accounts", err);
        }
    };

    const fetchClusters = async () => {
        if (!selectedAccount) {
            setMessage({ type: 'error', text: "Please select an AWS account" });
            return;
        }
        setLoading(true);
        setMessage(null);
        setClusters([]);
        setSelectedClusters(new Set());
        try {
            const response = await eksApi.listClusters({
                aws_config_id: parseInt(selectedAccount),
                region: regionOverride || undefined,
            });
            setClusters(response.clusters);
            if (response.clusters.length === 0) {
                setMessage({ type: 'success', text: "No clusters found in this region" });
            }
        } catch (err: any) {
            setMessage({ type: 'error', text: err.response?.data?.error || "Failed to list clusters" });
        } finally {
            setLoading(false);
        }
    };

    const toggleClusterSelection = (clusterName: string) => {
        const newSelection = new Set(selectedClusters);
        if (newSelection.has(clusterName)) {
            newSelection.delete(clusterName);
        } else {
            newSelection.add(clusterName);
        }
        setSelectedClusters(newSelection);
    };

    const handleImport = async () => {
        if (selectedClusters.size === 0) return;
        setImporting(true);
        setMessage(null);
        try {
            const response = await eksApi.importClusters({
                aws_config_id: parseInt(selectedAccount),
                region: regionOverride || undefined,
                cluster_names: Array.from(selectedClusters),
            });
            setMessage({ type: 'success', text: response.message });
            setClusters([]); // Clear list after import? or keep?
            setSelectedClusters(new Set());
        } catch (err: any) {
            setMessage({ type: 'error', text: err.response?.data?.error || "Failed to import clusters" });
        } finally {
            setImporting(false);
        }
    };

    return (
        <div className="bg-white dark:bg-zinc-800 shadow rounded-lg p-6">
            <h2 className="text-xl font-semibold mb-4 text-zinc-900 dark:text-zinc-100">Import from AWS EKS</h2>

            <div className="bg-white dark:bg-zinc-900 rounded-lg shadow-sm border border-zinc-200 dark:border-zinc-800 p-6 space-y-4">
                <div className="grid gap-4 md:grid-cols-2">
                    <div>
                        <label htmlFor="awsAccount" className="block text-sm font-medium text-zinc-700 dark:text-zinc-300">
                            AWS Account
                        </label>
                        <select
                            id="awsAccount"
                            value={selectedAccount}
                            onChange={(e) => setSelectedAccount(e.target.value)}
                            className="mt-1 block w-full rounded-md border-zinc-300 dark:border-zinc-700 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 dark:bg-zinc-900 dark:text-white sm:text-sm p-2"
                        >
                            {accounts.map(acc => (
                                <option key={acc.id} value={acc.id}>{acc.name} ({acc.region})</option>
                            ))}
                        </select>
                        {accounts.length === 0 && (
                            <p className="text-xs text-amber-600 mt-1">No AWS accounts linked. Please link an account above.</p>
                        )}
                    </div>
                    <div>
                        <label htmlFor="region" className="block text-sm font-medium text-zinc-700 dark:text-zinc-300">
                            Region Override (Optional)
                        </label>
                        <input
                            type="text"
                            id="region"
                            value={regionOverride}
                            onChange={(e) => setRegionOverride(e.target.value)}
                            placeholder="Leave empty to use account region"
                            className="mt-1 block w-full rounded-md border-zinc-300 dark:border-zinc-700 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 dark:bg-zinc-900 dark:text-white sm:text-sm"
                        />
                    </div>
                </div>

                <div className="flex justify-end">
                    <button
                        onClick={fetchClusters}
                        disabled={loading || !selectedAccount}
                        className="flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white rounded-md text-sm font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        {loading && <Loader2 className="w-4 h-4 animate-spin" />}
                        List EKS Clusters
                    </button>
                </div>
            </div>

            {message && message.type === 'error' && (
                <div className="mb-4 rounded-md bg-red-50 dark:bg-red-900/20 p-4 mt-4">
                    <div className="flex">
                        <div className="flex-shrink-0">
                            <XMarkIcon className="h-5 w-5 text-red-400" aria-hidden="true" />
                        </div>
                        <div className="ml-3">
                            <h3 className="text-sm font-medium text-red-800 dark:text-red-200">{message.text}</h3>
                        </div>
                    </div>
                </div>
            )}

            {message && message.type === 'success' && (
                <div className="mb-4 rounded-md bg-green-50 dark:bg-green-900/20 p-4">
                    <div className="flex">
                        <div className="flex-shrink-0">
                            <CheckIcon className="h-5 w-5 text-green-400" aria-hidden="true" />
                        </div>
                        <div className="ml-3">
                            <h3 className="text-sm font-medium text-green-800 dark:text-green-200">{message.text}</h3>
                        </div>
                    </div>
                </div>
            )}

            {clusters.length > 0 && (
                <div className="mt-6">
                    <h3 className="text-lg font-medium leading-6 text-zinc-900 dark:text-zinc-100 mb-3">Available Clusters</h3>
                    <div className="bg-zinc-50 dark:bg-zinc-900/50 rounded-md border border-zinc-200 dark:border-zinc-700 overflow-hidden">
                        <ul className="divide-y divide-zinc-200 dark:divide-zinc-700">
                            {clusters.map((cluster) => (
                                <li key={cluster.name} className="flex items-center px-4 py-3 hover:bg-zinc-100 dark:hover:bg-zinc-800 transition">
                                    <input
                                        type="checkbox"
                                        checked={selectedClusters.has(cluster.name)}
                                        onChange={() => toggleClusterSelection(cluster.name)}
                                        className="h-4 w-4 rounded border-zinc-300 text-indigo-600 focus:ring-indigo-500"
                                    />
                                    <div className="ml-3 flex-1">
                                        <p className="text-sm font-medium text-zinc-900 dark:text-zinc-100">{cluster.name}</p>
                                        <p className="text-xs text-zinc-500 dark:text-zinc-400">Account: {cluster.account_id}</p>
                                    </div>
                                    <div className="text-sm text-zinc-500 dark:text-zinc-400">
                                        {cluster.region}
                                    </div>
                                </li>
                            ))}
                        </ul>
                    </div>
                    <div className="mt-4 flex justify-end">
                        <button
                            onClick={handleImport}
                            disabled={selectedClusters.size === 0 || importing}
                            className="inline-flex items-center rounded-md border border-transparent bg-green-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2 disabled:opacity-50"
                        >
                            {importing ? <Loader2 className="w-4 h-4 animate-spin mr-2" /> : <PlusIcon className="-ml-1 mr-2 h-5 w-5" aria-hidden="true" />}
                            Import Selected ({selectedClusters.size})
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
}
