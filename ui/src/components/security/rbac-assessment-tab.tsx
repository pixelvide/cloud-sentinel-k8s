import { useQuery } from "@tanstack/react-query"
import { AlertTriangle, CheckCircle, ExternalLink, ShieldAlert, ShieldCheck } from "lucide-react"
import { useState } from "react"
import { securityApi, ConfigAuditCheck, RbacAssessmentReport, ClusterRbacAssessmentReport } from "@/api/security"
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

interface RbacAssessmentTabProps {
    namespace?: string
    kind: string
    name: string
}

export function RbacAssessmentTab({ namespace, kind, name }: RbacAssessmentTabProps) {
    const [searchTerm, setSearchTerm] = useState("")
    const [showPassed, setShowPassed] = useState(false)
    const isClusterScoped = ["ClusterRole", "ClusterRoleBinding"].includes(kind)

    const { data: status } = useQuery({
        queryKey: ["security", "status"],
        queryFn: () => securityApi.getStatus(),
    })

    const { data: reports, isLoading } = useQuery({
        queryKey: ["security", "rbac-assessment", namespace, kind, name],
        queryFn: () => {
            if (isClusterScoped) {
                return securityApi.getClusterRbacAssessmentReports(name)
            }
            return securityApi.getRbacAssessmentReports(namespace, kind, name)
        },
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
                            RBAC assessment scanning requires the Trivy Operator to be installed in your cluster.
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
        return <div className="p-6 text-center text-muted-foreground">Loading RBAC assessment reports...</div>
    }

    if (!reports?.items || reports.items.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center p-12 text-center border rounded-lg bg-muted/30">
                <div className="bg-secondary p-3 rounded-full mb-4">
                    <ShieldCheck className="h-8 w-8 text-muted-foreground" />
                </div>
                <h3 className="text-lg font-semibold text-foreground">No RBAC Assessment Reports Found</h3>
                <p className="text-sm text-muted-foreground mt-2 max-w-sm">
                    Trivy Operator hasn't scanned this resource yet, or no issues were found.
                </p>
            </div>
        )
    }

    // Aggregate all checks from all reports
    const allChecks = reports.items.flatMap((report: RbacAssessmentReport | ClusterRbacAssessmentReport) =>
        report.report.checks.map((check: ConfigAuditCheck) => ({
            ...check,
            reportName: report.metadata.name
        }))
    )

    // Filter checks
    const filteredChecks = allChecks.filter((check: ConfigAuditCheck & { reportName: string }) => {
        const matchesSearch =
            check.checkID.toLowerCase().includes(searchTerm.toLowerCase()) ||
            check.title?.toLowerCase().includes(searchTerm.toLowerCase()) ||
            check.category?.toLowerCase().includes(searchTerm.toLowerCase())

        if (showPassed) {
            return matchesSearch
        }
        return matchesSearch && !check.success
    })

    // Calculate summary
    const summary = {
        passed: allChecks.filter((c: ConfigAuditCheck) => c.success).length,
        failed: allChecks.filter((c: ConfigAuditCheck) => !c.success).length,
        critical: allChecks.filter((c: ConfigAuditCheck) => !c.success && c.severity === "CRITICAL").length,
        high: allChecks.filter((c: ConfigAuditCheck) => !c.success && c.severity === "HIGH").length,
        medium: allChecks.filter((c: ConfigAuditCheck) => !c.success && c.severity === "MEDIUM").length,
        low: allChecks.filter((c: ConfigAuditCheck) => !c.success && c.severity === "LOW").length,
    }

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
            {/* Summary Cards */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Failed Checks</CardTitle>
                        <AlertTriangle className="h-4 w-4 text-red-600 dark:text-red-400" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-red-600 dark:text-red-400">{summary.failed}</div>
                        <p className="text-xs text-muted-foreground">
                            {summary.critical} critical, {summary.high} high
                        </p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Passed Checks</CardTitle>
                        <CheckCircle className="h-4 w-4 text-green-600 dark:text-green-400" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-green-600 dark:text-green-400">{summary.passed}</div>
                        <p className="text-xs text-muted-foreground">RBAC compliant</p>
                    </CardContent>
                </Card>
            </div>

            {/* Search and Filter */}
            <div className="flex items-center space-x-2">
                <Input
                    placeholder="Search by check ID, title, or category..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="max-w-sm"
                />
                <Button
                    variant={showPassed ? "default" : "outline"}
                    size="sm"
                    onClick={() => setShowPassed(!showPassed)}
                >
                    {showPassed ? "Hide Passed" : "Show Passed"}
                </Button>
            </div>

            {/* Checks Table */}
            <div className="rounded-md border">
                <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead className="w-[80px]">Status</TableHead>
                            <TableHead className="w-[100px]">Severity</TableHead>
                            <TableHead className="w-[150px]">Check ID</TableHead>
                            <TableHead>Title</TableHead>
                            <TableHead className="w-[120px]">Category</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {filteredChecks.map((check: ConfigAuditCheck & { reportName: string }, i: number) => (
                            <TableRow key={check.checkID + i}>
                                <TableCell>
                                    {check.success ? (
                                        <CheckCircle className="h-5 w-5 text-green-600 dark:text-green-400" />
                                    ) : (
                                        <AlertTriangle className="h-5 w-5 text-red-600 dark:text-red-400" />
                                    )}
                                </TableCell>
                                <TableCell>
                                    <Badge variant="outline" className={getSeverityColor(check.severity)}>
                                        {check.severity}
                                    </Badge>
                                </TableCell>
                                <TableCell className="font-mono text-sm">{check.checkID}</TableCell>
                                <TableCell>
                                    <div className="flex flex-col">
                                        <span className="font-medium">{check.title}</span>
                                        {check.messages && check.messages.length > 0 && (
                                            <span className="text-xs text-muted-foreground mt-1 max-w-md truncate" title={check.messages.join(", ")}>
                                                {check.messages[0]}
                                            </span>
                                        )}
                                    </div>
                                </TableCell>
                                <TableCell>
                                    <Badge variant="secondary">{check.category}</Badge>
                                </TableCell>
                            </TableRow>
                        ))}
                        {filteredChecks.length === 0 && (
                            <TableRow>
                                <TableCell colSpan={5} className="h-24 text-center">
                                    No {showPassed ? "" : "failed "}checks found matching your search.
                                </TableCell>
                            </TableRow>
                        )}
                    </TableBody>
                </Table>
            </div>
        </div>
    )
}
