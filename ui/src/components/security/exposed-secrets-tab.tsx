import { useQuery } from "@tanstack/react-query"
import { ExternalLink, Key, ShieldAlert, ShieldCheck } from "lucide-react"
import { useState } from "react"
import { securityApi, ExposedSecretReport, ExposedSecret } from "@/api/security"
import { Badge } from "@/components/ui/badge"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"

interface ExposedSecretsTabProps {
    namespace?: string
    kind: string
    name: string
}

export function ExposedSecretsTab({ namespace, kind, name }: ExposedSecretsTabProps) {
    const [searchTerm, setSearchTerm] = useState("")

    const { data: status } = useQuery({
        queryKey: ["security", "status"],
        queryFn: () => securityApi.getStatus(),
    })

    const { data: reports, isLoading } = useQuery({
        queryKey: ["security", "secrets", namespace, kind, name],
        queryFn: () => securityApi.getExposedSecretReports(namespace, kind, name),
        enabled: !!status?.trivyInstalled,
    })

    if (status && !status.trivyInstalled) {
        return (
            <div className="p-6">
                <Alert variant="default" className="bg-blue-50 dark:bg-blue-950/30 border-blue-200 dark:border-blue-800">
                    <ShieldAlert className="h-5 w-5 text-blue-600 dark:text-blue-400" />
                    <AlertTitle className="text-blue-800 dark:text-blue-300 ml-2">Trivy Operator Not Installed</AlertTitle>
                    <AlertDescription className="text-blue-700 dark:text-blue-400 ml-2 mt-2">
                        <p className="mb-4">
                            Secret scanning requires the Trivy Operator to be installed in your cluster.
                        </p>
                        <Button variant="outline" className="border-blue-300 dark:border-blue-700 text-blue-700 dark:text-blue-300 hover:bg-blue-100 dark:hover:bg-blue-900/50" asChild>
                            <a href="https://aquasecurity.github.io/trivy-operator/latest/getting-started/installation/" target="_blank" rel="noreferrer">
                                Installation Guide <ExternalLink className="ml-2 h-4 w-4" />
                            </a>
                        </Button>
                    </AlertDescription>
                </Alert>
            </div>
        )
    }

    if (isLoading) {
        return <div className="p-6 text-center text-muted-foreground">Loading exposed secret reports...</div>
    }

    if (!reports?.items || reports.items.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center p-12 text-center border rounded-lg bg-muted/30">
                <div className="bg-secondary p-3 rounded-full mb-4">
                    <ShieldCheck className="h-8 w-8 text-muted-foreground" />
                </div>
                <h3 className="text-lg font-semibold text-foreground">No Exposed Secrets Found</h3>
                <p className="text-sm text-muted-foreground mt-2 max-w-sm">
                    Trivy Operator hasn't detected any exposed secrets in this resource's images.
                </p>
            </div>
        )
    }

    // Aggregate all secrets from all reports
    const allSecrets = reports.items.flatMap((report: ExposedSecretReport) =>
        (report.report.secrets || []).map((secret: ExposedSecret) => ({
            ...secret,
            image: report.report.artifact?.repository + ":" + report.report.artifact?.tag,
            reportName: report.metadata.name
        }))
    )

    // Filter secrets
    const filteredSecrets = allSecrets.filter((secret: ExposedSecret & { image: string }) =>
        secret.ruleID?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        secret.title?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        secret.target?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        secret.category?.toLowerCase().includes(searchTerm.toLowerCase())
    )

    // Calculate summary
    const summary = {
        total: allSecrets.length,
        critical: allSecrets.filter((s: ExposedSecret) => s.severity === "CRITICAL").length,
        high: allSecrets.filter((s: ExposedSecret) => s.severity === "HIGH").length,
        medium: allSecrets.filter((s: ExposedSecret) => s.severity === "MEDIUM").length,
        low: allSecrets.filter((s: ExposedSecret) => s.severity === "LOW").length,
    }

    const getSeverityColor = (severity: string) => {
        switch (severity?.toUpperCase()) {
            case "CRITICAL": return "bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-300 border-red-200 dark:border-red-800"
            case "HIGH": return "bg-orange-100 dark:bg-orange-900/30 text-orange-800 dark:text-orange-300 border-orange-200 dark:border-orange-800"
            case "MEDIUM": return "bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-300 border-yellow-200 dark:border-yellow-800"
            case "LOW": return "bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-300 border-blue-200 dark:border-blue-800"
            default: return "bg-gray-100 dark:bg-gray-800 text-gray-800 dark:text-gray-300 border-gray-200 dark:border-gray-700"
        }
    }

    return (
        <div className="space-y-6">
            {/* Summary Cards */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Exposed Secrets</CardTitle>
                        <Key className="h-4 w-4 text-red-600 dark:text-red-400" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-red-600 dark:text-red-400">{summary.total}</div>
                        <p className="text-xs text-muted-foreground">
                            {summary.critical} critical, {summary.high} high
                        </p>
                    </CardContent>
                </Card>
            </div>

            {summary.total > 0 && (
                <Alert variant="destructive" className="bg-red-50 dark:bg-red-950/30 border-red-200 dark:border-red-800">
                    <Key className="h-5 w-5" />
                    <AlertTitle>Secrets Detected!</AlertTitle>
                    <AlertDescription>
                        Secrets have been detected in container images. These should be removed and rotated immediately.
                    </AlertDescription>
                </Alert>
            )}

            {/* Search */}
            <div className="flex items-center space-x-2">
                <Input
                    placeholder="Search by rule ID, title, target, or category..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="max-w-sm"
                />
            </div>

            {/* Secrets Table */}
            <div className="rounded-md border">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead className="w-[100px]">Severity</TableHead>
                            <TableHead className="w-[150px]">Rule ID</TableHead>
                            <TableHead>Title</TableHead>
                            <TableHead>Target</TableHead>
                            <TableHead className="w-[120px]">Category</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {filteredSecrets.map((secret: ExposedSecret & { image: string; reportName: string }, i: number) => (
                            <TableRow key={secret.ruleID + i}>
                                <TableCell>
                                    <Badge variant="outline" className={getSeverityColor(secret.severity)}>
                                        {secret.severity}
                                    </Badge>
                                </TableCell>
                                <TableCell className="font-mono text-sm">{secret.ruleID}</TableCell>
                                <TableCell>
                                    <div className="flex flex-col">
                                        <span className="font-medium">{secret.title}</span>
                                        {secret.match && (
                                            <span className="text-xs text-muted-foreground mt-1 font-mono max-w-md truncate" title={secret.match}>
                                                {secret.match.substring(0, 50)}...
                                            </span>
                                        )}
                                    </div>
                                </TableCell>
                                <TableCell className="font-mono text-xs max-w-xs truncate" title={secret.target}>
                                    {secret.target}
                                </TableCell>
                                <TableCell>
                                    <Badge variant="secondary">{secret.category}</Badge>
                                </TableCell>
                            </TableRow>
                        ))}
                        {filteredSecrets.length === 0 && (
                            <TableRow>
                                <TableCell colSpan={5} className="h-24 text-center">
                                    No exposed secrets found matching your search.
                                </TableCell>
                            </TableRow>
                        )}
                    </TableBody>
                </Table>
            </div>
        </div>
    )
}
