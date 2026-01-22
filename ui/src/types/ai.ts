export interface AISettings {
  userID: number
  provider: string
  model: string
  apiKey: string
  baseUrl: string
  createdAt: string
  updatedAt: string
}

export interface ChatSession {
  id: string
  userID: number
  title: string
  createdAt: string
  updatedAt: string
  messages?: ChatMessage[]
}

export interface ChatMessage {
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
}

export interface ChatResponse {
  sessionID: string
  message: string
}
