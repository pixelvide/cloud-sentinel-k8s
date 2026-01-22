import { useCallback, useEffect, useState } from 'react'
import { IconCheck, IconRobot } from '@tabler/icons-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { getAIConfig, updateAIConfig } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

export function AIConfigManagement() {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [config, setConfig] = useState({
    provider: 'openai',
    model: 'gpt-3.5-turbo',
    apiKey: '',
    baseUrl: '',
  })

  useEffect(() => {
    getAIConfig()
      .then((data) => {
        if (data && data.provider) {
          setConfig({
            provider: data.provider || 'openai',
            model: data.model || 'gpt-3.5-turbo',
            apiKey: data.apiKey || '',
            baseUrl: data.baseUrl || '',
          })
        }
      })
      .catch((err) => {
        console.error(err)
      })
      .finally(() => setLoading(false))
  }, [])

  const handleSave = useCallback(async () => {
    setSaving(true)
    try {
      await updateAIConfig(config)
      toast.success(t('aiConfig.saved', 'AI configuration saved successfully'))
    } catch (error) {
      toast.error(t('aiConfig.saveError', 'Failed to save configuration'))
      console.error(error)
    } finally {
      setSaving(false)
    }
  }, [config, t])

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <IconRobot className="h-5 w-5" />
          {t('aiConfig.title', 'AI Assistant Configuration')}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="provider">{t('aiConfig.provider', 'Provider')}</Label>
          <Select
            value={config.provider}
            onValueChange={(v) => setConfig({ ...config, provider: v })}
          >
            <SelectTrigger id="provider">
              <SelectValue placeholder="Select provider" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="openai">OpenAI</SelectItem>
              <SelectItem value="google">Google Gemini</SelectItem>
              <SelectItem value="azure">Azure OpenAI</SelectItem>
              <SelectItem value="custom">Custom (LocalAI/vLLM)</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-2">
          <Label htmlFor="model">{t('aiConfig.model', 'Model Name')}</Label>
          <Input
            id="model"
            placeholder={
              config.provider === 'google'
                ? 'e.g. gemini-1.5-flash'
                : 'e.g. gpt-4, gpt-3.5-turbo'
            }
            value={config.model}
            onChange={(e) => setConfig({ ...config, model: e.target.value })}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="apiKey">{t('aiConfig.apiKey', 'API Key')}</Label>
          <Input
            id="apiKey"
            type="password"
            placeholder="sk-..."
            value={config.apiKey}
            onChange={(e) => setConfig({ ...config, apiKey: e.target.value })}
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="baseUrl">
            {t('aiConfig.baseUrl', 'Base URL (Optional)')}
          </Label>
          <Input
            id="baseUrl"
            placeholder="https://api.openai.com/v1"
            value={config.baseUrl}
            onChange={(e) => setConfig({ ...config, baseUrl: e.target.value })}
          />
          <p className="text-sm text-muted-foreground">
            {t(
              'aiConfig.baseUrlHelp',
              'Required for Azure or LocalAI. Leave empty for standard OpenAI.'
            )}
          </p>
        </div>

        <div className="pt-4">
          <Button onClick={handleSave} disabled={saving || loading}>
            {saving && <IconCheck className="mr-2 h-4 w-4 animate-spin" />}
            {t('common.save', 'Save Changes')}
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}
