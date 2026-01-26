import { createBrowserRouter, Navigate } from 'react-router-dom'

import App from './App'
import { InitCheckRoute } from './components/init-check-route'
import { ProtectedRoute } from './components/protected-route'
import {
  ClusterRedirector,
  RootRedirector,
} from './components/route-redirectors'
import { getSubPath } from './lib/subpath'
import { CRListPage } from './pages/cr-list-page'
import { HelmChartListPage } from './pages/helm-chart-list-page'
import { HelmReleaseListPage } from './pages/helm-release-list-page'
import { InitializationPage } from './pages/initialization'
import { LoginPage } from './pages/login'
import { Overview } from './pages/overview'
import { ResourceDetail } from './pages/resource-detail'
import { ResourceList } from './pages/resource-list'
import { SettingsPage } from './pages/settings'
import { SecurityDashboard } from './pages/security-dashboard'

const subPath = getSubPath()

export const router = createBrowserRouter(
  [
    {
      path: '/setup',
      element: <InitializationPage />,
    },
    {
      path: '/login',
      element: (
        <InitCheckRoute>
          <LoginPage />
        </InitCheckRoute>
      ),
    },
    {
      path: '/',
      element: (
        <InitCheckRoute>
          <ProtectedRoute>
            <App />
          </ProtectedRoute>
        </InitCheckRoute>
      ),
      children: [
        {
          index: true,
          element: <RootRedirector />,
        },
        {
          path: 'settings',
          element: <SettingsPage />,
        },
        {
          path: 'c/:cluster',
          children: [
            {
              index: true,
              element: <Navigate to="dashboard" replace />,
            },
            {
              path: 'dashboard',
              element: <Overview />,
            },
            {
              path: 'security',
              element: <SecurityDashboard />,
            },
            {
              path: 'helm-releases',
              element: <HelmReleaseListPage />,
            },
            {
              path: 'helm-charts',
              element: <HelmChartListPage />,
            },
            {
              path: 'crds/:crd',
              element: <CRListPage />,
            },
            {
              path: 'crds/:resource/:namespace/:name',
              element: <ResourceDetail />,
            },
            {
              path: 'crds/:resource/:name',
              element: <ResourceDetail />,
            },
            {
              path: ':resource/:name',
              element: <ResourceDetail />,
            },
            {
              path: ':resource',
              element: <ResourceList />,
            },
            {
              path: ':resource/:namespace/:name',
              element: <ResourceDetail />,
            },
          ],
        },
        {
          // Catch-all for legacy/absolute paths that forgot the cluster prefix
          path: '*',
          element: <ClusterRedirector />,
        },
      ],
    },
  ],
  {
    basename: subPath,
  }
)
