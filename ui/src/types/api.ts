// API types for Custom Resources

import {
  MutatingWebhookConfiguration,
  MutatingWebhookConfigurationList,
  ValidatingWebhookConfiguration,
  ValidatingWebhookConfigurationList,
} from 'kubernetes-types/admissionregistration/v1'
import {
  CustomResourceDefinition,
  CustomResourceDefinitionList,
} from 'kubernetes-types/apiextensions/v1'
import {
  DaemonSet,
  DaemonSetList,
  Deployment,
  DeploymentList,
  ReplicaSet,
  ReplicaSetList,
  StatefulSet,
  StatefulSetList,
} from 'kubernetes-types/apps/v1'
import {
  HorizontalPodAutoscaler,
  HorizontalPodAutoscalerList,
} from 'kubernetes-types/autoscaling/v2'
import { CronJob, CronJobList, Job, JobList } from 'kubernetes-types/batch/v1'
import { Lease, LeaseList } from 'kubernetes-types/coordination/v1'
import {
  ConfigMap,
  ConfigMapList,
  Endpoints,
  EndpointsList,
  Event,
  EventList,
  LimitRange,
  LimitRangeList,
  Namespace,
  NamespaceList,
  Node,
  PersistentVolume,
  PersistentVolumeClaim,
  PersistentVolumeClaimList,
  PersistentVolumeList,
  Pod,
  ResourceQuota,
  ResourceQuotaList,
  Secret,
  SecretList,
  Service,
  ServiceAccount,
  ServiceAccountList,
  ServiceList,
} from 'kubernetes-types/core/v1'
import {
  Ingress,
  IngressClass,
  IngressClassList,
  IngressList,
  NetworkPolicy,
  NetworkPolicyList,
} from 'kubernetes-types/networking/v1'
import { RuntimeClass, RuntimeClassList } from 'kubernetes-types/node/v1'
import {
  PodDisruptionBudget,
  PodDisruptionBudgetList,
} from 'kubernetes-types/policy/v1'
import {
  ClusterRole,
  ClusterRoleBinding,
  ClusterRoleBindingList,
  ClusterRoleList,
  Role as RawRole,
  RoleBinding,
  RoleBindingList,
  RoleList,
} from 'kubernetes-types/rbac/v1'
import {
  PriorityClass,
  PriorityClassList,
} from 'kubernetes-types/scheduling/v1'
import { StorageClass, StorageClassList } from 'kubernetes-types/storage/v1'

export interface CustomResource {
  apiVersion: string
  kind: string
  metadata: {
    name: string
    namespace?: string
    creationTimestamp: string
    uid?: string
    resourceVersion?: string
    labels?: Record<string, string>
    annotations?: Record<string, string>
  }
  spec?: Record<string, unknown>
  status?: Record<string, unknown>
}

export interface CustomResourceList {
  apiVersion: string
  kind: string
  items: CustomResource[]
  metadata?: {
    continue?: string
    remainingItemCount?: number
  }
}

export interface DeploymentRelatedResource {
  events: Event[]
  pods: Pod[]
  services: Service[]
}

// Resource type definitions
export type ResourceType =
  | 'pods'
  | 'deployments'
  | 'statefulsets'
  | 'daemonsets'
  | 'jobs'
  | 'cronjobs'
  | 'services'
  | 'endpoints'
  | 'configmaps'
  | 'secrets'
  | 'ingresses'
  | 'ingressclasses'
  | 'networkpolicies'
  | 'namespaces'
  | 'crds'
  | 'crs'
  | 'nodes'
  | 'events'
  | 'persistentvolumes'
  | 'persistentvolumeclaims'
  | 'storageclasses'
  | 'podmetrics'
  | 'replicasets'
  | 'serviceaccounts'
  | 'roles'
  | 'rolebindings'
  | 'clusterroles'
  | 'clusterrolebindings'
  | 'horizontalpodautoscalers'
  | 'poddisruptionbudgets'
  | 'priorityclasses'
  | 'runtimeclasses'
  | 'leases'
  | 'helmreleases'
  | 'mutatingwebhookconfigurations'
  | 'validatingwebhookconfigurations'
  | 'resourcequotas'
  | 'limitranges'

export const clusterScopeResources: ResourceType[] = [
  'crds',
  'namespaces',
  'persistentvolumes',
  'nodes',
  'storageclasses',
  'clusterroles',
  'clusterrolebindings',
  'priorityclasses',
  'runtimeclasses',
  'mutatingwebhookconfigurations',
  'validatingwebhookconfigurations',
  'ingressclasses',
]

type listMetadataType = {
  continue?: string
  remainingItemCount?: number
}

