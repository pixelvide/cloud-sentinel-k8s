import { apiClient } from "@/lib/api-client"

export interface Vulnerability {
    vulnerabilityID: string
    resource: string
    installedVersion: string
    fixedVersion: string
    severity: "CRITICAL" | "HIGH" | "MEDIUM" | "LOW" | "UNKNOWN"
    title: string
    description: string
    primaryLink: string
    links: string[]
    score: number
}

export interface VulnerabilitySummary {
    criticalCount: number
    highCount: number
    mediumCount: number
    lowCount: number
    unknownCount: number
}

// Generic summary for check-based reports (ConfigAudit, ExposedSecrets, RBAC)
export interface CheckSummary {
    criticalCount: number
    highCount: number
    mediumCount: number
    lowCount: number
}

export interface Artifact {
    repository: string
    tag: string
}

export interface Scanner {
    name: string
    vendor: string
    version: string
}

export interface VulnerabilityReportData {
    artifact: Artifact
    scanner: Scanner
    summary: VulnerabilitySummary
    vulnerabilities: Vulnerability[]
}

export interface VulnerabilityReport {
    metadata: {
        name: string
        namespace: string
        creationTimestamp: string
    }
    report: VulnerabilityReportData
}

// ConfigAuditReport types
export interface ConfigAuditCheck {
    checkID: string
    title: string
    description: string
    severity: "CRITICAL" | "HIGH" | "MEDIUM" | "LOW"
    category: string
    success: boolean
    messages: string[]
}

export interface ConfigAuditReportData {
    scanner: Scanner
    summary: CheckSummary
    checks: ConfigAuditCheck[]
}

export interface ConfigAuditReport {
    metadata: {
        name: string
        namespace: string
        creationTimestamp: string
    }
    report: ConfigAuditReportData
}

// ExposedSecretReport types
export interface ExposedSecret {
    target: string
    ruleID: string
    title: string
    category: string
    severity: string
    match: string
}

export interface ExposedSecretReportData {
    scanner: Scanner
    artifact: Artifact
    summary: CheckSummary
    secrets: ExposedSecret[]
}

export interface ExposedSecretReport {
    metadata: {
        name: string
        namespace: string
        creationTimestamp: string
    }
    report: ExposedSecretReportData
}

export interface SecurityStatus {
    trivyInstalled: boolean
}

export interface WorkloadSummary {
    namespace: string
    kind: string
    name: string
    vulnerabilities: VulnerabilitySummary
}

export interface ClusterSecuritySummary {
    totalVulnerabilities: VulnerabilitySummary
    totalConfigAuditIssues: CheckSummary
    totalRbacAssessmentIssues: CheckSummary
    totalExposedSecrets: CheckSummary
    vulnerableImages: number
    scannedImages: number
    topVulnerableWorkloads?: WorkloadSummary[]
    topMisconfigured?: WorkloadSummary[]
}

// ClusterComplianceReport types
export interface ComplianceSummary {
    failCount: number
    passCount: number
}

export interface ControlCheckSummary {
    id: string
    name: string
    severity: string
    totalFail: number
}

export interface ClusterComplianceReportSpec {
    name: string
    description: string
    version: string
}

export interface ClusterComplianceReportStatus {
    summary: ComplianceSummary
    summaryReport?: {
        controlCheck: ControlCheckSummary[]
    }
    updateTimestamp?: string
}

export interface ClusterComplianceReport {
    metadata: {
        name: string
        creationTimestamp: string
    }
    spec: ClusterComplianceReportSpec
    status: ClusterComplianceReportStatus
}


// InfraAssessmentReport types
export interface InfraAssessmentReport {
    metadata: {
        name: string
        namespace: string
        creationTimestamp: string
    }
    report: ConfigAuditReportData // Reusing ConfigAuditReportData as structure is likely identical
}

export interface ClusterInfraAssessmentReport {
    metadata: {
        name: string
        creationTimestamp: string
    }
    report: ConfigAuditReportData // Reusing ConfigAuditReportData
}

