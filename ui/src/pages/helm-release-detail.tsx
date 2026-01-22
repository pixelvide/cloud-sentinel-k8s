import { useState } from 'react'
import { IconLoader, IconRefresh, IconTrash } from '@tabler/icons-react'
import { useParams } from 'react-router-dom'

import { HelmRelease } from '@/types/api'
import { useResource } from '@/lib/api'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { ResponsiveTabs } from '@/components/ui/responsive-tabs'
import { ErrorMessage } from '@/components/error-message'
import { HelmReleaseHistoryTable } from '@/components/helm-release-history-table'
import { ResourceDeleteConfirmationDialog } from '@/components/resource-delete-confirmation-dialog'
import { YamlEditor } from '@/components/yaml-editor'

export function HelmReleaseDetail() {
  const { namespace, name } = useParams()
  const [refreshKey, setRefreshKey] = useState(0)
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)

  const {
    data: release,
    isLoading,
    isError,
    error,
    refetch: handleRefresh,
  } = useResource('helmreleases', name!, namespace)

  const helmRelease = release as unknown as HelmRelease

  const handleManualRefresh = async () => {
    setRefreshKey((prev) => prev + 1)
    await handleRefresh()
  }

  if (isLoading) {
    return (
      <div className="p-6">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-center gap-2">
              <IconLoader className="animate-spin" />
              <span>Loading Helm Release details...</span>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (isError || !helmRelease) {
    return (
      <ErrorMessage
        resourceName="Helm Release"
        error={error}
        refetch={handleRefresh}
      />
    )
  }

  return (
    <div className="space-y-2">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-bold">{name}</h1>
          {namespace && (
            <p className="text-muted-foreground">
              Namespace: <span className="font-medium">{namespace}</span>
            </p>
          )}
        </div>
        <div className="flex gap-2">
          <Button
            disabled={isLoading}
            variant="outline"
            size="sm"
            onClick={handleManualRefresh}
          >
            <IconRefresh className="w-4 h-4" />
            Refresh
          </Button>
          <Button
            variant="destructive"
            size="sm"
            onClick={() => setIsDeleteDialogOpen(true)}
          >
            <IconTrash className="w-4 h-4" />
            Delete
          </Button>
        </div>
      </div>

      <ResponsiveTabs
        tabs={[
          {
            value: 'overview',
            label: 'Overview',
            content: (
              <div className="space-y-6">
                <Card>
                  <CardHeader>
                    <CardTitle>Information</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div>
                        <Label className="text-xs text-muted-foreground">
                          Status
                        </Label>
                        <div className="mt-1">
                          <Badge
                            variant={
                              helmRelease.status === 'deployed'
                                ? 'default'
                                : 'secondary'
                            }
                          >
                            {helmRelease.status}
                          </Badge>
                        </div>
                      </div>
                      <div>
                        <Label className="text-xs text-muted-foreground">
                          Revision
                        </Label>
                        <p className="text-sm">{helmRelease.revision}</p>
                      </div>
                      <div>
                        <Label className="text-xs text-muted-foreground">
                          Chart
                        </Label>
                        <p className="text-sm">{helmRelease.chart}</p>
                      </div>
                      <div>
                        <Label className="text-xs text-muted-foreground">
                          App Version
                        </Label>
                        <p className="text-sm">{helmRelease.app_version}</p>
                      </div>
                      <div>
                        <Label className="text-xs text-muted-foreground">
                          Updated
                        </Label>
                        <p className="text-sm">
                          {new Date(helmRelease.updated).toLocaleString()}
                        </p>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </div>
            ),
          },
          {
            value: 'values',
            label: 'Values',
            content: (
              <div className="space-y-4">
                <YamlEditor
                  key={`values-${refreshKey}`}
                  value={helmRelease.values || ''}
                  title="Values"
                  onChange={() => {}}
                  readOnly
                />
              </div>
            ),
          },
          {
            value: 'notes',
            label: 'Notes',
            content: (
              <Card>
                <CardContent className="pt-6 overflow-auto">
                  <pre className="whitespace-pre-wrap font-mono text-sm">
                    {helmRelease.notes || 'No notes available'}
                  </pre>
                </CardContent>
              </Card>
            ),
          },
          {
            value: 'history',
            label: 'History',
            content: (
              <HelmReleaseHistoryTable namespace={namespace!} name={name!} />
            ),
          },
          {
            value: 'manifest',
            label: 'Manifest',
            content: (
              <div className="space-y-4">
                <YamlEditor
                  key={`manifest-${refreshKey}`}
                  value={helmRelease.manifest || ''}
                  title="Manifest"
                  onChange={() => {}}
                  readOnly
                />
              </div>
            ),
          },
        ]}
      />

      <ResourceDeleteConfirmationDialog
        open={isDeleteDialogOpen}
        onOpenChange={setIsDeleteDialogOpen}
        resourceName={name!}
        resourceType="helmreleases"
        namespace={namespace}
      />
    </div>
  )
}
