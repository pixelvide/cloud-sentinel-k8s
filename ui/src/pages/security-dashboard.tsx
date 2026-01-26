import { useQuery } from "@tanstack/react-query"
import { ShieldAlert, ShieldCheck, ShieldQuestion, Settings, Key, ClipboardList, CheckCircle, XCircle } from "lucide-react"
import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from "recharts"
import { securityApi, ClusterComplianceReport } from "@/api/security"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Link } from "react-router-dom"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { useState } from "react"

// Compliance Content Component
function ComplianceContent() {
    const [selectedReport, setSelectedReport] = useState<string>("")

    const { data: reports, isLoading } = useQuery({
        queryKey: ["security", "compliance-reports"],
        queryFn: () => securityApi.getComplianceReports(),
    })

    if (isLoading) {
        return <div className="text-center text-muted-foreground p-8">Loading compliance reports...</div>
    }

    if (!reports?.items || reports.items.length === 0) {
        return (
            <Card>
                <CardContent className="flex flex-col items-center justify-center p-12 text-center">
                    <div className="bg-secondary p-3 rounded-full mb-4">
                        <ClipboardList className="h-8 w-8 text-muted-foreground" />
                    </div>
                    <h3 className="text-lg font-semibold text-foreground">No Compliance Reports</h3>
                    <p className="text-sm text-muted-foreground mt-2 max-w-sm">
                        No ClusterComplianceReports found. Enable compliance scanning in Trivy Operator to see CIS benchmarks and other compliance checks.
                    </p>
                </CardContent>
            </Card>
        )
    }

    // Auto-select first report if none selected
    const currentReportName = selectedReport || reports.items[0]?.metadata.name
    const report = reports.items.find((r: ClusterComplianceReport) => r.metadata.name === currentReportName)

    if (!report) {
        return null
    }

    const passRate = report.status.summary.passCount + report.status.summary.failCount > 0
        ? Math.round((report.status.summary.passCount / (report.status.summary.passCount + report.status.summary.failCount)) * 100)
        : 0

    return (
        <div className="space-y-6">
            {/* Compliance Selector */}
            <div className="flex items-center gap-4">
                <label className="text-sm font-medium">Compliance Benchmark:</label>
                <Select value={currentReportName} onValueChange={setSelectedReport}>
                    <SelectTrigger className="w-[400px]">
                        <SelectValue placeholder="Select a compliance benchmark" />
                    </SelectTrigger>
                    <SelectContent>
                        {reports.items.map((r: ClusterComplianceReport) => {
                            const displayName = r.spec.name || r.metadata.name
                            const version = r.spec.version ? ` (v${r.spec.version})` : ""
                            return (
                                <SelectItem key={r.metadata.name} value={r.metadata.name}>
                                    {displayName}{version}
                                </SelectItem>
                            )
                        })}
                    </SelectContent>
                </Select>
            </div>

            {/* Selected Report */}
            <Card>
                <CardHeader>
                    <div className="flex items-center justify-between">
                        <div>
                            <CardTitle className="flex items-center gap-2">
                                <ClipboardList className="h-5 w-5" />
                                {report.spec.name || report.metadata.name}
                            </CardTitle>
                            <CardDescription>{report.spec.description || `Compliance benchmark: ${report.metadata.name}`}</CardDescription>
                        </div>
                        <Badge variant={passRate >= 80 ? "default" : passRate >= 50 ? "secondary" : "destructive"}>
                            {passRate}% Pass Rate
                        </Badge>
                    </div>
                </CardHeader>
                <CardContent>
                    <div className="grid gap-4 md:grid-cols-3 mb-6">
                        <div className="flex items-center gap-3">
                            <CheckCircle className="h-8 w-8 text-green-600 dark:text-green-400" />
                            <div>
                                <div className="text-2xl font-bold text-green-600 dark:text-green-400">{report.status.summary.passCount}</div>
                                <div className="text-xs text-muted-foreground">Passed</div>
                            </div>
                        </div>
                        <div className="flex items-center gap-3">
                            <XCircle className="h-8 w-8 text-red-600 dark:text-red-400" />
                            <div>
                                <div className="text-2xl font-bold text-red-600 dark:text-red-400">{report.status.summary.failCount}</div>
                                <div className="text-xs text-muted-foreground">Failed</div>
                            </div>
                        </div>
                        <div className="flex items-center gap-3">
                            <ShieldCheck className="h-8 w-8 text-muted-foreground" />
                            <div>
                                <div className="text-2xl font-bold">{report.spec.version}</div>
                                <div className="text-xs text-muted-foreground">Version</div>
                            </div>
                        </div>
                    </div>

                    {report.status.summaryReport?.controlCheck && report.status.summaryReport.controlCheck.length > 0 && (
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>Control ID</TableHead>
                                    <TableHead>Name</TableHead>
                                    <TableHead>Severity</TableHead>
                                    <TableHead>Failed</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {report.status.summaryReport.controlCheck.filter(c => c.totalFail > 0).slice(0, 15).map((control) => (
                                    <TableRow key={control.id}>
                                        <TableCell className="font-mono text-sm">{control.id}</TableCell>
                                        <TableCell>{control.name}</TableCell>
                                        <TableCell>
                                            <Badge variant="outline" className={
                                                control.severity === "CRITICAL" ? "bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-300" :
                                                    control.severity === "HIGH" ? "bg-orange-100 dark:bg-orange-900/30 text-orange-800 dark:text-orange-300" :
                                                        control.severity === "MEDIUM" ? "bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-300" :
                                                            "bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-300"
                                            }>
                                                {control.severity}
                                            </Badge>
                                        </TableCell>
                                        <TableCell>
                                            <Badge variant="destructive">{control.totalFail}</Badge>
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    )}
                </CardContent>
            </Card>
        </div>
    )
}



export function SecurityDashboard() {
    const { data: status } = useQuery({
        queryKey: ["security", "status"],
        queryFn: () => securityApi.getStatus(),
    })

    const { data: summary, isLoading } = useQuery({
        queryKey: ["security", "cluster-summary"],
        queryFn: () => securityApi.getClusterSummary(),
        enabled: !!status?.trivyInstalled,
    })

    const { data: topVulnerable, isLoading: loadingVulnerable } = useQuery({
        queryKey: ["security", "top-vulnerable-workloads"],
        queryFn: () => securityApi.getTopVulnerableWorkloads(),
        enabled: !!status?.trivyInstalled,
    })

    const { data: topMisconfigured, isLoading: loadingMisconfigured } = useQuery({
        queryKey: ["security", "top-misconfigured-workloads"],
        queryFn: () => securityApi.getTopMisconfiguredWorkloads(),
        enabled: !!status?.trivyInstalled,
    })

    if (status && !status.trivyInstalled) {
        return (
            <div className="p-6">
                <h1 className="text-2xl font-bold mb-6">Security Dashboard</h1>
                <Alert variant="default" className="bg-blue-50 dark:bg-blue-950/30 border-blue-200 dark:border-blue-800">
                    <ShieldAlert className="h-5 w-5 text-blue-600 dark:text-blue-400" />
                    <AlertTitle className="text-blue-800 dark:text-blue-300 ml-2">Trivy Operator Not Installed</AlertTitle>
                    <AlertDescription className="text-blue-700 dark:text-blue-400 ml-2 mt-2">
                        <p className="mb-4">
                            To view the security dashboard, you need to install the Trivy Operator in your cluster.
                        </p>
                        <a
                            href="https://aquasecurity.github.io/trivy-operator/latest/getting-started/installation/"
                            target="_blank"
                            rel="noreferrer"
                            className="font-medium underline hover:text-blue-900 dark:hover:text-blue-200"
                        >
                            Installation Guide
                        </a>
                    </AlertDescription>
                </Alert>
            </div>
        )
    }

    if (isLoading) {
        return <div className="p-6">Loading security dashboard...</div>
    }

    if (!summary) {
        return <div className="p-6">No data available.</div>
    }

    const { totalVulnerabilities: vulns, totalConfigAuditIssues: configAudit, totalExposedSecrets: secrets } = summary

    const vulnTotal = vulns.criticalCount + vulns.highCount + vulns.mediumCount + vulns.lowCount || 1
    const configTotal = (configAudit?.criticalCount || 0) + (configAudit?.highCount || 0) + (configAudit?.mediumCount || 0) + (configAudit?.lowCount || 0)
    const secretsTotal = (secrets?.criticalCount || 0) + (secrets?.highCount || 0) + (secrets?.mediumCount || 0) + (secrets?.lowCount || 0)

    return (
        <div className="p-6 space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold tracking-tight">Security Dashboard</h1>
                    <p className="text-muted-foreground">Overview of cluster security posture.</p>
                </div>
            </div>

            {/* Overview Summary Cards */}
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-6">
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Critical CVEs</CardTitle>
                        <ShieldAlert className="h-4 w-4 text-red-600 dark:text-red-400" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-red-600 dark:text-red-400">{vulns.criticalCount}</div>
                        <p className="text-xs text-muted-foreground">Immediate action</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">High CVEs</CardTitle>
                        <ShieldAlert className="h-4 w-4 text-orange-600 dark:text-orange-400" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-orange-600 dark:text-orange-400">{vulns.highCount}</div>
                        <p className="text-xs text-muted-foreground">Fix soon</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Config Issues</CardTitle>
                        <Settings className="h-4 w-4 text-yellow-600 dark:text-yellow-400" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-yellow-600 dark:text-yellow-400">{configTotal}</div>
                        <p className="text-xs text-muted-foreground">Misconfigurations</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Exposed Secrets</CardTitle>
                        <Key className="h-4 w-4 text-purple-600 dark:text-purple-400" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-purple-600 dark:text-purple-400">{secretsTotal}</div>
                        <p className="text-xs text-muted-foreground">In images</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Scanned Images</CardTitle>
                        <ShieldCheck className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{summary.scannedImages}</div>
                        <p className="text-xs text-muted-foreground">Total scanned</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Vulnerable Images</CardTitle>
                        <ShieldQuestion className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{summary.vulnerableImages}</div>
                        <p className="text-xs text-muted-foreground">With issues</p>
                    </CardContent>
                </Card>
            </div>

            {/* Tabbed Content */}
            <Tabs defaultValue="vulnerabilities" className="w-full">
                <TabsList>
                    <TabsTrigger value="vulnerabilities">Vulnerabilities</TabsTrigger>
                    <TabsTrigger value="config">Configuration</TabsTrigger>
                    <TabsTrigger value="compliance">Compliance</TabsTrigger>
                </TabsList>

                <TabsContent value="vulnerabilities" className="space-y-6 mt-4">
                    {/* Vulnerability Distribution */}
                    <Card>
                        <CardHeader>
                            <CardTitle>Vulnerability Distribution</CardTitle>
                            <CardDescription>
                                Breakdown of vulnerabilities by severity.
                            </CardDescription>
                        </CardHeader>
                        <CardContent className="pl-2">
                            <div className="flex flex-col md:flex-row items-center justify-between p-4">
                                <div className="h-[200px] w-full md:w-1/2">
                                    <ResponsiveContainer width="100%" height="100%">
                                        <PieChart>
                                            <Pie
                                                data={[
                                                    { name: "Critical", value: vulns.criticalCount, color: "#dc2626" },
                                                    { name: "High", value: vulns.highCount, color: "#ea580c" },
                                                    { name: "Medium", value: vulns.mediumCount, color: "#eab308" },
                                                    { name: "Low", value: vulns.lowCount, color: "#3b82f6" },
                                                ].filter(item => item.value > 0)}
                                                cx="50%"
                                                cy="50%"
                                                innerRadius={60}
                                                outerRadius={80}
                                                paddingAngle={5}
                                                dataKey="value"
                                            >
                                                {[
                                                    { name: "Critical", value: vulns.criticalCount, color: "#dc2626" },
                                                    { name: "High", value: vulns.highCount, color: "#ea580c" },
                                                    { name: "Medium", value: vulns.mediumCount, color: "#eab308" },
                                                    { name: "Low", value: vulns.lowCount, color: "#3b82f6" },
                                                ].filter(item => item.value > 0).map((entry, index) => (
                                                    <Cell key={`cell-${index}`} fill={entry.color} />
                                                ))}
                                            </Pie>
                                            <Tooltip
                                                contentStyle={{ backgroundColor: 'hsl(var(--card))', borderColor: 'hsl(var(--border))', borderRadius: 'var(--radius)' }}
                                                itemStyle={{ color: 'hsl(var(--foreground))' }}
                                            />
                                        </PieChart>
                                    </ResponsiveContainer>
                                </div>
                                <div className="w-full md:w-1/2 space-y-4">
                                    <div className="flex items-center gap-4">
                                        <div className="w-24 text-sm font-medium">Critical</div>
                                        <div className="flex-1 h-4 bg-secondary rounded-full overflow-hidden">
                                            <div className="h-full bg-red-500 dark:bg-red-600" style={{ width: `${(vulns.criticalCount / vulnTotal) * 100}%` }} />
                                        </div>
                                        <div className="w-12 text-sm text-right">{vulns.criticalCount}</div>
                                    </div>
                                    <div className="flex items-center gap-4">
                                        <div className="w-24 text-sm font-medium">High</div>
                                        <div className="flex-1 h-4 bg-secondary rounded-full overflow-hidden">
                                            <div className="h-full bg-orange-500 dark:bg-orange-600" style={{ width: `${(vulns.highCount / vulnTotal) * 100}%` }} />
                                        </div>
                                        <div className="w-12 text-sm text-right">{vulns.highCount}</div>
                                    </div>
                                    <div className="flex items-center gap-4">
                                        <div className="w-24 text-sm font-medium">Medium</div>
                                        <div className="flex-1 h-4 bg-secondary rounded-full overflow-hidden">
                                            <div className="h-full bg-yellow-500 dark:bg-yellow-600" style={{ width: `${(vulns.mediumCount / vulnTotal) * 100}%` }} />
                                        </div>
                                        <div className="w-12 text-sm text-right">{vulns.mediumCount}</div>
                                    </div>
                                    <div className="flex items-center gap-4">
                                        <div className="w-24 text-sm font-medium">Low</div>
                                        <div className="flex-1 h-4 bg-secondary rounded-full overflow-hidden">
                                            <div className="h-full bg-blue-500 dark:bg-blue-600" style={{ width: `${(vulns.lowCount / vulnTotal) * 100}%` }} />
                                        </div>
                                        <div className="w-12 text-sm text-right">{vulns.lowCount}</div>
                                    </div>
                                </div>
                            </div>
                        </CardContent>
                    </Card>

                    {/* Top Vulnerable Workloads */}
                    <Card className="h-full">
                        <CardHeader>
                            <CardTitle>Top Vulnerable Workloads</CardTitle>
                            <CardDescription>Workloads with the highest number of critical and high vulnerabilities.</CardDescription>
                        </CardHeader>
                        <CardContent>
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>Workload</TableHead>
                                        <TableHead>Critical</TableHead>
                                        <TableHead>High</TableHead>
                                        <TableHead>Medium</TableHead>
                                        <TableHead>Low</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {loadingVulnerable ? (
                                        <TableRow>
                                            <TableCell colSpan={5} className="text-center py-8 text-muted-foreground">
                                                Loading top vulnerable workloads...
                                            </TableCell>
                                        </TableRow>
                                    ) : (topVulnerable?.items || []).map((workload) => (
                                        <TableRow key={`${workload.namespace}-${workload.kind}-${workload.name}`}>
                                            <TableCell>
                                                <div className="flex flex-col">
                                                    <span className="font-medium text-sm">
                                                        <Link
                                                            to={`../${workload.kind.toLowerCase()}s/${workload.namespace}/${workload.name}`}
                                                            className="hover:underline text-blue-600 dark:text-blue-400"
                                                        >
                                                            {workload.name}
                                                        </Link>
                                                    </span>
                                                    <span className="text-xs text-muted-foreground">{workload.kind} • {workload.namespace}</span>
                                                </div>
                                            </TableCell>
                                            <TableCell>
                                                {workload.vulnerabilities.criticalCount > 0 && (
                                                    <Badge variant="destructive" className="bg-red-600 hover:bg-red-700">
                                                        {workload.vulnerabilities.criticalCount}
                                                    </Badge>
                                                )}
                                            </TableCell>
                                            <TableCell>
                                                {workload.vulnerabilities.highCount > 0 && (
                                                    <Badge variant="secondary" className="bg-orange-500/15 text-orange-700 dark:text-orange-400 border-orange-200 dark:border-orange-800">
                                                        {workload.vulnerabilities.highCount}
                                                    </Badge>
                                                )}
                                            </TableCell>
                                            <TableCell>
                                                {workload.vulnerabilities.mediumCount > 0 && (
                                                    <span className="text-yellow-600 dark:text-yellow-400 font-medium">
                                                        {workload.vulnerabilities.mediumCount}
                                                    </span>
                                                )}
                                            </TableCell>
                                            <TableCell>
                                                {workload.vulnerabilities.lowCount > 0 && (
                                                    <span className="text-blue-600 dark:text-blue-400">
                                                        {workload.vulnerabilities.lowCount}
                                                    </span>
                                                )}
                                            </TableCell>
                                        </TableRow>
                                    ))}
                                    {!loadingVulnerable && (!topVulnerable?.items || topVulnerable.items.length === 0) && (
                                        <TableRow>
                                            <TableCell colSpan={5} className="text-center py-8 text-muted-foreground">
                                                No vulnerable workloads found.
                                            </TableCell>
                                        </TableRow>
                                    )}
                                </TableBody>
                            </Table>
                        </CardContent>
                    </Card>
                </TabsContent>

                <TabsContent value="config" className="space-y-6 mt-4">
                    {/* Config Audit Distribution */}
                    {configTotal > 0 && (
                        <Card>
                            <CardHeader>
                                <CardTitle>Configuration Issues Distribution</CardTitle>
                                <CardDescription>
                                    Breakdown of Kubernetes misconfigurations by severity.
                                </CardDescription>
                            </CardHeader>
                            <CardContent className="pl-2">
                                <div className="w-full space-y-4 p-4">
                                    <div className="flex items-center gap-4">
                                        <div className="w-24 text-sm font-medium">Critical</div>
                                        <div className="flex-1 h-4 bg-secondary rounded-full overflow-hidden">
                                            <div className="h-full bg-red-500 dark:bg-red-600" style={{ width: `${((configAudit?.criticalCount || 0) / (configTotal || 1)) * 100}%` }} />
                                        </div>
                                        <div className="w-12 text-sm text-right">{configAudit?.criticalCount || 0}</div>
                                    </div>
                                    <div className="flex items-center gap-4">
                                        <div className="w-24 text-sm font-medium">High</div>
                                        <div className="flex-1 h-4 bg-secondary rounded-full overflow-hidden">
                                            <div className="h-full bg-orange-500 dark:bg-orange-600" style={{ width: `${((configAudit?.highCount || 0) / (configTotal || 1)) * 100}%` }} />
                                        </div>
                                        <div className="w-12 text-sm text-right">{configAudit?.highCount || 0}</div>
                                    </div>
                                    <div className="flex items-center gap-4">
                                        <div className="w-24 text-sm font-medium">Medium</div>
                                        <div className="flex-1 h-4 bg-secondary rounded-full overflow-hidden">
                                            <div className="h-full bg-yellow-500 dark:bg-yellow-600" style={{ width: `${((configAudit?.mediumCount || 0) / (configTotal || 1)) * 100}%` }} />
                                        </div>
                                        <div className="w-12 text-sm text-right">{configAudit?.mediumCount || 0}</div>
                                    </div>
                                    <div className="flex items-center gap-4">
                                        <div className="w-24 text-sm font-medium">Low</div>
                                        <div className="flex-1 h-4 bg-secondary rounded-full overflow-hidden">
                                            <div className="h-full bg-blue-500 dark:bg-blue-600" style={{ width: `${((configAudit?.lowCount || 0) / (configTotal || 1)) * 100}%` }} />
                                        </div>
                                        <div className="w-12 text-sm text-right">{configAudit?.lowCount || 0}</div>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    )}

                    {/* Top Misconfigured Workloads */}
                    <Card className="h-full">
                        <CardHeader>
                            <CardTitle>Top Misconfigured Workloads</CardTitle>
                            <CardDescription>Workloads with the most configuration issues.</CardDescription>
                        </CardHeader>
                        <CardContent>
                            <Table>
                                <TableHeader>
                                    <TableRow>
                                        <TableHead>Workload</TableHead>
                                        <TableHead>Critical</TableHead>
                                        <TableHead>High</TableHead>
                                        <TableHead>Medium</TableHead>
                                        <TableHead>Low</TableHead>
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {loadingMisconfigured ? (
                                        <TableRow>
                                            <TableCell colSpan={5} className="text-center py-8 text-muted-foreground">
                                                Loading top misconfigured workloads...
                                            </TableCell>
                                        </TableRow>
                                    ) : (topMisconfigured?.items || []).map((workload) => (
                                        <TableRow key={`${workload.namespace}-${workload.kind}-${workload.name}`}>
                                            <TableCell>
                                                <div className="flex flex-col">
                                                    <span className="font-medium text-sm">
                                                        <Link
                                                            to={`../${workload.kind.toLowerCase()}s/${workload.namespace}/${workload.name}`}
                                                            className="hover:underline text-blue-600 dark:text-blue-400"
                                                        >
                                                            {workload.name}
                                                        </Link>
                                                    </span>
                                                    <span className="text-xs text-muted-foreground">{workload.kind} • {workload.namespace}</span>
                                                </div>
                                            </TableCell>
                                            <TableCell>
                                                {workload.vulnerabilities.criticalCount > 0 && (
                                                    <Badge variant="destructive" className="bg-red-600 hover:bg-red-700">
                                                        {workload.vulnerabilities.criticalCount}
                                                    </Badge>
                                                )}
                                            </TableCell>
                                            <TableCell>
                                                {workload.vulnerabilities.highCount > 0 && (
                                                    <Badge variant="secondary" className="bg-orange-500/15 text-orange-700 dark:text-orange-400 border-orange-200 dark:border-orange-800">
                                                        {workload.vulnerabilities.highCount}
                                                    </Badge>
                                                )}
                                            </TableCell>
                                            <TableCell>
                                                {workload.vulnerabilities.mediumCount > 0 && (
                                                    <span className="text-yellow-600 dark:text-yellow-400 font-medium">
                                                        {workload.vulnerabilities.mediumCount}
                                                    </span>
                                                )}
                                            </TableCell>
                                            <TableCell>
                                                {workload.vulnerabilities.lowCount > 0 && (
                                                    <span className="text-blue-600 dark:text-blue-400">
                                                        {workload.vulnerabilities.lowCount}
                                                    </span>
                                                )}
                                            </TableCell>
                                        </TableRow>
                                    ))}
                                    {!loadingMisconfigured && (!topMisconfigured?.items || topMisconfigured.items.length === 0) && (
                                        <TableRow>
                                            <TableCell colSpan={5} className="text-center py-8 text-muted-foreground">
                                                No misconfigured workloads found.
                                            </TableCell>
                                        </TableRow>
                                    )}
                                </TableBody>
                            </Table>
                        </CardContent>
                    </Card>
                </TabsContent>

                <TabsContent value="compliance" className="space-y-6 mt-4">
                    <ComplianceContent />
                </TabsContent>
            </Tabs>
        </div>
    )
}