// RbacAssessmentReport types
export interface RbacAssessmentReport {
    metadata: {
        name: string
        namespace: string
        creationTimestamp: string
    }
    report: ConfigAuditReportData // Reusing ConfigAuditReportData
}

export interface ClusterRbacAssessmentReport {
    metadata: {
        name: string
        creationTimestamp: string
    }
    report: ConfigAuditReportData // Reusing ConfigAuditReportData
}

export const securityApi = {
    getStatus: () => apiClient.get<SecurityStatus>("/security/status"),

    getReports: (namespace: string | undefined, workloadKind: string, workloadName: string) => {
        const params = new URLSearchParams({
            workloadKind,
            workloadName
        })
        if (namespace) params.append("namespace", namespace)
        return apiClient.get<{ items: VulnerabilityReport[] }>(`/security/reports?${params.toString()}`)
    },

    getConfigAuditReports: (namespace: string | undefined, workloadKind: string, workloadName: string) => {
        const params = new URLSearchParams({
            workloadKind,
            workloadName
        })
        if (namespace) params.append("namespace", namespace)
        return apiClient.get<{ items: ConfigAuditReport[] }>(`/security/config-audit/reports?${params.toString()}`)
    },

    getInfraAssessmentReports: (namespace: string | undefined, workloadKind: string, workloadName: string) => {
        const params = new URLSearchParams({
            workloadKind,
            workloadName
        })
        if (namespace) params.append("namespace", namespace)
        return apiClient.get<{ items: InfraAssessmentReport[] }>(`/security/infra-assessment/reports?${params.toString()}`)
    },

    getClusterInfraAssessmentReports: (workloadName: string) => {
        // For cluster scoped resources, we might filter by name if needed, or just fetch all and filter client side
        // Assuming the backend supports filtering by name for ClusterInfraAssessmentReport if relevant
        // Or we just rely on labels. For now, let's assume a similar pattern or just fetching all if no filter.
        // But since this is specific to a Node, we likely need to find the report for that Node.
        // The standard Trivy Operator creates a report with name like "node-<node-name>" or similar.
        // Let's assume we can query by label or name. 
        // Based on existing pattern, we'll pass the workload name (Node name).
        const params = new URLSearchParams({
            workloadName
        })
        return apiClient.get<{ items: ClusterInfraAssessmentReport[] }>(`/security/cluster-infra-assessment/reports?${params.toString()}`)
    },

    getRbacAssessmentReports: (namespace: string | undefined, workloadKind: string, workloadName: string) => {
        const params = new URLSearchParams({
            workloadKind,
            workloadName
        })
        if (namespace) params.append("namespace", namespace)
        return apiClient.get<{ items: RbacAssessmentReport[] }>(`/security/rbac-assessment/reports?${params.toString()}`)
    },

    getClusterRbacAssessmentReports: (workloadName: string) => {
        const params = new URLSearchParams({
            workloadName
        })
        return apiClient.get<{ items: ClusterRbacAssessmentReport[] }>(`/security/cluster-rbac-assessment/reports?${params.toString()}`)
    },

    getExposedSecretReports: (namespace: string | undefined, workloadKind: string, workloadName: string) => {
        const params = new URLSearchParams({
            workloadKind,
            workloadName
        })
        if (namespace) params.append("namespace", namespace)
        return apiClient.get<{ items: ExposedSecretReport[] }>(`/security/secrets/reports?${params.toString()}`)
    },

    getComplianceReports: () => apiClient.get<{ items: ClusterComplianceReport[] }>("/security/compliance/reports"),

    getClusterSummary: () => apiClient.get<ClusterSecuritySummary>("/security/summary"),

    getTopVulnerableWorkloads: () => apiClient.get<{ items: WorkloadSummary[] }>("/security/reports/top-vulnerable"),

    getTopMisconfiguredWorkloads: () => apiClient.get<{ items: WorkloadSummary[] }>("/security/reports/top-misconfigured"),

    getTopRbacRiskyWorkloads: () => apiClient.get<{ items: WorkloadSummary[] }>("/security/reports/top-rbac-risky"),
}
