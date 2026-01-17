import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { ClientLayout } from './components/ClientLayout';
import { Toaster } from 'sonner';

// Imports
import HomePage from './app/page';
import LoginPage from './app/login/page';
import ExecPage from './app/exec/page';

// Kube Access
import ClusterRolesPage from './app/kube-access/cluster-roles/page';
import ClusterRoleBindingsPage from './app/kube-access/cluster-role-bindings/page';
import RolesPage from './app/kube-access/roles/page';
import RoleBindingsPage from './app/kube-access/role-bindings/page';
import ServiceAccountsPage from './app/kube-access/service-accounts/page';

// Kube Config
import ConfigMapsPage from './app/kube-config/config-maps/page';
import SecretsPage from './app/kube-config/secrets/page';
import HPAPage from './app/kube-config/hpa/page';
import LeasesPage from './app/kube-config/leases/page';
import LimitRangesPage from './app/kube-config/limit-ranges/page';
import MutatingWebhooksPage from './app/kube-config/mutating-webhooks/page';
import PodDisruptionBudgetsPage from './app/kube-config/pod-disruption-budgets/page';
import PriorityClassesPage from './app/kube-config/priority-classes/page';
import ResourceQuotasPage from './app/kube-config/resource-quotas/page';
import RuntimeClassesPage from './app/kube-config/runtime-classes/page';
import ValidatingWebhooksPage from './app/kube-config/validating-webhooks/page';

// Kube CRDs
import CRDsPage from './app/kube-crds/page';
import CRDDetailPage from './app/kube-crds/[crd]/page';

// Kube Events
import EventsPage from './app/kube-events/page';

// Kube Helm
import ChartsPage from './app/kube-helm/charts/page';
import ReleasesPage from './app/kube-helm/releases/page';

// Kube Namespaces
import NamespacesPage from './app/kube-namespaces/page';

// Kube Network
import EndpointsPage from './app/kube-network/endpoints/page';
import IngressClassesPage from './app/kube-network/ingress-classes/page';
import IngressesPage from './app/kube-network/ingresses/page';
import NetworkPoliciesPage from './app/kube-network/network-policies/page';
import PortForwardingPage from './app/kube-network/port-forwarding/page';
import ServicesPage from './app/kube-network/services/page';

// Kube Nodes
import NodesPage from './app/kube-nodes/page';

// Kube Storage
import PVCsPage from './app/kube-storage/persistent-volume-claims/page';
import PVsPage from './app/kube-storage/persistent-volumes/page';
import StorageClassesPage from './app/kube-storage/storage-classes/page';

// Kube Workload
import PodsPage from './app/kube-workload/pods/page';
import DeploymentsPage from './app/kube-workload/deployments/page';
import DaemonSetsPage from './app/kube-workload/daemon-sets/page';
import StatefulSetsPage from './app/kube-workload/stateful-sets/page';
import ReplicaSetsPage from './app/kube-workload/replica-sets/page';
import ReplicationControllersPage from './app/kube-workload/replication-controllers/page';
import JobsPage from './app/kube-workload/jobs/page';
import CronJobsPage from './app/kube-workload/cron-jobs/page';

// Settings
import AuditLogsPage from './app/settings/audit-logs/page';
import ClustersPage from './app/settings/clusters/page';
import EKSPage from './app/settings/eks/page';
import GitlabPage from './app/settings/gitlab/page';


