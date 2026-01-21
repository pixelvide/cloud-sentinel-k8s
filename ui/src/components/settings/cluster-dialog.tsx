import { useEffect, useState } from 'react'
import { IconEdit, IconServer } from '@tabler/icons-react'
import { useTranslation } from 'react-i18next'

import { Cluster } from '@/types/api'
import { ClusterCreateRequest, ClusterUpdateRequest, ImportClustersRequest } from '@/lib/api'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'

interface ClusterDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  cluster?: Cluster | null
  isImportMode?: boolean
  onSubmit: (clusterData: ClusterCreateRequest | ClusterUpdateRequest | ImportClustersRequest) => void
}

export function ClusterDialog({
  open,
  onOpenChange,
  cluster,
  isImportMode = false,
  onSubmit,
}: ClusterDialogProps) {
  const { t } = useTranslation()
  const isEditMode = !!cluster

  const [formData, setFormData] = useState({
    name: '',
    description: '',
    config: '',
    prometheusURL: '',
    enabled: true,
    isDefault: false,
    inCluster: false,
    skipSystemSync: false,
  })

  useEffect(() => {
    if (cluster) {
      setFormData({
        name: cluster.name,
        description: cluster.description || '',
        config: cluster.config || '',
        prometheusURL: cluster.prometheusURL || '',
        enabled: cluster.enabled,
        isDefault: cluster.isDefault,
        inCluster: cluster.inCluster,
        skipSystemSync: cluster.skipSystemSync || false,
      })
    }
  }, [cluster, open])

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (file) {
      const reader = new FileReader()
      reader.onload = (e) => {
        const content = e.target?.result as string
        handleChange('config', content)
      }
      reader.readAsText(file)
    }
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (isImportMode) {
      onSubmit({ config: formData.config, inCluster: formData.inCluster })
    } else {
      onSubmit(formData)
    }
  }

  const handleChange = (field: string, value: string | boolean) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }))
  }

  const resetForm = () => {
    setFormData({
      name: '',
      description: '',
      config: '',
      prometheusURL: '',
      enabled: true,
      isDefault: false,
      inCluster: false,
      skipSystemSync: false,
    })
  }

  const handleOpenChange = (newOpen: boolean) => {
    onOpenChange(newOpen)
    if (!newOpen && !isEditMode) {
      // Reset form when closing add dialog
      resetForm()
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {isEditMode ? (
              <IconEdit className="h-5 w-5" />
            ) : (
              <IconServer className="h-5 w-5" />
            )}
            {isEditMode
              ? t('clusterManagement.dialog.edit.title', 'Edit Cluster')
              : isImportMode
                ? t('clusterManagement.dialog.import.title', 'Import Clusters')
                : t('clusterManagement.dialog.add.title', 'Add New Cluster')}
          </DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {!isImportMode && (
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="cluster-name">
                  {t('clusterManagement.form.name.label', 'Cluster Name')} *
                </Label>
                <Input
                  id="cluster-name"
                  value={formData.name}
                  onChange={(e) => handleChange('name', e.target.value)}
                  placeholder={t(
                    'clusterManagement.form.name.placeholder',
                    'e.g., production, staging'
                  )}
                  required={!isImportMode}
                />
              </div>

              {!isEditMode && (
                <div className="space-y-2">
                  <Label htmlFor="cluster-type">
                    {t('clusterManagement.form.type.label', 'Cluster Type')}
                  </Label>
                  <Select
                    value={formData.inCluster ? 'inCluster' : 'external'}
                    onValueChange={(value) =>
                      handleChange('inCluster', value === 'inCluster')
                    }
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="external">
                        {t(
                          'clusterManagement.form.type.external',
                          'External Cluster'
                        )}
                      </SelectItem>
                      <SelectItem value="inCluster">
                        {t(
                          'clusterManagement.form.type.inCluster',
                          'In-Cluster'
                        )}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              )}
            </div>
          )}

          {!isImportMode && (
            <div className="space-y-2">
              <Label htmlFor="cluster-description">
                {t('clusterManagement.form.description.label', 'Description')}
              </Label>
              <Textarea
                id="cluster-description"
                value={formData.description}
                onChange={(e) => handleChange('description', e.target.value)}
                placeholder={t(
                  'clusterManagement.form.description.placeholder',
                  'Brief description of this cluster'
                )}
                rows={2}
              />
            </div>
          )}

          {!formData.inCluster && (
            <div className="space-y-2">
              <Label htmlFor="cluster-config">
                {t('clusterManagement.form.config.label', 'Kubeconfig')}
                {!isEditMode && ' *'}
              </Label>
              {isEditMode && (
                <p className="text-xs text-muted-foreground">
                  {t(
                    'clusterManagement.form.config.editNote',
                    'Leave empty to keep current configuration'
                  )}
                </p>
              )}

              {!isEditMode && (
                <div className="space-y-2 mb-2">
                  <Input
                    type="file"
                    onChange={handleFileSelect}
                    className="cursor-pointer"
                  />
                  <p className="text-xs text-muted-foreground">
                    {t(
                      'clusterManagement.form.config.fileHint',
                      'Or upload a kubeconfig file'
                    )}
                  </p>
                </div>
              )}

              <Textarea
                id="cluster-config"
                value={formData.config}
                onChange={(e) => handleChange('config', e.target.value)}
                placeholder={t(
                  'clusterManagement.form.kubeconfig.placeholder',
                  'Paste your kubeconfig content here...'
                )}
                rows={isImportMode ? 12 : 8}
                className="text-sm"
                required={!isEditMode && !formData.inCluster}
              />
            </div>
          )}

          {!isImportMode && (
            <div className="space-y-2">
              <Label htmlFor="prometheus-url">
                {t(
                  'clusterManagement.form.prometheusURL.label',
                  'Prometheus URL'
                )}
              </Label>
              <Input
                id="prometheus-url"
                value={formData.prometheusURL}
                onChange={(e) => handleChange('prometheusURL', e.target.value)}
                type="url"
              />
            </div>
          )}

          {/* Cluster Status Controls */}
          {!isImportMode && (
            <div className="space-y-4 border-t pt-4">
              {/* Enabled Status */}
              <div className="flex items-center justify-between">
                <div className="space-y-1">
                  <Label htmlFor="cluster-enabled">
                    {t(
                      'clusterManagement.form.enabled.label',
                      'Enable Cluster'
                    )}
                  </Label>
                </div>
                <Switch
                  id="cluster-enabled"
                  checked={formData.enabled}
                  onCheckedChange={(checked) =>
                    handleChange('enabled', checked)
                  }
                />
              </div>

              {/* Default Status */}
              <div className="flex items-center justify-between">
                <div className="space-y-1">
                  <Label htmlFor="cluster-default">
                    {t(
                      'clusterManagement.form.isDefault.label',
                      'Set as Default'
                    )}
                  </Label>
                  <p className="text-xs text-muted-foreground">
                    {t(
                      'clusterManagement.form.isDefault.help',
                      'Use this cluster as the default for new operations'
                    )}
                  </p>
                </div>
                <Switch
                  id="cluster-default"
                  checked={formData.isDefault}
                  onCheckedChange={(checked) =>
                    handleChange('isDefault', checked)
                  }
                />
              </div>

              {/* Skip System Sync */}
              <div className="flex items-center justify-between">
                <div className="space-y-1">
                  <Label htmlFor="cluster-skip-sync">
                    {t(
                      'clusterManagement.form.skipSystemSync.label',
                      'Skip System Sync'
                    )}
                  </Label>
                  <p className="text-xs text-muted-foreground">
                    {t(
                      'clusterManagement.form.skipSystemSync.help',
                      'Enable this if the cluster requires user-specific authentication and has no system-wide credentials.'
                    )}
                  </p>
                </div>
                <Switch
                  id="cluster-skip-sync"
                  checked={formData.skipSystemSync}
                  onCheckedChange={(checked) =>
                    handleChange('skipSystemSync', checked)
                  }
                />
              </div>
            </div>
          )}

          {formData.inCluster && (
            <div className="p-4 bg-blue-50 dark:bg-blue-950/20 rounded-lg border border-blue-200 dark:border-blue-800">
              <p className="text-sm text-blue-700 dark:text-blue-300">
                {t(
                  'clusterManagement.form.inCluster.note',
                  'This cluster uses the in-cluster service account configuration. No additional kubeconfig is required.'
                )}
              </p>
            </div>
          )}
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => handleOpenChange(false)}
            >
              {t('common.cancel', 'Cancel')}
            </Button>
            <Button
              type="submit"
              disabled={
                (!isImportMode && !formData.name) ||
                (!isEditMode && !formData.inCluster && !formData.config)
              }
            >
              {isEditMode
                ? t('clusterManagement.actions.save', 'Save Changes')
                : isImportMode
                  ? t('clusterManagement.actions.import', 'Import Clusters')
                  : t('clusterManagement.actions.add', 'Add Cluster')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
