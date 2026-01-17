"use client";

import { PageHeader } from "@/components/PageHeader";
import EKSImport from "@/components/settings/EKSImport";

import AWSAccountManagement from "@/components/settings/AWSAccountManagement";

export default function EKSSettingsPage() {
    return (
        <div className="flex flex-col h-full bg-zinc-50 dark:bg-zinc-950">
            <PageHeader
                title="AWS EKS Settings"
                description="Manage AWS EKS cluster connections"
            />

            <div className="p-6 space-y-8 max-w-5xl mx-auto w-full">
                <AWSAccountManagement />
                <div className="border-t border-zinc-200 dark:border-zinc-800 pt-8">
                    <h2 className="text-lg font-semibold text-zinc-900 dark:text-zinc-100 mb-4">Import Clusters</h2>
                    <EKSImport />
                </div>
            </div>
        </div>
    );
}
