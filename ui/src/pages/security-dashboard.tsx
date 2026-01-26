import { useQuery } from "@tanstack/react-query"
import { ShieldAlert, ShieldCheck, ShieldQuestion } from "lucide-react"
import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from "recharts"
import { securityApi } from "@/api/security"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Link } from "react-router-dom"

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

    const { totalVulnerabilities: vulns } = summary

    const total = vulns.criticalCount + vulns.highCount + vulns.mediumCount + vulns.lowCount || 1

    return (
        <div className="p-6 space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold tracking-tight">Security Dashboard</h1>
                    <p className="text-muted-foreground">Overview of cluster security posture.</p>
                </div>
            </div>

            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Critical Vulnerabilities</CardTitle>
                        <ShieldAlert className="h-4 w-4 text-red-600 dark:text-red-400" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-red-600 dark:text-red-400">{vulns.criticalCount}</div>
                        <p className="text-xs text-muted-foreground">Requires immediate attention</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">High Vulnerabilities</CardTitle>
                        <ShieldAlert className="h-4 w-4 text-orange-600 dark:text-orange-400" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold text-orange-600 dark:text-orange-400">{vulns.highCount}</div>
                        <p className="text-xs text-muted-foreground">Should be fixed soon</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Scanned Images</CardTitle>
                        <ShieldCheck className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{summary.scannedImages}</div>
                        <p className="text-xs text-muted-foreground">Total images scanned</p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">Vulnerable Images</CardTitle>
                        <ShieldQuestion className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{summary.vulnerableImages}</div>
                        <p className="text-xs text-muted-foreground">Images with at least one issue</p>
                    </CardContent>
                </Card>
            </div>

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
                                    <div className="h-full bg-red-500 dark:bg-red-600" style={{ width: `${(vulns.criticalCount / total) * 100}%` }} />
                                </div>
                                <div className="w-12 text-sm text-right">{vulns.criticalCount}</div>
                            </div>
                            <div className="flex items-center gap-4">
                                <div className="w-24 text-sm font-medium">High</div>
                                <div className="flex-1 h-4 bg-secondary rounded-full overflow-hidden">
                                    <div className="h-full bg-orange-500 dark:bg-orange-600" style={{ width: `${(vulns.highCount / total) * 100}%` }} />
                                </div>
                                <div className="w-12 text-sm text-right">{vulns.highCount}</div>
                            </div>
                            <div className="flex items-center gap-4">
                                <div className="w-24 text-sm font-medium">Medium</div>
                                <div className="flex-1 h-4 bg-secondary rounded-full overflow-hidden">
                                    <div className="h-full bg-yellow-500 dark:bg-yellow-600" style={{ width: `${(vulns.mediumCount / total) * 100}%` }} />
                                </div>
                                <div className="w-12 text-sm text-right">{vulns.mediumCount}</div>
                            </div>
                            <div className="flex items-center gap-4">
                                <div className="w-24 text-sm font-medium">Low</div>
                                <div className="flex-1 h-4 bg-secondary rounded-full overflow-hidden">
                                    <div className="h-full bg-blue-500 dark:bg-blue-600" style={{ width: `${(vulns.lowCount / total) * 100}%` }} />
                                </div>
                                <div className="w-12 text-sm text-right">{vulns.lowCount}</div>
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>

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
                            {(summary.topVulnerableWorkloads || []).map((workload) => (
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
                            {(!summary.topVulnerableWorkloads || summary.topVulnerableWorkloads.length === 0) && (
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
        </div>
    )
}
