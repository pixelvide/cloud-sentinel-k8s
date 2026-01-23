import { useMemo } from 'react'
import { useAuth } from '@/contexts/auth-context'
import { useTranslation } from 'react-i18next'

import { usePageTitle } from '@/hooks/use-page-title'
import { ResponsiveTabs } from '@/components/ui/responsive-tabs'
import { AIAdminManagement } from '@/components/settings/ai-admin-management'
import { AIConfigManagement } from '@/components/settings/ai-config-management'
import { APIKeyManagement } from '@/components/settings/apikey-management'
import { AuditLog } from '@/components/settings/audit-log'
import { AWSConfigManagement } from '@/components/settings/aws-config-management'
import { ClusterManagement } from '@/components/settings/cluster-management'
import { GitlabConfigManagement } from '@/components/settings/gitlab-config-management'
import { OAuthProviderManagement } from '@/components/settings/oauth-provider-management'
import { RBACManagement } from '@/components/settings/rbac-management'
import { TemplateManagement } from '@/components/settings/template-management'
import { UserManagement } from '@/components/settings/user-management'

export function SettingsPage() {
  const { t } = useTranslation()

  const { user } = useAuth()

  usePageTitle('Settings')

  const tabs = useMemo(() => {
    const allTabs = [
      {
        value: 'clusters',
        label: t('settings.tabs.clusters', 'Cluster'),
        content: <ClusterManagement />,
        adminOnly: true,
      },
      {
        value: 'oauth',
        label: t('settings.tabs.oauth', 'OAuth'),
        content: <OAuthProviderManagement />,
        adminOnly: true,
      },
      {
        value: 'rbac',
        label: t('settings.tabs.rbac', 'RBAC'),
        content: <RBACManagement />,
        adminOnly: true,
      },
      {
        value: 'users',
        label: t('settings.tabs.users', 'User'),
        content: <UserManagement />,
        adminOnly: true,
      },
      {
        value: 'gitlab',
        label: t('settings.tabs.gitlab', 'GitLab'),
        content: <GitlabConfigManagement />,
        adminOnly: false,
      },
      {
        value: 'aws',
        label: t('settings.tabs.aws', 'AWS'),
        content: <AWSConfigManagement />,
        adminOnly: false,
      },
      {
        value: 'apikeys',
        label: t('settings.tabs.apikeys', 'API Keys'),
        content: <APIKeyManagement />,
        adminOnly: false,
      },
      {
        value: 'ai',
        label: t('settings.tabs.ai', 'AI Assistant'),
        content: <AIConfigManagement />,
        adminOnly: false,
      },
      {
        value: 'ai-admin',
        label: t('settings.tabs.aiAdmin', 'AI Administration'),
        content: <AIAdminManagement />,
        adminOnly: true,
      },
      {
        value: 'templates',
        label: t('settings.tabs.templates', 'Templates'),
        content: <TemplateManagement />,
        adminOnly: false,
      },
      {
        value: 'audit',
        label: t('settings.tabs.audit', 'Audit'),
        content: <AuditLog />,
        adminOnly: true,
      },
    ]

    return allTabs.filter((tab) => !tab.adminOnly || user?.isAdmin())
  }, [t, user])

  return (
    <div className="space-y-2">
      <div className="mb-4">
        <div className="flex items-center gap-3 mb-2">
          <h1 className="text-3xl">{t('settings.title', 'Settings')}</h1>
        </div>
        <p className="text-muted-foreground">
          {t('settings.description', 'Manage clusters, roles and permissions')}
        </p>
      </div>

      <ResponsiveTabs tabs={tabs} />
    </div>
  )
}