function App() {
    return (
        <BrowserRouter>
            <ClientLayout>
                <Routes>
                    <Route path="/" element={<HomePage />} />
                    <Route path="/login" element={<LoginPage />} />
                    <Route path="/exec" element={<ExecPage />} />

                    {/* Kube Access */}
                    <Route path="/kube-access/cluster-roles" element={<ClusterRolesPage />} />
                    <Route path="/kube-access/cluster-role-bindings" element={<ClusterRoleBindingsPage />} />
                    <Route path="/kube-access/roles" element={<RolesPage />} />
                    <Route path="/kube-access/role-bindings" element={<RoleBindingsPage />} />
                    <Route path="/kube-access/service-accounts" element={<ServiceAccountsPage />} />

                    {/* Kube Config */}
                    <Route path="/kube-config/config-maps" element={<ConfigMapsPage />} />
                    <Route path="/kube-config/secrets" element={<SecretsPage />} />
                    <Route path="/kube-config/hpa" element={<HPAPage />} />
                    <Route path="/kube-config/leases" element={<LeasesPage />} />
                    <Route path="/kube-config/limit-ranges" element={<LimitRangesPage />} />
                    <Route path="/kube-config/mutating-webhooks" element={<MutatingWebhooksPage />} />
                    <Route path="/kube-config/pod-disruption-budgets" element={<PodDisruptionBudgetsPage />} />
                    <Route path="/kube-config/priority-classes" element={<PriorityClassesPage />} />
                    <Route path="/kube-config/resource-quotas" element={<ResourceQuotasPage />} />
                    <Route path="/kube-config/runtime-classes" element={<RuntimeClassesPage />} />
                    <Route path="/kube-config/validating-webhooks" element={<ValidatingWebhooksPage />} />

                    {/* Kube CRDs */}
                    <Route path="/kube-crds" element={<CRDsPage />} />
                    <Route path="/kube-crds/:crd" element={<CRDDetailPage />} />

                    {/* Kube Events */}
                    <Route path="/kube-events" element={<EventsPage />} />

                    {/* Kube Helm */}
                    <Route path="/kube-helm/charts" element={<ChartsPage />} />
                    <Route path="/kube-helm/releases" element={<ReleasesPage />} />

                    {/* Kube Namespaces */}
                    <Route path="/kube-namespaces" element={<NamespacesPage />} />

                    {/* Kube Network */}
                    <Route path="/kube-network/endpoints" element={<EndpointsPage />} />
                    <Route path="/kube-network/ingress-classes" element={<IngressClassesPage />} />
                    <Route path="/kube-network/ingresses" element={<IngressesPage />} />
                    <Route path="/kube-network/network-policies" element={<NetworkPoliciesPage />} />
                    <Route path="/kube-network/port-forwarding" element={<PortForwardingPage />} />
                    <Route path="/kube-network/services" element={<ServicesPage />} />

                    {/* Kube Nodes */}
                    <Route path="/kube-nodes" element={<NodesPage />} />

                    {/* Kube Storage */}
                    <Route path="/kube-storage/persistent-volume-claims" element={<PVCsPage />} />
                    <Route path="/kube-storage/persistent-volumes" element={<PVsPage />} />
                    <Route path="/kube-storage/storage-classes" element={<StorageClassesPage />} />

                    {/* Kube Workload */}
                    <Route path="/kube-workload/pods" element={<PodsPage />} />
                    <Route path="/kube-workload/deployments" element={<DeploymentsPage />} />
                    <Route path="/kube-workload/daemon-sets" element={<DaemonSetsPage />} />
                    <Route path="/kube-workload/stateful-sets" element={<StatefulSetsPage />} />
                    <Route path="/kube-workload/replica-sets" element={<ReplicaSetsPage />} />
                    <Route path="/kube-workload/replication-controllers" element={<ReplicationControllersPage />} />
                    <Route path="/kube-workload/jobs" element={<JobsPage />} />
                    <Route path="/kube-workload/cron-jobs" element={<CronJobsPage />} />

                    {/* Settings */}
                    <Route path="/settings/audit-logs" element={<AuditLogsPage />} />
                    <Route path="/settings/clusters" element={<ClustersPage />} />
                    <Route path="/settings/eks" element={<EKSPage />} />
                    <Route path="/settings/gitlab" element={<GitlabPage />} />
                </Routes>
                <Toaster position="bottom-left" richColors />
            </ClientLayout>
        </BrowserRouter>
    );
}

export default App;
