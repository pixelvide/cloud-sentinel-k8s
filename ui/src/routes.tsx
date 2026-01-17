import { createBrowserRouter, Navigate } from "react-router-dom";

// Imports - using relative paths for now, will update if needed
import HomePage from "./pages/page";
import LoginPage from "./pages/login/page";
import ExecPage from "./pages/exec/page";

// Kube Access
import ClusterRolesPage from "./pages/kube-access/cluster-roles/page";
import ClusterRoleBindingsPage from "./pages/kube-access/cluster-role-bindings/page";
import RolesPage from "./pages/kube-access/roles/page";
import RoleBindingsPage from "./pages/kube-access/role-bindings/page";
import ServiceAccountsPage from "./pages/kube-access/service-accounts/page";

// Kube Config
import ConfigMapsPage from "./pages/kube-config/config-maps/page";
import SecretsPage from "./pages/kube-config/secrets/page";
import HPAPage from "./pages/kube-config/hpa/page";
import LeasesPage from "./pages/kube-config/leases/page";
import LimitRangesPage from "./pages/kube-config/limit-ranges/page";
import MutatingWebhooksPage from "./pages/kube-config/mutating-webhooks/page";
import PodDisruptionBudgetsPage from "./pages/kube-config/pod-disruption-budgets/page";
import PriorityClassesPage from "./pages/kube-config/priority-classes/page";
import ResourceQuotasPage from "./pages/kube-config/resource-quotas/page";
import RuntimeClassesPage from "./pages/kube-config/runtime-classes/page";
import ValidatingWebhooksPage from "./pages/kube-config/validating-webhooks/page";

// Kube CRDs
import CRDsPage from "./pages/kube-crds/page";
import CRDDetailPage from "./pages/kube-crds/[crd]/page";

// Kube Events
import EventsPage from "./pages/kube-events/page";

// Kube Helm
import ChartsPage from "./pages/kube-helm/charts/page";
import ReleasesPage from "./pages/kube-helm/releases/page";

// Kube Namespaces
import NamespacesPage from "./pages/kube-namespaces/page";

// Kube Network
import EndpointsPage from "./pages/kube-network/endpoints/page";
import IngressClassesPage from "./pages/kube-network/ingress-classes/page";
import IngressesPage from "./pages/kube-network/ingresses/page";
import NetworkPoliciesPage from "./pages/kube-network/network-policies/page";
import PortForwardingPage from "./pages/kube-network/port-forwarding/page";
import ServicesPage from "./pages/kube-network/services/page";

// Kube Nodes
import NodesPage from "./pages/kube-nodes/page";

// Kube Storage
import PVCsPage from "./pages/kube-storage/persistent-volume-claims/page";
import PVsPage from "./pages/kube-storage/persistent-volumes/page";
import StorageClassesPage from "./pages/kube-storage/storage-classes/page";

// Kube Workload
import PodsPage from "./pages/kube-workload/pods/page";
import DeploymentsPage from "./pages/kube-workload/deployments/page";
import DaemonSetsPage from "./pages/kube-workload/daemon-sets/page";
import StatefulSetsPage from "./pages/kube-workload/stateful-sets/page";
import ReplicaSetsPage from "./pages/kube-workload/replica-sets/page";
import ReplicationControllersPage from "./pages/kube-workload/replication-controllers/page";
import JobsPage from "./pages/kube-workload/jobs/page";
import CronJobsPage from "./pages/kube-workload/cron-jobs/page";

// Settings
import AuditLogsPage from "./pages/settings/audit-logs/page";
import ClustersPage from "./pages/settings/clusters/page";
import EKSPage from "./pages/settings/eks/page";
import GitlabPage from "./pages/settings/gitlab/page";

import App from "./App";

export const router = createBrowserRouter([
    {
        path: "/login",
        element: <LoginPage />,
    },
    {
        path: "/",
        element: <App />,
        children: [
            {
                index: true,
                element: <HomePage />,
            },
            {
                path: "exec",
                element: <ExecPage />,
            },
            {
                path: "kube-access",
                children: [
                    { path: "cluster-roles", element: <ClusterRolesPage /> },
                    { path: "cluster-role-bindings", element: <ClusterRoleBindingsPage /> },
                    { path: "roles", element: <RolesPage /> },
                    { path: "role-bindings", element: <RoleBindingsPage /> },
                    { path: "service-accounts", element: <ServiceAccountsPage /> },
                ]
            },
            {
                path: "kube-config",
                children: [
                    { path: "config-maps", element: <ConfigMapsPage /> },
                    { path: "secrets", element: <SecretsPage /> },
                    { path: "hpa", element: <HPAPage /> },
                    { path: "leases", element: <LeasesPage /> },
                    { path: "limit-ranges", element: <LimitRangesPage /> },
                    { path: "mutating-webhooks", element: <MutatingWebhooksPage /> },
                    { path: "pod-disruption-budgets", element: <PodDisruptionBudgetsPage /> },
                    { path: "priority-classes", element: <PriorityClassesPage /> },
                    { path: "resource-quotas", element: <ResourceQuotasPage /> },
                    { path: "runtime-classes", element: <RuntimeClassesPage /> },
                    { path: "validating-webhooks", element: <ValidatingWebhooksPage /> },
                ]
            },
            {
                path: "kube-crds",
                children: [
                    { index: true, element: <CRDsPage /> },
                    { path: ":crd", element: <CRDDetailPage /> },
                ]
            },
            {
                path: "kube-events",
                element: <EventsPage />,
            },
            {
                path: "kube-helm",
                children: [
                    { path: "charts", element: <ChartsPage /> },
                    { path: "releases", element: <ReleasesPage /> },
                ]
            },
            {
                path: "kube-namespaces",
                element: <NamespacesPage />,
            },
            {
                path: "kube-network",
                children: [
                    { path: "endpoints", element: <EndpointsPage /> },
                    { path: "ingress-classes", element: <IngressClassesPage /> },
                    { path: "ingresses", element: <IngressesPage /> },
                    { path: "network-policies", element: <NetworkPoliciesPage /> },
                    { path: "port-forwarding", element: <PortForwardingPage /> },
                    { path: "services", element: <ServicesPage /> },
                ]
            },
            {
                path: "kube-nodes",
                element: <NodesPage />,
            },
            {
                path: "kube-storage",
                children: [
                    { path: "persistent-volume-claims", element: <PVCsPage /> },
                    { path: "persistent-volumes", element: <PVsPage /> },
                    { path: "storage-classes", element: <StorageClassesPage /> },
                ]
            },
            {
                path: "kube-workload",
                children: [
                    { path: "pods", element: <PodsPage /> },
                    { path: "deployments", element: <DeploymentsPage /> },
                    { path: "daemon-sets", element: <DaemonSetsPage /> },
                    { path: "stateful-sets", element: <StatefulSetsPage /> },
                    { path: "replica-sets", element: <ReplicaSetsPage /> },
                    { path: "replication-controllers", element: <ReplicationControllersPage /> },
                    { path: "jobs", element: <JobsPage /> },
                    { path: "cron-jobs", element: <CronJobsPage /> },
                ]
            },
            {
                path: "settings",
                children: [
                    { path: "audit-logs", element: <AuditLogsPage /> },
                    { path: "clusters", element: <ClustersPage /> },
                    { path: "eks", element: <EKSPage /> },
                    { path: "gitlab", element: <GitlabPage /> },
                ]
            },
        ],
    },
]);
