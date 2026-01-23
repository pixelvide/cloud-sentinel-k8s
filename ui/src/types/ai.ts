export interface AIProviderProfile {
  id: number
  name: string
  provider: 'gemini' | 'openai' | 'azure' | 'custom'
  baseUrl: string
  defaultModel: string
  apiKey?: string
  isSystem: boolean
  isEnabled: boolean
  allowUserOverride: boolean
  allowedModels: string[]
  createdAt: string
  updatedAt: string
}

export interface AISettings {
  id?: number
  userID: number
  profileID: number
  apiKey: string
  modelOverride: string
  isActive: boolean
  isDefault?: boolean
  createdAt?: string
  updatedAt?: string
}

export interface AIChatSession {
  id: string
  userID: number
  title: string
  createdAt: string
  updatedAt: string
  messages?: AIChatMessage[]
}

export interface AIChatMessage {
  id?: number
  sessionID?: string
  role: 'system' | 'user' | 'assistant' | 'tool'
  content: string
  toolCalls?: string // JSON string
  toolID?: string
  createdAt: string
}

export interface ChatRequest {
  sessionID: string
  message: string
  model?: string
}

export interface ChatResponse {
  sessionID: string
  message: string
}
export interface AIModelsResponse {
  models: string[]
  default: string
  provider: string
  message?: string
}
