import { useQuery } from "@tanstack/react-query"
import { ExternalLink, ShieldAlert, Bug, Settings, Key, Server, ShieldCheck } from "lucide-react"
import { securityApi } from "@/api/security"
import { Button } from "@/components/ui/button"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { ConfigAuditTab } from "./config-audit-tab"
import { ExposedSecretsTab } from "./exposed-secrets-tab"
import { InfraAssessmentTab } from "./infra-assessment-tab"
import { VulnerabilitiesTab } from "./vulnerabilities-tab"
import { RbacAssessmentTab } from "./rbac-assessment-tab"

type SecurityTabType = 'vulnerabilities' | 'config-audit' | 'infra-assessment' | 'secrets' | 'rbac-assessment'

interface SecurityTabProps {
    namespace?: string
    kind: string
    name: string
    visibleTabs?: SecurityTabType[]
}

export function SecurityTab({ namespace, kind, name, visibleTabs }: SecurityTabProps) {
    const { data: status } = useQuery({
        queryKey: ["security", "status"],
        queryFn: () => securityApi.getStatus(),
    })

    const checkRbacVisibility = () => {
        const rbacKinds = ['ClusterRole', 'Role']
        return rbacKinds.includes(kind)
    }

    const defaultTabs: SecurityTabType[] = ['vulnerabilities', 'config-audit', 'infra-assessment', 'secrets']
    if (checkRbacVisibility()) {
        defaultTabs.push('rbac-assessment')
    }

    const tabsToShow: SecurityTabType[] = visibleTabs || defaultTabs
    const defaultTab = tabsToShow.length > 0 ? tabsToShow[0] : 'vulnerabilities'

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

    if (tabsToShow.length === 0) {
        return null
    }

    return (
        <div className="space-y-6">
            <Tabs defaultValue={defaultTab} className="w-full">
                <TabsList className="grid w-full grid-cols-5 lg:w-[680px]">
                    {tabsToShow.includes('vulnerabilities') && (
                        <TabsTrigger value="vulnerabilities" className="flex items-center gap-2">
                            <span className="hidden sm:inline">Vulnerabilities</span>
                            <span className="sm:hidden">CVEs</span>
                        </TabsTrigger>
                    )}
                    {tabsToShow.includes('config-audit') && (
                        <TabsTrigger value="config-audit" className="flex items-center gap-2">
                            <span className="hidden sm:inline">Config Audit</span>
                            <span className="sm:hidden">Config</span>
                        </TabsTrigger>
                    )}
                    {tabsToShow.includes('infra-assessment') && (
                        <TabsTrigger value="infra-assessment" className="flex items-center gap-2">
                            <span className="hidden sm:inline">Infrastructure</span>
                            <span className="sm:hidden">Infra</span>
                        </TabsTrigger>
                    )}
                    {tabsToShow.includes('rbac-assessment') && (
                        <TabsTrigger value="rbac-assessment" className="flex items-center gap-2">
                            <span className="hidden sm:inline">RBAC Audit</span>
                            <span className="sm:hidden">RBAC</span>
                        </TabsTrigger>
                    )}
                    {tabsToShow.includes('secrets') && (
                        <TabsTrigger value="secrets" className="flex items-center gap-2">
                            <span className="hidden sm:inline">Secrets</span>
                            <span className="sm:hidden">Secrets</span>
                        </TabsTrigger>
                    )}
                </TabsList>

                {tabsToShow.includes('vulnerabilities') && (
                    <TabsContent value="vulnerabilities" className="mt-6">
                        <VulnerabilitiesTab namespace={namespace} kind={kind} name={name} />
                    </TabsContent>
                )}

                {tabsToShow.includes('config-audit') && (
                    <TabsContent value="config-audit" className="mt-6">
                        <ConfigAuditTab namespace={namespace} kind={kind} name={name} />
                    </TabsContent>
                )}

                {tabsToShow.includes('infra-assessment') && (
                    <TabsContent value="infra-assessment" className="mt-6">
                        <InfraAssessmentTab namespace={namespace} kind={kind} name={name} />
                    </TabsContent>
                )}

                {tabsToShow.includes('rbac-assessment') && (
                    <TabsContent value="rbac-assessment" className="mt-6">
                        <RbacAssessmentTab namespace={namespace} kind={kind} name={name} />
                    </TabsContent>
                )}

                {tabsToShow.includes('secrets') && (
                    <TabsContent value="secrets" className="mt-6">
                        <ExposedSecretsTab namespace={namespace} kind={kind} name={name} />
                    </TabsContent>
                )}
            </Tabs>
        </div>
    )
}
