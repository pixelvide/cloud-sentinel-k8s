import { useCallback, useEffect, useState } from 'react'
import { IconPlus, IconRobot, IconTrash } from '@tabler/icons-react'
import { toast } from 'sonner'

import {
    createAIProfile,
    deleteAIProfile,
    fetchAdminAIConfig,
    fetchAIProfiles,
    toggleAIProfile,
    updateAIGovernance,
    updateAIProfile,
} from '@/lib/api'
import { AIProviderProfile } from '@/types/ai'
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
import { Switch } from '@/components/ui/switch'

export function AIAdminManagement() {
    const [loading, setLoading] = useState(true)
    const [profiles, setProfiles] = useState<AIProviderProfile[]>([])
    const [adminConfig, setAdminConfig] = useState<{
        allow_user_keys: string
        force_user_keys: string
        system_profile: AIProviderProfile | null
    } | null>(null)

    const loadData = useCallback(async () => {
        setLoading(true)
        try {
            const [p, c] = await Promise.all([fetchAIProfiles(), fetchAdminAIConfig()])
            setProfiles(p)
            setAdminConfig(c)
        } catch (err) {
            console.error(err)
            toast.error('Failed to load AI administration data')
        } finally {
            setLoading(false)
        }
    }, [])

    useEffect(() => {
        loadData()
    }, [loadData])

    const handleSaveGovernance = async () => {
        if (!adminConfig) return
        try {
            await updateAIGovernance({
                allow_user_keys: adminConfig.allow_user_keys,
                force_user_keys: adminConfig.force_user_keys,
            })
            toast.success('AI Governance updated')
        } catch (err) {
            toast.error('Failed to update AI governance')
        }
    }

    const handleCreateProfile = async () => {
        try {
            await createAIProfile({
                name: 'New Profile',
                provider: 'gemini',
                defaultModel: 'gemini-1.5-flash',
            })
            loadData()
            toast.success('Profile created')
        } catch (err) {
            toast.error('Failed to create profile')
        }
    }

    const handleUpdateProfile = async (profile: AIProviderProfile) => {
        try {
            await updateAIProfile(profile.id, profile)
            toast.success('Profile updated')
        } catch (err) {
            toast.error('Failed to update profile')
        }
    }

    const handleDeleteProfile = async (id: number) => {
        if (!confirm('Are you sure you want to delete this profile?')) return
        try {
            await deleteAIProfile(id)
            loadData()
            toast.success('Profile deleted')
        } catch (err) {
            toast.error('Failed to delete profile')
        }
    }

    const handleToggleProfile = async (id: number) => {
        try {
            const updated = await toggleAIProfile(id)
            setProfiles((prev) =>
                prev.map((p) => (p.id === id ? updated : p))
            )
            toast.success(updated.isEnabled ? 'Profile enabled' : 'Profile disabled')
        } catch (err) {
            toast.error('Failed to toggle profile')
        }
    }

    if (loading) return <div>Loading Admin AI Settings...</div>

    return (
        <div className="space-y-6">
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <IconRobot className="h-5 w-5" />
                        Global AI Governance
                    </CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="flex items-center justify-between">
                        <div className="space-y-0.5">
                            <Label>Allow User API Keys (BYOK)</Label>
                            <p className="text-sm text-muted-foreground">
                                Users can provide their own API keys to override system defaults.
                            </p>
                        </div>
                        <Switch
                            checked={adminConfig?.allow_user_keys === 'true'}
                            onCheckedChange={(v) =>
                                setAdminConfig((prev) =>
                                    prev ? { ...prev, allow_user_keys: v ? 'true' : 'false' } : null
                                )
                            }
                        />
                    </div>

                    <div className="flex items-center justify-between">
                        <div className="space-y-0.5">
                            <Label>Force User API Keys</Label>
                            <p className="text-sm text-muted-foreground">
                                Users MUST provide their own API keys; system settings will not be used.
                            </p>
                        </div>
                        <Switch
                            checked={adminConfig?.force_user_keys === 'true'}
                            onCheckedChange={(v) =>
                                setAdminConfig((prev) =>
                                    prev ? { ...prev, force_user_keys: v ? 'true' : 'false' } : null
                                )
                            }
                        />
                    </div>

                    <Button onClick={handleSaveGovernance}>Save Governance Rules</Button>
                </CardContent>
            </Card>

            <Card>
                <CardHeader className="flex flex-row items-center justify-between">
                    <CardTitle>AI Provider Profiles</CardTitle>
                    <Button size="sm" onClick={handleCreateProfile}>
                        <IconPlus className="mr-2 h-4 w-4" /> Add Profile
                    </Button>
                </CardHeader>
                <CardContent className="space-y-6">
                    {profiles.map((profile) => (
                        <div key={profile.id} className={`space-y-4 border p-4 rounded-lg ${!profile.isEnabled ? 'opacity-60 bg-muted/30' : ''}`}>
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label>Profile Name</Label>
                                    <Input
                                        value={profile.name}
                                        onChange={(e) => {
                                            const newProfiles = profiles.map((p) =>
                                                p.id === profile.id ? { ...p, name: e.target.value } : p
                                            )
                                            setProfiles(newProfiles)
                                        }}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label>Provider Type</Label>
                                    <Select
                                        value={profile.provider}
                                        onValueChange={(v: any) => {
                                            const newProfiles = profiles.map((p) =>
                                                p.id === profile.id ? { ...p, provider: v } : p
                                            )
                                            setProfiles(newProfiles)
                                        }}
                                    >
                                        <SelectTrigger>
                                            <SelectValue />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectItem value="gemini">Gemini</SelectItem>
                                            <SelectItem value="openai">OpenAI</SelectItem>
                                            <SelectItem value="azure">Azure</SelectItem>
                                            <SelectItem value="custom">Custom</SelectItem>
                                        </SelectContent>
                                    </Select>
                                </div>
                                <div className="space-y-2">
                                    <Label>Base URL</Label>
                                    <Input
                                        value={profile.baseUrl}
                                        onChange={(e) => {
                                            const newProfiles = profiles.map((p) =>
                                                p.id === profile.id ? { ...p, baseUrl: e.target.value } : p
                                            )
                                            setProfiles(newProfiles)
                                        }}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label>Default Model</Label>
                                    <Input
                                        value={profile.defaultModel}
                                        onChange={(e) => {
                                            const newProfiles = profiles.map((p) =>
                                                p.id === profile.id
                                                    ? { ...p, defaultModel: e.target.value }
                                                    : p
                                            )
                                            setProfiles(newProfiles)
                                        }}
                                    />
                                </div>
                                <div className="space-y-2">
                                    <Label>System API Key</Label>
                                    <Input
                                        type="password"
                                        value={profile.apiKey || ''}
                                        placeholder="Secrets are not visible after saving"
                                        onChange={(e) => {
                                            const newProfiles = profiles.map((p) =>
                                                p.id === profile.id
                                                    ? { ...p, apiKey: e.target.value }
                                                    : p
                                            )
                                            setProfiles(newProfiles)
                                        }}
                                    />
                                </div>
                                <div className="space-y-2 col-span-2">
                                    <Label>Allowed Models (Comma separated, empty for any)</Label>
                                    <Input
                                        value={profile.allowedModels?.join(', ') || ''}
                                        placeholder="e.g. gpt-4, gpt-3.5-turbo"
                                        onChange={(e) => {
                                            // Store raw value temporarily to allow typing commas and spaces
                                            const rawValue = e.target.value
                                            const newProfiles = profiles.map((p) =>
                                                p.id === profile.id
                                                    ? { ...p, allowedModels: rawValue ? [rawValue] : [] }
                                                    : p
                                            )
                                            setProfiles(newProfiles)
                                        }}
                                        onBlur={(e) => {
                                            // Process the comma-separated list when user leaves the field
                                            const models = e.target.value.split(',').map(m => m.trim()).filter(m => m !== '')
                                            const newProfiles = profiles.map((p) =>
                                                p.id === profile.id
                                                    ? { ...p, allowedModels: models.length > 0 ? models : [] }
                                                    : p
                                            )
                                            setProfiles(newProfiles)
                                        }}
                                    />
                                </div>
                            </div>
                            <div className="flex items-center justify-between">
                                <div className="flex items-center gap-6">
                                    <div className="flex items-center gap-2">
                                        <Switch
                                            checked={profile.isEnabled}
                                            onCheckedChange={() => handleToggleProfile(profile.id)}
                                        />
                                        <Label className={profile.isEnabled ? 'text-green-600 font-semibold' : 'text-muted-foreground'}>
                                            {profile.isEnabled ? 'Enabled' : 'Disabled'}
                                        </Label>
                                    </div>
                                    <div className="flex items-center gap-2">
                                        <Switch
                                            checked={profile.allowUserOverride}
                                            onCheckedChange={(v) => {
                                                const newProfiles = profiles.map((p) =>
                                                    p.id === profile.id ? { ...p, allowUserOverride: v } : p
                                                )
                                                setProfiles(newProfiles)
                                            }}
                                        />
                                        <Label>Allow User Overrides</Label>
                                    </div>
                                    <div className="flex items-center gap-2">
                                        <Switch
                                            checked={profile.isSystem}
                                            onCheckedChange={(v) => {
                                                // Unset isSystem for others in local state for better UX
                                                const newProfiles = profiles.map((p) => {
                                                    if (p.id === profile.id) return { ...p, isSystem: v }
                                                    if (v) return { ...p, isSystem: false }
                                                    return p
                                                })
                                                setProfiles(newProfiles)
                                            }}
                                        />
                                        <Label className="text-primary font-semibold">
                                            System Default
                                        </Label>
                                    </div>
                                </div>
                                <div className="flex gap-2">
                                    <Button
                                        variant="destructive"
                                        size="sm"
                                        onClick={() => handleDeleteProfile(profile.id)}
                                    >
                                        <IconTrash className="h-4 w-4" />
                                    </Button>
                                    <Button size="sm" onClick={() => handleUpdateProfile(profile)}>
                                        Update Profile
                                    </Button>
                                </div>
                            </div>
                        </div>
                    ))}
                </CardContent>
            </Card>

        </div>
    )
}
