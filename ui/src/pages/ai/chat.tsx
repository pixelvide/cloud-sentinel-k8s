import { useCallback, useEffect, useRef, useState } from 'react'
import {
  IconCpu,
  IconMessage,
  IconPlus,
  IconRobot,
  IconSend,
  IconTrash,
  IconUser,
} from '@tabler/icons-react'
import { clsx } from 'clsx'
import { useTranslation } from 'react-i18next'
import { useNavigate, useParams } from 'react-router-dom'
import { toast } from 'sonner'

import { ChatMessage, ChatSession } from '@/types/ai'
import {
  deleteChatSession,
  getChatSession,
  listChatSessions,
  sendChatMessage,
} from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Textarea } from '@/components/ui/textarea'

export function AIChatPage() {
  const { t } = useTranslation()
  const { id } = useParams()
  const navigate = useNavigate()

  const [sessions, setSessions] = useState<ChatSession[]>([])
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [inputValue, setInputValue] = useState('')
  const [loading, setLoading] = useState(false)
  const [sending, setSending] = useState(false)

  const messagesEndRef = useRef<HTMLDivElement>(null)

  const fetchSessions = useCallback(async () => {
    try {
      const data = await listChatSessions()
      setSessions(data)
    } catch (error) {
      console.error(error)
      toast.error(t('aiChat.errors.loadSessions', 'Failed to load sessions'))
    }
  }, [t])

  useEffect(() => {
    fetchSessions()
  }, [fetchSessions])

  useEffect(() => {
    if (id) {
      setLoading(true)
      getChatSession(id)
        .then((session) => {
          setMessages(session.messages || [])
        })
        .catch((err) => {
          console.error(err)
          toast.error(t('aiChat.errors.loadSession', 'Failed to load session'))
          navigate('/ai/chat')
        })
        .finally(() => setLoading(false))
    } else {
      setMessages([])
    }
  }, [id, navigate, t])

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const handleNewChat = () => {
    navigate('/ai/chat')
  }

  const handleDeleteSession = async (
    e: React.MouseEvent,
    sessionId: string
  ) => {
    e.stopPropagation()
    if (!confirm(t('common.confirmDelete', 'Are you sure?'))) return

    try {
      await deleteChatSession(sessionId)
      setSessions((prev) => prev.filter((s) => s.id !== sessionId))
      if (id === sessionId) {
        navigate('/ai/chat')
      }
      toast.success(t('common.deleted', 'Deleted'))
    } catch (error) {
      console.error(error)
      toast.error('Failed to delete')
    }
  }

  const handleSend = async () => {
    if (!inputValue.trim() || sending) return

    const userMsg: ChatMessage = {
      role: 'user',
      content: inputValue,
      createdAt: new Date().toISOString(),
    }

    setMessages((prev) => [...prev, userMsg])
    setInputValue('')
    setSending(true)

    try {
      const clusterName = localStorage.getItem('current-cluster') || undefined
      const response = await sendChatMessage(
        {
          sessionID: id || '',
          message: userMsg.content,
        },
        clusterName
      )

      // If it was a new session, we need to navigate to the new ID
      if (!id) {
        await fetchSessions() // Refresh list to show new title
        navigate(`/ai/chat/${response.sessionID}`)
        return // The useEffect will load the messages
      }

      // Append assistant response
      const assistantMsg: ChatMessage = {
        role: 'assistant',
        content: response.message,
        createdAt: new Date().toISOString(),
      }
      setMessages((prev) => [...prev, assistantMsg])

      // Also refresh session list to update timestamp/title
      fetchSessions()
    } catch (error) {
      console.error(error)
      toast.error(t('aiChat.errors.send', 'Failed to send message'))
      // Remove user message? Or show error state?
    } finally {
      setSending(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  return (
    <div className="flex h-[calc(100vh-6rem)] gap-4">
      {/* Sidebar */}
      <Card className="w-64 flex flex-col overflow-hidden">
        <div className="p-4 border-b">
          <Button className="w-full justify-start" onClick={handleNewChat}>
            <IconPlus className="mr-2 h-4 w-4" />
            {t('aiChat.newChat', 'New Chat')}
          </Button>
        </div>
        <div className="flex-1 overflow-y-auto p-2 space-y-1">
          {sessions.map((session) => (
            <div
              key={session.id}
              className={clsx(
                'flex items-center justify-between p-2 rounded-md cursor-pointer hover:bg-accent text-sm group',
                id === session.id && 'bg-accent'
              )}
              onClick={() => navigate(`/ai/chat/${session.id}`)}
            >
              <div className="flex items-center overflow-hidden">
                <IconMessage className="mr-2 h-4 w-4 flex-shrink-0 text-muted-foreground" />
                <span className="truncate">{session.title || 'Untitled'}</span>
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="h-6 w-6 opacity-0 group-hover:opacity-100"
                onClick={(e) => handleDeleteSession(e, session.id)}
              >
                <IconTrash className="h-3 w-3 text-destructive" />
              </Button>
            </div>
          ))}
        </div>
      </Card>

      {/* Chat Area */}
      <Card className="flex-1 flex flex-col overflow-hidden">
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {messages.length === 0 && !loading && (
            <div className="h-full flex flex-col items-center justify-center text-muted-foreground">
              <IconRobot className="h-12 w-12 mb-4 opacity-20" />
              <p>
                {t(
                  'aiChat.empty',
                  'Start a conversation with the AI Assistant'
                )}
              </p>
            </div>
          )}

          {messages.map((msg, idx) => (
            <div
              key={idx}
              className={clsx(
                'flex gap-3',
                msg.role === 'user' ? 'justify-end' : 'justify-start'
              )}
            >
              {msg.role !== 'user' && (
                <div className="h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0">
                  {msg.role === 'assistant' ? (
                    <IconRobot className="h-5 w-5" />
                  ) : (
                    <IconCpu className="h-5 w-5" />
                  )}
                </div>
              )}

              <div
                className={clsx(
                  'rounded-lg p-3 max-w-[80%] whitespace-pre-wrap',
                  msg.role === 'user'
                    ? 'bg-primary text-primary-foreground'
                    : msg.role === 'tool'
                      ? 'bg-muted font-mono text-xs'
                      : 'bg-accent'
                )}
              >
                {msg.role === 'tool' && (
                  <div className="text-xs opacity-70 mb-1">Tool Output</div>
                )}
                {msg.content}
              </div>

              {msg.role === 'user' && (
                <div className="h-8 w-8 rounded-full bg-primary flex items-center justify-center flex-shrink-0">
                  <IconUser className="h-5 w-5 text-primary-foreground" />
                </div>
              )}
            </div>
          ))}

          {sending && (
            <div className="flex gap-3 justify-start">
              <div className="h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0">
                <IconRobot className="h-5 w-5" />
              </div>
              <div className="bg-accent rounded-lg p-3">
                <span className="animate-pulse">Thinking...</span>
              </div>
            </div>
          )}
          <div ref={messagesEndRef} />
        </div>

        <div className="p-4 border-t bg-background">
          <div className="flex gap-2">
            <Textarea
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder={t('aiChat.placeholder', 'Ask about your cluster...')}
              className="min-h-[60px]"
            />
            <Button
              onClick={handleSend}
              disabled={!inputValue.trim() || sending}
              className="h-auto"
            >
              <IconSend className="h-4 w-4" />
            </Button>
          </div>
          <div className="text-xs text-muted-foreground mt-2">
            Tip: Ask to &quot;list pods&quot;, &quot;check logs for pod X&quot;,
            or &quot;scale deployment Y&quot;.
          </div>
        </div>
      </Card>
    </div>
  )
}
