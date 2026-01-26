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
    vulnerableImages: number
    scannedImages: number
    topVulnerableWorkloads: WorkloadSummary[]
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

    getClusterSummary: () => apiClient.get<ClusterSecuritySummary>("/security/summary"),
}