// Define resource type mappings
export interface ResourcesTypeMap {
  pods: {
    items: PodWithMetrics[]
    metadata?: listMetadataType
  }
  deployments: DeploymentList
  statefulsets: StatefulSetList
  daemonsets: DaemonSetList
  jobs: JobList
  cronjobs: CronJobList
  services: ServiceList
  endpoints: EndpointsList
  configmaps: ConfigMapList
  secrets: SecretList
  persistentvolumeclaims: PersistentVolumeClaimList
  ingresses: IngressList
  ingressclasses: IngressClassList
  networkpolicies: NetworkPolicyList
  namespaces: NamespaceList
  crds: CustomResourceDefinitionList
  crs: {
    items: CustomResource[]
    metadata?: listMetadataType
  }
  nodes: {
    items: NodeWithMetrics[]
    metadata?: listMetadataType
  }
  events: EventList
  persistentvolumes: PersistentVolumeList
  storageclasses: StorageClassList
  podmetrics: {
    items: PodMetrics[]
    metadata?: listMetadataType
  }
  replicasets: ReplicaSetList
  serviceaccounts: ServiceAccountList
  roles: RoleList
  rolebindings: RoleBindingList
  clusterroles: ClusterRoleList
  clusterrolebindings: ClusterRoleBindingList
  horizontalpodautoscalers: HorizontalPodAutoscalerList
  poddisruptionbudgets: PodDisruptionBudgetList
  priorityclasses: PriorityClassList
  runtimeclasses: RuntimeClassList
  leases: LeaseList
  resourcequotas: ResourceQuotaList
  limitranges: LimitRangeList
  helmreleases: {
    items: HelmRelease[]
  }
  mutatingwebhookconfigurations: MutatingWebhookConfigurationList
  validatingwebhookconfigurations: ValidatingWebhookConfigurationList
}

export interface PodMetrics {
  metadata: {
    name: string
    namespace: string
    labels?: Record<string, string>
    annotations?: Record<string, string>
    creationTimestamp?: string
    uid?: string
    resourceVersion?: string
  }
  containers: {
    name: string // container name
    usage: {
      cpu: string // 214572390n
      memory: string // 2956516Ki
    }
  }[]
}

export type MetricsData = {
  cpuUsage?: number
  memoryUsage?: number
  cpuLimit?: number
  memoryLimit?: number
  cpuRequest?: number
  memoryRequest?: number
  pods?: number
  podsLimit?: number
}

export type PodWithMetrics = Pod & {
  metrics?: MetricsData
}

export type NodeWithMetrics = Node & {
  metrics?: MetricsData
}

export interface ResourceTypeMap {
  pods: PodWithMetrics
  deployments: Deployment
  statefulsets: StatefulSet
  daemonsets: DaemonSet
  jobs: Job
  cronjobs: CronJob
  services: Service
  endpoints: Endpoints
  configmaps: ConfigMap
  secrets: Secret
  persistentvolumeclaims: PersistentVolumeClaim
  ingresses: Ingress
  ingressclasses: IngressClass
  networkpolicies: NetworkPolicy
  namespaces: Namespace
  crds: CustomResourceDefinition
  crs: CustomResource
  nodes: NodeWithMetrics
  events: Event
  persistentvolumes: PersistentVolume
  storageclasses: StorageClass
  replicasets: ReplicaSet
  podmetrics: PodMetrics
  serviceaccounts: ServiceAccount
  roles: RawRole
  rolebindings: RoleBinding
  clusterroles: ClusterRole
  clusterrolebindings: ClusterRoleBinding
  horizontalpodautoscalers: HorizontalPodAutoscaler
  poddisruptionbudgets: PodDisruptionBudget
  priorityclasses: PriorityClass
  runtimeclasses: RuntimeClass
  leases: Lease
  resourcequotas: ResourceQuota
  limitranges: LimitRange
  helmreleases: HelmRelease
  mutatingwebhookconfigurations: MutatingWebhookConfiguration
  validatingwebhookconfigurations: ValidatingWebhookConfiguration
}

export interface HelmRelease {
  name: string
  namespace: string
  revision: number
  status: string
  chart: string
  app_version: string
  updated: string
}

export interface RecentEvent {
  type: string
  reason: string
  message: string
  involvedObjectKind: string
  involvedObjectName: string
  namespace?: string
  timestamp: string
}

export interface UsageDataPoint {
  timestamp: string
  value: number
}

export interface ResourceUsageHistory {
  cpu: UsageDataPoint[]
  memory: UsageDataPoint[]
  networkIn: UsageDataPoint[]
  networkOut: UsageDataPoint[]
  diskRead: UsageDataPoint[]
  diskWrite: UsageDataPoint[]
}

