import { useCallback, useMemo } from 'react'
import { createColumnHelper } from '@tanstack/react-table'
import { useTranslation } from 'react-i18next'

import { HelmRelease } from '@/types/api'
import { Badge } from '@/components/ui/badge'
import { ResourceTable } from '@/components/resource-table'

export function HelmReleaseListPage() {
  const { t } = useTranslation()

  // Define column helper
  const columnHelper = createColumnHelper<HelmRelease>()

  const columns = useMemo(
    () => [
      columnHelper.accessor('name', {
        header: t('common.name'),
        cell: (info) => <div className="font-medium">{info.getValue()}</div>,
      }),
      columnHelper.accessor('revision', {
        header: 'Revision',
      }),
      columnHelper.accessor('status', {
        header: t('common.status'),
        cell: (info) => (
          <Badge
            variant={info.getValue() === 'deployed' ? 'default' : 'secondary'}
          >
            {info.getValue()}
          </Badge>
        ),
      }),
      columnHelper.accessor('chart', {
        header: 'Chart',
      }),
      columnHelper.accessor('app_version', {
        header: 'App Version',
      }),
      columnHelper.accessor('updated', {
        header: 'Updated',
      }),
    ],
    [columnHelper, t]
  )

  const searchFilter = useCallback((item: HelmRelease, query: string) => {
    return (
      item.name.toLowerCase().includes(query) ||
      item.chart.toLowerCase().includes(query)
    )
  }, [])

  return (
    <ResourceTable
      resourceName="HelmReleases"
      columns={columns}
      searchQueryFilter={searchFilter}
    />
  )
}
