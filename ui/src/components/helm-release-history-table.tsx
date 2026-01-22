import { useEffect, useMemo, useState } from 'react'
import { IconRotateClockwise } from '@tabler/icons-react'
import { createColumnHelper } from '@tanstack/react-table'
import { Loader2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { HelmRelease } from '@/types/api'
import { withSubPath } from '@/lib/subpath'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { ResourceTable } from '@/components/resource-table'

interface HelmReleaseHistoryTableProps {
  namespace: string
  name: string
}

export function HelmReleaseHistoryTable({
  namespace,
  name,
}: HelmReleaseHistoryTableProps) {
  const { t } = useTranslation()
  const [rollbackRevision, setRollbackRevision] = useState<number | null>(null)
  const [isRollingBack, setIsRollingBack] = useState(false)
  const [refreshKey, setRefreshKey] = useState(0)
  const [history, setHistory] = useState<HelmRelease[]>([])
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    const fetchHistory = async () => {
      setIsLoading(true)
      try {
        const response = await fetch(
          withSubPath(`/api/v1/helmreleases/${namespace}/${name}/history`)
        )
        if (response.ok) {
          const data = await response.json()
          setHistory(data.items || [])
        } else {
          console.error('Failed to fetch history')
          toast.error('Failed to fetch release history')
        }
      } catch (error) {
        console.error('Failed to fetch history', error)
        toast.error('Failed to fetch release history')
      } finally {
        setIsLoading(false)
      }
    }
    fetchHistory()
  }, [namespace, name, refreshKey])

  const handleRollback = async () => {
    if (!rollbackRevision) return

    setIsRollingBack(true)
    try {
      const response = await fetch(
        withSubPath(`/api/v1/helmreleases/${namespace}/${name}/rollback`),
        {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ revision: rollbackRevision }),
        }
      )

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        throw new Error(errorData.error || 'Failed to rollback')
      }

      toast.success(
        t('helm_release.rollback_success', { revision: rollbackRevision })
      )
      setRollbackRevision(null)
      setRefreshKey((k) => k + 1)
    } catch (error) {
      toast.error(
        t('helm_release.rollback_failed', {
          error: error instanceof Error ? error.message : String(error),
        })
      )
      console.error(error)
    } finally {
      setIsRollingBack(false)
    }
  }

  const columnHelper = createColumnHelper<HelmRelease>()

  const columns = useMemo(
    () => [
      columnHelper.accessor('revision', {
        header: t('helm_release.revision', 'Revision'),
        cell: (info) => <div className="font-medium">{info.getValue()}</div>,
      }),
      columnHelper.accessor('updated', {
        header: t('common.updated', 'Updated'),
        cell: (info) => (
          <div className="text-muted-foreground">
            {new Date(info.getValue()).toLocaleString()}
          </div>
        ),
      }),
      columnHelper.accessor('status', {
        header: t('common.status', 'Status'),
        cell: (info) => (
          <Badge
            variant={info.getValue() === 'deployed' ? 'default' : 'secondary'}
          >
            {info.getValue()}
          </Badge>
        ),
      }),
      columnHelper.accessor('chart', {
        header: t('helm_release.chart', 'Chart'),
      }),
      columnHelper.accessor('app_version', {
        header: t('helm_release.app_version', 'App Version'),
      }),
      columnHelper.display({
        id: 'actions',
        header: t('common.actions', 'Actions'),
        cell: ({ row }) => {
          if (row.original.status === 'deployed') return null

          return (
            <Button
              variant="ghost"
              size="icon"
              onClick={() => setRollbackRevision(row.original.revision)}
              title={t('helm_release.rollback', 'Rollback')}
            >
              <IconRotateClockwise className="h-4 w-4" />
            </Button>
          )
        },
      }),
    ],
    [columnHelper, t]
  )

  return (
    <>
      <ResourceTable
        resourceName="helmreleases"
        columns={columns}
        data={history}
        isLoading={isLoading}
        hideFilter
        disablePagination
      />

      <Dialog
        open={!!rollbackRevision}
        onOpenChange={(open) => !open && setRollbackRevision(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {t('helm_release.rollback_title', 'Rollback Release')}
            </DialogTitle>
            <DialogDescription>
              {t(
                'helm_release.rollback_description',
                'Are you sure you want to rollback to revision {{revision}}?',
                {
                  revision: rollbackRevision,
                }
              )}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setRollbackRevision(null)}
              disabled={isRollingBack}
            >
              {t('common.cancel', 'Cancel')}
            </Button>
            <Button onClick={handleRollback} disabled={isRollingBack}>
              {isRollingBack && (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              )}
              {t('common.confirm', 'Confirm')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
