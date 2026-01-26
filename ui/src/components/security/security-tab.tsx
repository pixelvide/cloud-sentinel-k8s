import { useQuery } from "@tanstack/react-query"
import { ExternalLink, ShieldAlert, ShieldCheck } from "lucide-react"
import { useState } from "react"
import { securityApi, VulnerabilityReport, Vulnerability } from "@/api/security"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
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

interface SecurityTabProps {
    namespace?: string
    kind: string
    name: string
}

export function SecurityTab({ namespace, kind, name }: SecurityTabProps) {
    const [searchTerm, setSearchTerm] = useState("")

    const { data: status } = useQuery({
        queryKey: ["security", "status"],
        queryFn: () => securityApi.getStatus(),
    })

    // Map K8s kind to Trivy resource kind (usually the same, but case sensitivity might matter)
    // Trivy Operator typically uses the exact Kind string found in metadata (e.g. "Pod", "Deployment")
    const { data: reports, isLoading } = useQuery({
        queryKey: ["security", "reports", namespace, kind, name],
        queryFn: () => securityApi.getReports(namespace, kind, name),
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
                            Security scanning requires the Trivy Operator to be installed in your cluster.
                            It automatically scans your workloads for vulnerabilities.
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
        return <div className="p-6 text-center text-muted-foreground">Loading security reports...</div>
    }

    if (!reports?.items || reports.items.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center p-12 text-center border rounded-lg bg-muted/30">
                <div className="bg-secondary p-3 rounded-full mb-4">
                    <ShieldCheck className="h-8 w-8 text-muted-foreground" />
                </div>
                <h3 className="text-lg font-semibold text-foreground">No Security Reports Found</h3>
                <p className="text-sm text-muted-foreground mt-2 max-w-sm">
                    Trivy Operator hasn't scanned this resource yet, or no vulnerabilities were found.
                    Check if the operator is running and configured to scan this namespace.
                </p>
            </div>
        )
    }

    const allVulnerabilities = reports.items.flatMap((report: VulnerabilityReport) =>
        report.report.vulnerabilities.map((v: Vulnerability) => ({ ...v, image: report.report.artifact.repository + ":" + report.report.artifact.tag }))
    )

    const filteredVulns = allVulnerabilities.filter((v: Vulnerability & { image: string }) =>
        v.vulnerabilityID.toLowerCase().includes(searchTerm.toLowerCase()) ||
        v.title?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        v.resource.toLowerCase().includes(searchTerm.toLowerCase())
    )

    const getSeverityColor = (severity: string) => {
        switch (severity) {
            case "CRITICAL": return "bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-300 border-red-200 dark:border-red-800"
            case "HIGH": return "bg-orange-100 dark:bg-orange-900/30 text-orange-800 dark:text-orange-300 border-orange-200 dark:border-orange-800"
            case "MEDIUM": return "bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-300 border-yellow-200 dark:border-yellow-800"
            case "LOW": return "bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-300 border-blue-200 dark:border-blue-800"
            default: return "bg-gray-100 dark:bg-gray-800 text-gray-800 dark:text-gray-300 border-gray-200 dark:border-gray-700"
        }
    }

    return (
        <div className="space-y-6">
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                {reports.items.map((report: VulnerabilityReport) => (
                    <Card key={report.metadata.name}>
                        <CardHeader className="p-4 pb-2">
                            <CardTitle className="text-sm font-medium truncate" title={report.report.artifact.repository}>
                                {report.report.artifact.repository}
                            </CardTitle>
                            <CardDescription className="text-xs truncate" title={report.report.artifact.tag}>
                                Tag: {report.report.artifact.tag}
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="p-4 pt-2">
                            <div className="flex justify-between text-xs mt-2">
                                <div className="flex flex-col items-center">
                                    <span className="font-bold text-red-600 dark:text-red-400">{report.report.summary.criticalCount}</span>
                                    <span className="text-muted-foreground">Crit</span>
                                </div>
                                <div className="flex flex-col items-center">
                                    <span className="font-bold text-orange-600 dark:text-orange-400">{report.report.summary.highCount}</span>
                                    <span className="text-muted-foreground">High</span>
                                </div>
                                <div className="flex flex-col items-center">
                                    <span className="font-bold text-yellow-600 dark:text-yellow-400">{report.report.summary.mediumCount}</span>
                                    <span className="text-muted-foreground">Med</span>
                                </div>
                                <div className="flex flex-col items-center">
                                    <span className="font-bold text-blue-600 dark:text-blue-400">{report.report.summary.lowCount}</span>
                                    <span className="text-muted-foreground">Low</span>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                ))}
            </div>

            <div className="flex items-center space-x-2">
                <Input
                    placeholder="Search CVE, package, or description..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="max-w-sm"
                />
            </div>

            <div className="rounded-md border">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead className="w-[100px]">Severity</TableHead>
                            <TableHead>ID</TableHead>
                            <TableHead>Package</TableHead>
                            <TableHead>Installed</TableHead>
                            <TableHead>Fixed In</TableHead>
                            <TableHead>Title</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {filteredVulns.map((v: Vulnerability & { image: string }, i: number) => (
                            <TableRow key={v.vulnerabilityID + i}>
                                <TableCell>
                                    <Badge variant="outline" className={getSeverityColor(v.severity)}>
                                        {v.severity}
                                    </Badge>
                                </TableCell>
                                <TableCell className="font-medium">
                                    {v.primaryLink ? (
                                        <a href={v.primaryLink} target="_blank" rel="noreferrer" className="hover:underline flex items-center">
                                            {v.vulnerabilityID} <ExternalLink className="h-3 w-3 ml-1 opacity-50" />
                                        </a>
                                    ) : v.vulnerabilityID}
                                </TableCell>
                                <TableCell>{v.resource}</TableCell>
                                <TableCell>{v.installedVersion}</TableCell>
                                <TableCell>{v.fixedVersion || "-"}</TableCell>
                                <TableCell className="max-w-md truncate" title={v.title || v.description}>
                                    {v.title || v.description}
                                </TableCell>
                            </TableRow>
                        ))}
                        {filteredVulns.length === 0 && (
                            <TableRow>
                                <TableCell colSpan={6} className="h-24 text-center">
                                    No vulnerabilities found matching your search.
                                </TableCell>
                            </TableRow>
                        )}
                    </TableBody>
                </Table>
            </div>
        </div>
    )
}
