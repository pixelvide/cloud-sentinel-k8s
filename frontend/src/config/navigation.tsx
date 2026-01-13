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
    AlertCircle,
    Settings,
    Cloud,
    History
} from "lucide-react";

export interface NavigationItem {
    path: string;
    title: string;
    description: string;
    icon: any;
    category?: 'Workloads' | 'Config' | 'Network' | 'Settings';
    searchPlaceholder?: string;
    isClusterWide?: boolean;
    showHeader?: boolean;
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
        path: "/namespaces",
        title: "Namespaces",
        description: "Manage cluster namespaces",
        icon: Layers,
        searchPlaceholder: "Search namespaces...",
        isClusterWide: true,
        showHeader: true
    },
    {
        path: "/nodes",
        title: "Nodes",
        description: "Cluster nodes and capacity",
        icon: HardDrive,
        searchPlaceholder: "Search nodes...",
        isClusterWide: true,
        showHeader: true
    },
    // Workloads
    {
        path: "/pods",
        title: "Pods",
        description: "Manage workload instances",
        icon: Box,
        category: 'Workloads',
        searchPlaceholder: "Search pods...",
        showHeader: true
    },
    {
        path: "/deployments",
        title: "Deployments",
        description: "Manage application deployments",
        icon: Layers,
        category: 'Workloads',
        searchPlaceholder: "Search deployments...",
        showHeader: true
    },
    {
        path: "/daemonsets",
        title: "DaemonSets",
        description: "Manage daemon set workloads",
        icon: Server,
        category: 'Workloads',
        searchPlaceholder: "Search daemonsets...",
        showHeader: true
    },
    {
        path: "/statefulsets",
        title: "StatefulSets",
        description: "Manage stateful applications",
        icon: Database,
        category: 'Workloads',
        searchPlaceholder: "Search statefulsets...",
        showHeader: true
    },
    {
        path: "/replicasets",
        title: "ReplicaSets",
        description: "Manage replica set workloads",
        icon: Layers,
        category: 'Workloads',
        searchPlaceholder: "Search replicasets...",
        showHeader: true
    },
    {
        path: "/replicationcontrollers",
        title: "Replication Controllers",
        description: "Legacy workload management",
        icon: Boxes,
        category: 'Workloads',
        searchPlaceholder: "Search replication controllers...",
        showHeader: true
    },
    {
        path: "/jobs",
        title: "Jobs",
        description: "Manage batch jobs",
        icon: PlayCircle,
        category: 'Workloads',
        searchPlaceholder: "Search jobs...",
        showHeader: true
    },
    {
        path: "/cronjobs",
        title: "CronJobs",
        description: "Manage scheduled jobs",
        icon: Clock,
        category: 'Workloads',
        searchPlaceholder: "Search cronjobs...",
        showHeader: true
    },
    // Config
    {
        path: "/configmaps",
        title: "Config Maps",
        description: "Manage configuration data",
        icon: FileCode,
        category: 'Config',
        searchPlaceholder: "Search config maps...",
        showHeader: true
    },
    {
        path: "/secrets",
        title: "Secrets",
        description: "Manage sensitive information",
        icon: Lock,
        category: 'Config',
        searchPlaceholder: "Search secrets...",
        showHeader: true
    },
    {
        path: "/resourcequotas",
        title: "Resource Quotas",
        description: "Manage resource limits",
        icon: Scale,
        category: 'Config',
        searchPlaceholder: "Search resource quotas...",
        showHeader: true
    },
    {
        path: "/limitranges",
        title: "Limit Ranges",
        description: "Manage container resource limits",
        icon: Zap,
        category: 'Config',
        searchPlaceholder: "Search limit ranges...",
        showHeader: true
    },
    {
        path: "/hpa",
        title: "HPA",
        description: "Horizontal Pod Autoscalers",
        icon: Activity,
        category: 'Config',
        searchPlaceholder: "Search HPA...",
        showHeader: true
    },
    {
        path: "/pdbs",
        title: "PDBs",
        description: "Pod Disruption Budgets",
        icon: ShieldCheck,
        category: 'Config',
        searchPlaceholder: "Search PDBs...",
        showHeader: true
    },
    {
        path: "/priorityclasses",
        title: "Priority Classes",
        description: "Cluster-wide priority scheduling",
        icon: ArrowUpCircle,
        category: 'Config',
        searchPlaceholder: "Search priority classes...",
        isClusterWide: true,
        showHeader: true
    },
    {
        path: "/runtimeclasses",
        title: "Runtime Classes",
        description: "Cluster-wide container runtime configurations",
        icon: Cpu,
        category: 'Config',
        searchPlaceholder: "Search runtime classes...",
        isClusterWide: true,
        showHeader: true
    },
    {
        path: "/leases",
        title: "Leases",
        description: "Distributed coordination and locking",
        icon: Key,
        category: 'Config',
        searchPlaceholder: "Search leases...",
        showHeader: true
    },
    {
        path: "/mutatingwebhooks",
        title: "Mutating Webhooks",
        description: "Cluster-wide mutation configurations",
        icon: Zap,
        category: 'Config',
        searchPlaceholder: "Search mutating webhooks...",
        isClusterWide: true,
        showHeader: true
    },
    {
        path: "/validatingwebhooks",
        title: "Validating Webhooks",
        description: "Cluster-wide validation configurations",
        icon: Zap,
        category: 'Config',
        searchPlaceholder: "Search validating webhooks...",
        isClusterWide: true,
        showHeader: true
    },
    // Network
    {
        path: "/services",
        title: "Services",
        description: "Manage networking endpoints",
        icon: Grid,
        category: 'Network',
        searchPlaceholder: "Search services...",
        showHeader: true
    },
    {
        path: "/endpoints",
        title: "Endpoints",
        description: "Manage service endpoints",
        icon: Network,
        category: 'Network',
        searchPlaceholder: "Search endpoints...",
        showHeader: true
    },
    {
        path: "/ingresses",
        title: "Ingresses",
        description: "Manage external access",
        icon: Globe,
        category: 'Network',
        searchPlaceholder: "Search ingresses...",
        showHeader: true
    },
    {
        path: "/ingressclasses",
        title: "Ingress Classes",
        description: "Manage ingress controllers",
        icon: Globe,
        category: 'Network',
        searchPlaceholder: "Search ingress classes...",
        isClusterWide: true,
        showHeader: true
    },
    {
        path: "/networkpolicies",
        title: "Network Policies",
        description: "Manage network security policies",
        icon: Shield,
        category: 'Network',
        searchPlaceholder: "Search network policies...",
        showHeader: true
    },
    {
        path: "/portforwarding",
        title: "Port Forwarding",
        description: "Manage active port forwards",
        icon: Share2,
        category: 'Network',
        searchPlaceholder: "Search port forwards...",
        showHeader: true
    },
    // Top-level Events
    {
        path: "/events",
        title: "Events",
        description: "Cluster events and alerts",
        icon: AlertCircle,
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
    }
];
