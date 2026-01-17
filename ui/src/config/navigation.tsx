import {
    Home,
    Layers,
    HardDrive,
    Box,
    Boxes,
    Server,
    Database,
    PlayCircle,
    Clock,
    FileCode,
    Lock,
    Scale,
    Zap,
    Activity,
    ShieldCheck,
    ArrowUpCircle,
    Cpu,
    Key,
    Grid,
    Globe,
    Network,
    Shield,
    Share2,
    Link2,
    UserCheck,
    AlertCircle,
    Settings,
    Cloud,
    History,
    Anchor,
    CircleDashed
} from "lucide-react";

export interface NavigationItem {
    path: string;
    title: string;
    description: string;
    icon: any;
    category?: 'Workloads' | 'Config' | 'Network' | 'Storage' | 'Access Control' | 'Settings' | 'Helm' | 'Custom Resources';
    kind?: string; // Standard K8s Kind
    searchPlaceholder?: string;
    isClusterWide?: boolean;
    showHeader?: boolean;
    hideFromSidebar?: boolean;
}

export const NAVIGATION_CONFIG: NavigationItem[] = [
    {
        path: "/",
        title: "Dashboard",
        description: "Cluster overview and health status",
        icon: Home,
        showHeader: true
    },
    {
        path: "/kube-namespaces",
        title: "Namespaces",
        description: "Manage cluster namespaces",
        icon: Layers,
        kind: "Namespace",
        searchPlaceholder: "Search namespaces...",
        isClusterWide: true,
        showHeader: true
    },
    {
        path: "/kube-nodes",
        title: "Nodes",
        description: "Cluster nodes and capacity",
        icon: HardDrive,
        kind: "Node",
        searchPlaceholder: "Search nodes...",
        isClusterWide: true,
        showHeader: true
    },
    {
        path: "/kube-workload/pods",
        title: "Pods",
        description: "Manage workload instances",
        icon: Box,
        kind: "Pod",
        category: 'Workloads',
        searchPlaceholder: "Search pods...",
        showHeader: true
    },
    {
        path: "/kube-workload/deployments",
        title: "Deployments",
        description: "Manage application deployments",
        icon: Layers,
        kind: "Deployment",
        category: 'Workloads',
        searchPlaceholder: "Search deployments...",
        showHeader: true
    },
    {
        path: "/kube-workload/daemon-sets",
        title: "DaemonSets",
        description: "Manage daemon set workloads",
        icon: Server,
        kind: "DaemonSet",
        category: 'Workloads',
        searchPlaceholder: "Search daemonsets...",
        showHeader: true
    },
    {
        path: "/kube-workload/stateful-sets",
        title: "StatefulSets",
        description: "Manage stateful applications",
        icon: Database,
        kind: "StatefulSet",
        category: 'Workloads',
        searchPlaceholder: "Search statefulsets...",
        showHeader: true
    },
    {
        path: "/kube-workload/replica-sets",
        title: "ReplicaSets",
        description: "Manage replica set workloads",
        icon: Layers,
        kind: "ReplicaSet",
        category: 'Workloads',
        searchPlaceholder: "Search replicasets...",
        showHeader: true
    },
    {
        path: "/kube-workload/replication-controllers",
        title: "Replication Controllers",
        description: "Legacy workload management",
        icon: Boxes,
        kind: "ReplicationController",
        category: 'Workloads',
        searchPlaceholder: "Search replication controllers...",
        showHeader: true
    },
    {
        path: "/kube-workload/jobs",
        title: "Jobs",
        description: "Manage batch jobs",
        icon: PlayCircle,
        kind: "Job",
        category: 'Workloads',
        searchPlaceholder: "Search jobs...",
        showHeader: true
    },
    {
        path: "/kube-workload/cron-jobs",
        title: "CronJobs",
        description: "Manage scheduled jobs",
        icon: Clock,
        kind: "CronJob",
        category: 'Workloads',
        searchPlaceholder: "Search cronjobs...",
        showHeader: true
    },
    // Config
    {
        path: "/kube-crds",
        title: "Definitions",
        description: "Manage Custom Resource Definitions",
        icon: FileCode,
        category: 'Custom Resources',
        searchPlaceholder: "Search CRDs...",
        showHeader: true
    },
    {
        path: "/kube-crds/[crd]",
        title: "Custom Resources",
        description: "Manage Custom Resource Definitions",
        icon: FileCode,
        category: 'Custom Resources',
        searchPlaceholder: "Search resources...",
        showHeader: true,
        hideFromSidebar: true
    },
    {
        path: "/kube-config/config-maps",
        title: "Config Map",
        description: "Manage configuration data",
        icon: FileCode,
        kind: "ConfigMap",
        category: 'Config',
        searchPlaceholder: "Search config maps...",
        showHeader: true
    },
    {
        path: "/kube-config/secrets",
        title: "Secrets",
        description: "Manage sensitive information",
        icon: Lock,
        kind: "Secret",
        category: 'Config',
        searchPlaceholder: "Search secrets...",
        showHeader: true
    },
    {
        path: "/kube-config/resource-quotas",
        title: "Resource Quotas",
        description: "Manage resource limits",
        icon: Scale,
        kind: "ResourceQuota",
        category: 'Config',
        searchPlaceholder: "Search resource quotas...",
        showHeader: true
    },
    {
        path: "/kube-config/limit-ranges",
        title: "Limit Ranges",
        description: "Manage container resource limits",
        icon: Zap,
        kind: "LimitRange",
        category: 'Config',
        searchPlaceholder: "Search limit ranges...",
        showHeader: true
    },
    {
        path: "/kube-config/hpa",
        title: "HPA",
        description: "Horizontal Pod Autoscalers",
        icon: Activity,
        kind: "HorizontalPodAutoscaler",
        category: 'Config',
        searchPlaceholder: "Search HPA...",
        showHeader: true
    },
    {
        path: "/kube-config/pod-disruption-budgets",
        title: "PDBs",
        description: "Pod Disruption Budgets",
        icon: ShieldCheck,
        kind: "PodDisruptionBudget",
        category: 'Config',
        searchPlaceholder: "Search PDBs...",
        showHeader: true
    },
    {
        path: "/kube-config/priority-classes",
        title: "Priority Classes",
        description: "Cluster-wide priority scheduling",
        icon: ArrowUpCircle,
        kind: "PriorityClass",
        category: 'Config',
        searchPlaceholder: "Search priority classes...",
        isClusterWide: true,
        showHeader: true
    },
    {
        path: "/kube-config/runtime-classes",
        title: "Runtime Classes",
        description: "Cluster-wide container runtime configurations",
        icon: Cpu,
        kind: "RuntimeClass",
        category: 'Config',
        searchPlaceholder: "Search runtime classes...",
        isClusterWide: true,
        showHeader: true
    },
    {
        path: "/kube-config/leases",
        title: "Leases",
        description: "Distributed coordination and locking",
        icon: Key,
        kind: "Lease",
        category: 'Config',
        searchPlaceholder: "Search leases...",
        showHeader: true
    },
    {
        path: "/kube-config/mutating-webhooks",
        title: "Mutating Webhooks",
        description: "Cluster-wide mutation configurations",
        icon: Zap,
        kind: "MutatingWebhookConfiguration",
        category: 'Config',
        searchPlaceholder: "Search mutating webhooks...",
        isClusterWide: true,
        showHeader: true
    },
    {
        path: "/kube-config/validating-webhooks",
        title: "Validating Webhooks",
        description: "Cluster-wide validation configurations",
        icon: Zap,
        kind: "ValidatingWebhookConfiguration",
        category: 'Config',
        searchPlaceholder: "Search validating webhooks...",
        isClusterWide: true,
        showHeader: true
    },
    // Network
    {
        path: "/kube-network/services",
        title: "Services",
        description: "Manage networking endpoints",
        icon: Grid,
        kind: "Service",
        category: 'Network',
        searchPlaceholder: "Search services...",
        showHeader: true
    },
    {
        path: "/kube-network/endpoints",
        title: "Endpoints",
        description: "Manage service endpoints",
        icon: Network,
        kind: "Endpoints",
        category: 'Network',
        searchPlaceholder: "Search endpoints...",
        showHeader: true
    },
    {
        path: "/kube-network/ingresses",
        title: "Ingresses",
        description: "Manage external access",
        icon: Globe,
        kind: "Ingress",
        category: 'Network',
        searchPlaceholder: "Search ingresses...",
        showHeader: true
    },
    {
        path: "/kube-network/ingress-classes",
        title: "Ingress Classes",
        description: "Manage ingress controllers",
        icon: Globe,
        kind: "IngressClass",
        category: 'Network',
        searchPlaceholder: "Search ingress classes...",
        isClusterWide: true,
        showHeader: true
    },
    {
        path: "/kube-network/network-policies",
        title: "Network Policies",
        description: "Manage network security policies",
        icon: Shield,
        kind: "NetworkPolicy",
        category: 'Network',
        searchPlaceholder: "Search network policies...",
        showHeader: true
    },
    {
        path: "/kube-network/port-forwarding",
        title: "Port Forwarding",
        description: "Manage active port forwards",
        icon: Share2,
        category: 'Network',
        searchPlaceholder: "Search port forwards...",
        showHeader: true
    },
    // Storage
    {
        path: "/kube-storage/persistent-volume-claims",
        title: "Persistent Volume Claims",
        description: "Manage storage requests",
        icon: Database,
        kind: "PersistentVolumeClaim",
        category: 'Storage',
        searchPlaceholder: "Search PVCs...",
        showHeader: true
    },
    {
        path: "/kube-storage/persistent-volumes",
        title: "Persistent Volumes",
        description: "Manage cluster storage volumes",
        icon: HardDrive,
        kind: "PersistentVolume",
        category: 'Storage',
        searchPlaceholder: "Search PVs...",
        isClusterWide: true,
        showHeader: true
    },
    {
        path: "/kube-storage/storage-classes",
        title: "Storage Classes",
        description: "Manage storage provisioning",
        icon: Layers,
        kind: "StorageClass",
        category: 'Storage',
        searchPlaceholder: "Search storage classes...",
        isClusterWide: true,
        showHeader: true
    },
    // Access Control
    {
        path: "/kube-access/service-accounts",
        title: "Service Accounts",
        description: "Manage identity for processes",
        icon: UserCheck,
        kind: "ServiceAccount",
        category: 'Access Control',
        searchPlaceholder: "Search service accounts...",
        showHeader: true
    },
    {
        path: "/kube-access/cluster-roles",
        title: "Cluster Roles",
        description: "Manage cluster-wide permissions",
        icon: ShieldCheck,
        kind: "ClusterRole",
        category: 'Access Control',
        searchPlaceholder: "Search cluster roles...",
        isClusterWide: true,
        showHeader: true
    },
    {
        path: "/kube-access/roles",
        title: "Roles",
        description: "Manage namespace permissions",
        icon: Lock,
        kind: "Role",
        category: 'Access Control',
        searchPlaceholder: "Search roles...",
        showHeader: true
    },
    {
        path: "/kube-access/cluster-role-bindings",
        title: "Cluster Role Bindings",
        description: "Manage cluster-wide role assignments",
        icon: Link2,
        kind: "ClusterRoleBinding",
        category: 'Access Control',
        searchPlaceholder: "Search cluster role bindings...",
        isClusterWide: true,
        showHeader: true
    },
    {
        path: "/kube-access/role-bindings",
        title: "Role Bindings",
        description: "Manage namespace role assignments",
        icon: Link2,
        kind: "RoleBinding",
        category: 'Access Control',
        searchPlaceholder: "Search role bindings...",
        showHeader: true
    },
    {
        path: "/kube-events",
        title: "Events",
        description: "Cluster events and alerts",
        icon: AlertCircle,
        kind: "Event",
        searchPlaceholder: "Search events...",
        showHeader: true
    },
    // Settings
    {
        path: "/settings/gitlab",
        title: "GitLab Settings",
        description: "Configure GitLab integration",
        icon: Settings,
        category: 'Settings',
        showHeader: false
    },
    {
        path: "/settings/clusters",
        title: "Cluster Settings",
        description: "Manage connected clusters",
        icon: Cloud,
        category: 'Settings',
        showHeader: false
    },
    {
        path: "/settings/audit-logs",
        title: "Audit Logs",
        description: "View system audit logs",
        icon: History,
        category: 'Settings',
        showHeader: false
    },
    {
        path: "/settings/eks",
        title: "AWS EKS Settings",
        description: "Connect AWS EKS clusters",
        icon: Cloud,
        category: 'Settings',
        showHeader: false
    },
    // Helm
    {
        path: "/kube-helm/releases",
        title: "Helm Releases",
        description: "Manage Helm releases",
        icon: Anchor,
        category: 'Helm',
        showHeader: true
    },
    {
        path: "/kube-helm/charts",
        title: "Helm Charts",
        description: "Browse Helm charts",
        icon: CircleDashed,
        category: 'Helm',
        showHeader: true
    }
];
