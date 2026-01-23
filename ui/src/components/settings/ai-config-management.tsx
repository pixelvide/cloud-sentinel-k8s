import { useCallback, useEffect, useState } from 'react'
import { IconRobot, IconStar, IconStarFilled, IconTrash, IconPlus } from '@tabler/icons-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { fetchAIProfiles, listAIConfigs, updateAIConfig, deleteAIConfig } from '@/lib/api'
import { AIProviderProfile, AISettings } from '@/types/ai'
import { useAuth } from '@/contexts/auth-context'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'

export function AIConfigManagement() {
  const { t } = useTranslation()
  const { checkAuth } = useAuth()
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [profiles, setProfiles] = useState<AIProviderProfile[]>([])
  const [userConfigs, setUserConfigs] = useState<AISettings[]>([])
  const [editingConfig, setEditingConfig] = useState<Partial<AISettings> | null>(null)

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const [p, c] = await Promise.all([fetchAIProfiles(), listAIConfigs()])
      setProfiles(p)
      setUserConfigs(c)
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadData()
  }, [loadData])

  const handleSave = useCallback(async () => {
    if (!editingConfig || !editingConfig.profileID) return
    setSaving(true)
    try {
      await updateAIConfig(editingConfig)
      await loadData()
      await checkAuth()
      setEditingConfig(null)
      toast.success(t('aiConfig.saved', 'AI configuration saved successfully'))
    } catch (error) {
      toast.error(t('aiConfig.saveError', 'Failed to save configuration'))
      console.error(error)
    } finally {
      setSaving(false)
    }
  }, [editingConfig, loadData, checkAuth, t])

  const handleSetDefault = useCallback(async (config: AISettings) => {
    try {
      await updateAIConfig({ ...config, isDefault: true })
      await loadData()
      toast.success(t('aiConfig.defaultSet', 'Default profile updated'))
    } catch (error) {
      toast.error(t('aiConfig.saveError', 'Failed to update default profile'))
    }
  }, [loadData, t])

  const handleDelete = useCallback(async (id: number) => {
    if (!confirm(t('common.confirmDelete', 'Are you sure you want to delete this configuration?'))) return
    try {
      // We need a deleteAIConfig endpoint or use updateAIConfig with a flag
      // Let's assume we have it or add it
      await deleteAIConfig(id)
      await loadData()
      toast.success(t('aiConfig.deleted', 'Configuration removed'))
    } catch (error) {
      toast.error(t('aiConfig.deleteError', 'Failed to remove configuration'))
    }
  }, [loadData, t])

  if (loading) return <div>Loading AI Settings...</div>

  const selectedProfile = profiles.find((p) => p.id === editingConfig?.profileID)

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <IconRobot className="h-5 w-5" />
            {t('aiConfig.title', 'AI Assistant Profiles')}
          </CardTitle>
          {!editingConfig && (
            <Button size="sm" onClick={() => setEditingConfig({ profileID: profiles[0]?.id, isActive: true, isDefault: userConfigs.length === 0 })}>
              <IconPlus className="h-4 w-4 mr-2" />
              {t('aiConfig.add', 'Add Profile')}
            </Button>
          )}
        </CardHeader>
        <CardContent>
          {userConfigs.length === 0 ? (
            <div className="text-center py-6 text-muted-foreground">
              {t('aiConfig.noConfigs', 'No AI profiles configured. Add one to get started.')}
            </div>
          ) : (
            <div className="space-y-4">
              {userConfigs.map((cfg) => {
                const profile = profiles.find(p => p.id === cfg.profileID)
                return (
                  <div key={cfg.id} className="flex items-center justify-between p-3 border rounded-lg bg-card/50">
                    <div className="flex flex-col">
                      <div className="flex items-center gap-2">
                        <span className="font-medium">{profile?.name || 'Unknown Profile'}</span>
                        {cfg.isDefault && (
                          <Badge variant="secondary" className="gap-1 px-1.5 h-5">
                            <IconStarFilled className="h-3 w-3 text-yellow-500" />
                            {t('aiConfig.default', 'Default')}
                          </Badge>
                        )}
                      </div>
                      <span className="text-xs text-muted-foreground">
                        {profile?.provider} {cfg.modelOverride ? `• ${cfg.modelOverride}` : ''}
                      </span>
                    </div>
                    <div className="flex items-center gap-2">
                      {!cfg.isDefault && (
                        <Button variant="ghost" size="icon" title={t('aiConfig.setAsDefault', 'Set as Default')} onClick={() => handleSetDefault(cfg)}>
                          <IconStar className="h-4 w-4" />
                        </Button>
                      )}
                      <Button variant="ghost" size="icon" onClick={() => setEditingConfig(cfg)}>
                        {t('common.edit', 'Edit')}
                      </Button>
                      <Button variant="ghost" size="icon" className="text-destructive" onClick={() => cfg.id && handleDelete(cfg.id)}>
                        <IconTrash className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </CardContent>
      </Card>

      {editingConfig && (
        <Card>
          <CardHeader>
            <CardTitle>{editingConfig.id ? t('aiConfig.edit', 'Edit Configuration') : t('aiConfig.add', 'Add Configuration')}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="profile">{t('aiConfig.profile', 'AI Profile')}</Label>
              <Select
                value={String(editingConfig.profileID)}
                disabled={!!editingConfig.id}
                onValueChange={(v) =>
                  setEditingConfig({ ...editingConfig, profileID: parseInt(v), apiKey: '', modelOverride: '' })
                }
              >
                <SelectTrigger id="profile">
                  <SelectValue placeholder="Select an AI profile" />
                </SelectTrigger>
                <SelectContent>
                  {profiles.filter((p) => p.isEnabled).map((p) => (
                    <SelectItem key={p.id} value={String(p.id)}>
                      {p.name} ({p.provider})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {selectedProfile?.allowUserOverride && (
              <>
                <div className="space-y-2">
                  <Label htmlFor="apiKey">
                    {t('aiConfig.apiKey', 'Personal API Key')}
                  </Label>
                  <Input
                    id="apiKey"
                    type="password"
                    placeholder="Leave blank to use system default"
                    value={editingConfig.apiKey || ''}
                    onChange={(e) => setEditingConfig({ ...editingConfig, apiKey: e.target.value })}
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="model">
                    {t('aiConfig.model', 'Model Override')}
                  </Label>
                  {selectedProfile.allowedModels && selectedProfile.allowedModels.length > 0 ? (
                    <Select
                      value={editingConfig.modelOverride || 'default'}
                      onValueChange={(v) =>
                        setEditingConfig({ ...editingConfig, modelOverride: v === 'default' ? '' : v })
                      }
                    >
                      <SelectTrigger id="model">
                        <SelectValue placeholder="Select a model" />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="default">
                          {t('aiConfig.defaultModel', 'Default')}: {selectedProfile.defaultModel}
                        </SelectItem>
                        {selectedProfile.allowedModels.map((m) => (
                          <SelectItem key={m} value={m}>
                            {m}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  ) : (
                    <Input
                      id="model"
                      placeholder={`Default: ${selectedProfile.defaultModel}`}
                      value={editingConfig.modelOverride || ''}
                      onChange={(e) =>
                        setEditingConfig({ ...editingConfig, modelOverride: e.target.value })
                      }
                    />
                  )}
                </div>
              </>
            )}

            <div className="flex items-center space-x-2 py-2">
              <Switch
                id="isDefault"
                checked={editingConfig.isDefault || false}
                onCheckedChange={(checked) => setEditingConfig({ ...editingConfig, isDefault: checked })}
              />
              <Label htmlFor="isDefault">{t('aiConfig.setAsDefault', 'Set as Default profile')}</Label>
            </div>

            <div className="flex gap-2 pt-4">
              <Button onClick={handleSave} disabled={saving}>
                {saving ? t('common.saving', 'Saving...') : t('common.save', 'Save Changes')}
              </Button>
              <Button variant="ghost" onClick={() => setEditingConfig(null)}>
                {t('common.cancel', 'Cancel')}
              </Button>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