// Pod monitoring types
export interface PodMetrics {
  cpu: UsageDataPoint[]
  memory: UsageDataPoint[]
  networkIn?: UsageDataPoint[]
  networkOut?: UsageDataPoint[]
  diskRead?: UsageDataPoint[]
  diskWrite?: UsageDataPoint[]
  fallback?: boolean
}

export interface OverviewData {
  totalNodes: number
  readyNodes: number
  totalPods: number
  runningPods: number
  totalNamespaces: number
  totalServices: number
  prometheusEnabled: boolean
  resource: {
    cpu: {
      allocatable: number
      requested: number
      limited: number
    }
    memory: {
      allocatable: number
      requested: number
      limited: number
    }
  }
}

// Pagination types
export interface PaginationInfo {
  hasNextPage: boolean
  nextContinueToken?: string
  remainingItems?: number
}

export interface PaginationOptions {
  limit?: number
  continueToken?: string
}

// Pod current metrics types
export interface PodCurrentMetrics {
  podName: string
  namespace: string
  cpu: number // CPU cores
  memory: number // Memory in MB
}

export interface ImageTagInfo {
  name: string
  timestamp?: string
}

export interface RelatedResources {
  type: ResourceType
  name: string
  namespace?: string
  apiVersion?: string
}

export interface Cluster {
  id: number
  name: string
  description?: string
  version?: string
  config?: string
  enabled: boolean
  inCluster: boolean
  isDefault: boolean
  createdAt: string
  updatedAt: string
  prometheusURL?: string
  error?: string
  skipSystemSync?: boolean
}

export interface OAuthProvider {
  id: number
  name: string
  clientId: string
  clientSecret: string
  authUrl?: string
  tokenUrl?: string
  userInfoUrl?: string
  scopes?: string
  issuer?: string
  enabled: boolean
  createdAt: string
  updatedAt: string
}

export interface RoleAssignment {
  id: number
  roleId: number
  subjectType: 'user' | 'group'
  subject: string
  createdAt: string
  updatedAt: string
}

export interface Role {
  id: number
  name: string
  description?: string
  isSystem?: boolean
  clusters: string[]
  namespaces: string[]
  resources: string[]
  verbs: string[]
  assignments?: RoleAssignment[]
  createdAt: string
  updatedAt: string
}

export interface UserItem {
  id: number
  username: string
  provider: string
  createdAt: string
  lastLoginAt?: string
  enabled?: boolean
  avatar_url?: string
  name?: string
  roles?: Role[]
}

export interface FetchUserListResponse {
  users: UserItem[]
  total: number
  page: number
  size: number
}

export interface PersonalAccessToken {
  id: number
  userId: number
  name: string
  prefix: string
  expiresAt?: string
  lastUsedAt?: string
  lastUsedIP?: string
  createdAt: string
  updatedAt: string
  roles?: Role[]
  user?: UserItem
}

// Alias for backward compatibility during migration
export type APIKey = PersonalAccessToken

// Resource History types
export interface ResourceHistory {
  id: number
  clusterName: string
  resourceType: string
  resourceName: string
  namespace: string
  operationType: string
  resourceYaml: string
  previousYaml: string
  success: boolean
  errorMessage: string
  operatorId: number
  operator: {
    username: string
    provider: string
  }
  createdAt: string
  updatedAt: string
}

export interface ResourceHistoryResponse {
  data: ResourceHistory[]
  pagination: {
    page: number
    pageSize: number
    total: number
    totalPages: number
    hasNextPage: boolean
    hasPrevPage: boolean
  }
}

export interface AuditLogResponse {
  data: ResourceHistory[]
  total: number
  page: number
  size: number
}

export interface GitlabHost {
  id: number
  gitlab_host: string
  is_https: boolean
  created_at: string
  updated_at: string
}

export interface UserGitlabConfig {
  id: number
  user_id: number
  gitlab_host_id: number
  token: string
  is_validated: boolean
  gitlab_host: GitlabHost
  created_at: string
  updated_at: string
}

export interface ResourceTemplate {
  ID: number
  name: string
  description: string
  yaml: string
}

export interface Anomaly {
  severity: 'critical' | 'high' | 'medium' | 'low' | 'info'
  title: string
  message: string
  remediation?: string
  ruleId: string
  docUrl?: string
}

export interface ResourceAnalysis {
  anomalies: Anomaly[]
  summary?: string
  score?: number
}

// Extend existing types with analysis
export interface WithAnalysis {
  analysis?: ResourceAnalysis
}

export interface UserAWSConfig {
  user_id: number
  credentials_content: string
  createdAt: string
  updatedAt: string
}
