import { useEffect, useRef, useState } from 'react'
import { useLocation, matchPath, useNavigate } from 'react-router-dom'
import {
    IconRobot,
    IconSend,
    IconUser,
    IconX,
    IconMinimize,
    IconMaximize,
    IconHistory,
    IconChevronLeft,
    IconTrash,
    IconPlus,
    IconBulb,
    IconChevronDown,
    IconArrowsMaximize,
    IconArrowsMinimize,
} from '@tabler/icons-react'
import * as Collapsible from '@radix-ui/react-collapsible'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import { clsx } from 'clsx'
import { format } from 'date-fns'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { AIChatMessage, AIModelsResponse, AIChatSession } from '@/types/ai'
import {
    fetchAIModels,
    listAIChatSessions,
    getAIChatSession,
    deleteAIChatSession,
} from '@/lib/api'
import { getSubPath, withSubPath } from '@/lib/subpath'
import { useAuth } from '@/contexts/auth-context'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Textarea } from '@/components/ui/textarea'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'

export function FloatingAIChat() {
    const { t } = useTranslation()
    const { config } = useAuth()
    const [isOpen, setIsOpen] = useState(false)
    const [isMinimized, setIsMinimized] = useState(false)
    const [messages, setMessages] = useState<AIChatMessage[]>([])
    const [inputValue, setInputValue] = useState('')
    const [sending, setSending] = useState(false)
    const [sessionId, setSessionId] = useState<string | null>(null)
    const [selectedModel, setSelectedModel] = useState('')
    const [availableModels, setAvailableModels] = useState<string[]>([])
    const [configMessage, setConfigMessage] = useState<string | null>(null)
    const [showHistory, setShowHistory] = useState(false)
    const [sessions, setSessions] = useState<AIChatSession[]>([])
    const [loadingSessions, setLoadingSessions] = useState(false)
    const [isExpanded, setIsExpanded] = useState(false)

    const location = useLocation()
    const navigate = useNavigate()

    const messagesEndRef = useRef<HTMLDivElement>(null)

    useEffect(() => {
        const handleToggle = () => setIsOpen((prev) => !prev)
        window.addEventListener('toggle-ai-chat', handleToggle)

        // Fetch models
        const loadModels = async () => {
            try {
                const response: AIModelsResponse = await fetchAIModels()
                setAvailableModels(response.models)
                setSelectedModel(response.default || (response.models.length > 0 ? response.models[0] : ''))
                if (response.message) {
                    setConfigMessage(response.message)
                }
            } catch (error) {
                console.error('Failed to fetch AI models', error)
            }
        }
        loadModels()

        return () => window.removeEventListener('toggle-ai-chat', handleToggle)
    }, [])

    const scrollContainerRef = useRef<HTMLDivElement>(null)
    const shouldAutoScrollRef = useRef(true)

    // Track if user is at the bottom
    const handleScroll = () => {
        if (!scrollContainerRef.current) return
        const { scrollTop, scrollHeight, clientHeight } = scrollContainerRef.current
        // Tolerance of 20px
        const isAtBottom = scrollHeight - scrollTop - clientHeight < 50
        shouldAutoScrollRef.current = isAtBottom
    }

    // Initial scroll on open
    useEffect(() => {
        if (isOpen && !isMinimized && messagesEndRef.current) {
            messagesEndRef.current.scrollIntoView({ behavior: 'auto' })
            shouldAutoScrollRef.current = true
        }
    }, [isOpen, isMinimized])

    // Auto-scroll on new messages if sticky
    useEffect(() => {
        if (isOpen && !isMinimized && shouldAutoScrollRef.current && messagesEndRef.current) {
            messagesEndRef.current.scrollIntoView({ behavior: 'smooth' })
        }
    }, [messages, isOpen, isMinimized])

    const getContextFromUrl = () => {
        let pathname = location.pathname
        const subPath = getSubPath()

        // Strip subpath if present to match routes correctly
        if (subPath && subPath !== '/' && pathname.startsWith(subPath)) {
            pathname = pathname.substring(subPath.length)
        }

        // Match specific resource: /c/:cluster/:kind/:namespace/:name
        // e.g. /c/local/pods/default/nginx-123
        // Note: Our routes are sometimes :resource/:namespace/:name
        const matchNamespaced = matchPath('/c/:cluster/:kind/:namespace/:name', pathname)
        if (matchNamespaced) {
            return {
                route: pathname,
                kind: matchNamespaced.params.kind,
                namespace: matchNamespaced.params.namespace,
                name: matchNamespaced.params.name
            }
        }

        // Match simple resource: /c/:cluster/:kind/:name (e.g. nodes, namespaces)
        const matchSimple = matchPath('/c/:cluster/:kind/:name', pathname)
        if (matchSimple) {
            return {
                route: pathname,
                kind: matchSimple.params.kind,
                name: matchSimple.params.name
            }
        }

        // Match list: /c/:cluster/:kind
        const matchList = matchPath('/c/:cluster/:kind', pathname)
        if (matchList) {
            return {
                route: pathname,
                kind: matchList.params.kind
            }
        }

        // Match cluster root
        const matchCluster = matchPath('/c/:cluster/dashboard', pathname)
        if (matchCluster) {
            return {
                route: pathname,
                kind: 'Reviewing Cluster Dashboard'
            }
        }

        return { route: pathname }
    }

    const handleSend = async () => {
        if (!inputValue.trim() || sending) return

        const userMsg: AIChatMessage = {
            role: 'user',
            content: inputValue,
            createdAt: new Date().toISOString(),
        }

        setMessages((prev) => [...prev, userMsg])
        setInputValue('')
        setSending(true)
        shouldAutoScrollRef.current = true

        // Add an empty assistant message that we'll fill as we stream
        const initialAssistantMsg: AIChatMessage = {
            role: 'assistant',
            content: '',
            createdAt: new Date().toISOString(),
        }
        setMessages((prev) => [...prev, initialAssistantMsg])

        try {
            const clusterName = localStorage.getItem('current-cluster') || undefined

            const response = await fetch(withSubPath('/api/v1/ai/chat'), {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-Cluster': clusterName || '',
                },
                credentials: 'include',
                body: JSON.stringify({
                    sessionID: sessionId || '',
                    message: userMsg.content,
                    model: selectedModel,
                    context: getContextFromUrl(),
                }),
            })

            if (!response.ok) {
                throw new Error('Failed to send message')
            }

            const reader = response.body?.getReader()
            const decoder = new TextDecoder()
            let accumulatedContent = ''

            if (reader) {
                while (true) {
                    const { done, value } = await reader.read()
                    if (done) break

                    const chunk = decoder.decode(value, { stream: true })
                    const lines = chunk.split('\n')

                    for (const line of lines) {
                        if (line.startsWith('event:')) {
                            const eventType = line.replace('event:', '').trim()
                            // The data follows the event line
                            const dataLine = lines[lines.indexOf(line) + 1]
                            if (dataLine && dataLine.startsWith('data:')) {
                                try {
                                    const data = JSON.parse(dataLine.replace('data:', '').trim())

                                    if (eventType === 'session') {
                                        if (!sessionId) setSessionId(data.sessionID)
                                    } else if (eventType === 'message') {
                                        accumulatedContent += data.content
                                        setMessages((prev) => {
                                            const newMsgs = [...prev]
                                            newMsgs[newMsgs.length - 1] = {
                                                ...newMsgs[newMsgs.length - 1],
                                                content: accumulatedContent,
                                            }
                                            return newMsgs
                                        })
                                    } else if (eventType === 'status') {
                                        // Optional: handle status updates if needed
                                        console.log('AI Status:', data.status)
                                    } else if (eventType === 'error') {
                                        toast.error(data.error)
                                    }
                                } catch {
                                    // Ignore parse errors for incomplete JSON
                                }
                            }
                        }

                        // Check for tool calls in the line for navigation
                        if (line.includes('<tool_call>')) {
                            try {
                                // Simple extraction for now, might need robust parsing if split across lines
                                // format: <tool_call>JSON</tool_call>
                                const start = line.indexOf('<tool_call>') + 11
                                const end = line.indexOf('</tool_call>')
                                if (end > start) {
                                    const jsonStr = line.substring(start, end)
                                    const toolCall = JSON.parse(jsonStr)
                                    if (toolCall.name === 'navigate_to') {
                                        const args = JSON.parse(toolCall.arguments)
                                        if (args.path) {
                                            navigate(withSubPath(args.path))
                                        } else if (args.page) {
                                            const cluster = localStorage.getItem('current-cluster') || 'local'
                                            navigate(withSubPath(`/c/${cluster}/${args.page}`))
                                        }
                                    }
                                }
                            } catch {
                                // ignore parsing errors during stream
                            }
                        }
                    }
                }
            }
        } catch (error) {
            console.error(error)
            toast.error(t('aiChat.errors.send', 'Failed to send message'))
        } finally {
            setSending(false)
        }
    }

    const loadSessions = async () => {
        setLoadingSessions(true)
        try {
            const data = await listAIChatSessions()
            setSessions(data)
        } catch (error) {
            console.error('Failed to load sessions', error)
        } finally {
            setLoadingSessions(false)
        }
    }

    const handleSelectSession = async (session: AIChatSession) => {
        try {
            const fullSession = await getAIChatSession(session.id)
            setMessages(fullSession.messages || [])
            setSessionId(session.id)
            setShowHistory(false)
        } catch (error) {
            console.error('Failed to load session details', error)
            toast.error(t('aiChat.errors.loadSession', 'Failed to load session'))
        }
    }

    const handleDeleteSession = async (e: React.MouseEvent, id: string) => {
        e.stopPropagation()
        try {
            await deleteAIChatSession(id)
            setSessions((prev) => prev.filter((s) => s.id !== id))
            if (sessionId === id) {
                setSessionId(null)
                setMessages([])
            }
            toast.success(t('aiChat.sessionDeleted', 'Session deleted'))
        } catch (error) {
            console.error('Failed to delete session', error)
            toast.error(t('aiChat.errors.deleteSession', 'Failed to delete session'))
        }
    }

    const handleNewChat = () => {
        setSessionId(null)
        setMessages([])
        setShowHistory(false)
    }

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault()
            handleSend()
        }
    }

    if (!isOpen || !config?.is_ai_chat_enabled) return null

    return (
        <Card
            className={clsx(
                'fixed right-6 bottom-6 shadow-2xl z-[9999] flex flex-col transition-all duration-300 overflow-hidden border border-primary/20 bg-background p-0 gap-0',
                isMinimized
                    ? 'h-14 w-96'
                    : isExpanded
                        ? 'w-[800px] h-[800px] max-w-[calc(100vw-3rem)] max-h-[calc(100vh-3rem)]'
                        : 'w-96 h-[600px] max-h-[80vh]'
            )}
        >
            {/* Header */}
            <div className="p-3 bg-primary text-primary-foreground flex items-center justify-between cursor-pointer" onClick={() => setIsMinimized(!isMinimized)}>
                <div className="flex items-center gap-2">
                    {showHistory ? (
                        <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6 text-primary-foreground hover:bg-primary-foreground/10"
                            onClick={(e) => {
                                e.stopPropagation()
                                setShowHistory(false)
                            }}
                        >
                            <IconChevronLeft className="h-4 w-4" />
                        </Button>
                    ) : (
                        <IconRobot className="h-5 w-5" />
                    )}
                    <span className="font-semibold text-sm">
                        {showHistory ? t('aiChat.history', 'History') : 'Cloud Sentinel AI'}
                    </span>
                </div>
                <div className="flex items-center gap-1" onClick={(e) => e.stopPropagation()}>
                    {!showHistory && !isMinimized && (
                        <>
                            <Button
                                variant="ghost"
                                size="icon"
                                title={t('aiChat.newChat', 'New Chat')}
                                className="h-8 w-8 text-primary-foreground hover:bg-primary-foreground/10"
                                onClick={handleNewChat}
                            >
                                <IconPlus className="h-4 w-4" />
                            </Button>
                            <Button
                                variant="ghost"
                                size="icon"
                                title={t('aiChat.viewHistory', 'View History')}
                                className="h-8 w-8 text-primary-foreground hover:bg-primary-foreground/10"
                                onClick={() => {
                                    setShowHistory(true)
                                    loadSessions()
                                }}
                            >
                                <IconHistory className="h-4 w-4" />
                            </Button>
                            <Button
                                variant="ghost"
                                size="icon"
                                title={isExpanded ? t('aiChat.contract', 'Contract') : t('aiChat.expand', 'Expand')}
                                className="h-8 w-8 text-primary-foreground hover:bg-primary-foreground/10"
                                onClick={() => setIsExpanded(!isExpanded)}
                            >
                                {isExpanded ? (
                                    <IconArrowsMinimize className="h-4 w-4" />
                                ) : (
                                    <IconArrowsMaximize className="h-4 w-4" />
                                )}
                            </Button>
                        </>
                    )}
                    <Button variant="ghost" size="icon" className="h-8 w-8 text-primary-foreground hover:bg-primary-foreground/10" onClick={() => setIsMinimized(!isMinimized)}>
                        {isMinimized ? <IconMaximize className="h-4 w-4" /> : <IconMinimize className="h-4 w-4" />}
                    </Button>
                    <Button variant="ghost" size="icon" className="h-8 w-8 text-primary-foreground hover:bg-primary-foreground/10" onClick={() => setIsOpen(false)}>
                        <IconX className="h-4 w-4" />
                    </Button>
                </div>
            </div>

            {!isMinimized && showHistory && (
                <div className="flex-1 overflow-hidden flex flex-col">
                    <ScrollArea className="flex-1 p-2">
                        {loadingSessions ? (
                            <div className="flex justify-center p-10">
                                <span className="w-2 h-2 bg-primary rounded-full animate-ping" />
                            </div>
                        ) : sessions.length === 0 ? (
                            <div className="flex flex-col items-center justify-center p-10 text-muted-foreground">
                                <IconHistory className="h-10 w-10 mb-2 opacity-20" />
                                <p className="text-xs">{t('aiChat.noHistory', 'No history found')}</p>
                            </div>
                        ) : (
                            <div className="space-y-1">
                                {sessions.map((session) => (
                                    <div
                                        key={session.id}
                                        onClick={() => handleSelectSession(session)}
                                        className={clsx(
                                            "group flex items-center justify-between p-3 rounded-lg cursor-pointer transition-colors",
                                            sessionId === session.id ? "bg-primary/10 border border-primary/20" : "hover:bg-muted"
                                        )}
                                    >
                                        <div className="flex flex-col gap-0.5 overflow-hidden">
                                            <span className="text-sm font-medium truncate pr-4">
                                                {session.title || t('aiChat.newChat', 'New Chat')}
                                            </span>
                                            <span className="text-[10px] text-muted-foreground">
                                                {format(new Date(session.updatedAt), 'MMM d, yyyy HH:mm')}
                                            </span>
                                        </div>
                                        <Button
                                            variant="ghost"
                                            size="icon"
                                            className="h-7 w-7 opacity-0 group-hover:opacity-100 transition-opacity hover:text-destructive hover:bg-destructive/10"
                                            onClick={(e) => handleDeleteSession(e, session.id)}
                                        >
                                            <IconTrash className="h-4 w-4" />
                                        </Button>
                                    </div>
                                ))}
                            </div>
                        )}
                    </ScrollArea>
                </div>
            )}

            {!isMinimized && !showHistory && (
                <>
                    {/* Messages */}
                    <div className="flex-1 overflow-hidden">
                        <ScrollArea
                            className="h-full p-4"
                            ref={scrollContainerRef}
                            onScroll={handleScroll}
                        >
                            {messages.length === 0 && (
                                <div className="h-full flex flex-col items-center justify-center text-muted-foreground pt-10">
                                    <IconRobot className="h-10 w-10 mb-2 opacity-20" />
                                    <p className="text-xs text-center px-4">
                                        {t('aiChat.empty', 'Ask me anything about your Kubernetes clusters. I can list pods, check logs, and analyze security.')}
                                    </p>
                                </div>
                            )}

                            <div className="space-y-4 pb-2">
                                {messages.map((msg, idx) => (
                                    <div key={idx} className={clsx("flex gap-2", msg.role === 'user' ? 'justify-end' : 'justify-start')}>
                                        {msg.role !== 'user' && (
                                            <div className="h-7 w-7 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0 mt-1">
                                                <IconRobot className="h-4 w-4 text-primary" />
                                            </div>
                                        )}
                                        {msg.role === 'user' ? (
                                            <div className="bg-primary text-primary-foreground rounded-2xl rounded-br-none px-3 py-2 shadow-sm max-w-[80%]">
                                                <ReactMarkdown
                                                    remarkPlugins={[remarkGfm]}
                                                    components={{
                                                        p: ({ children }) => <p className="mb-2 last:mb-0">{children}</p>,
                                                        code: ({ children }) => (
                                                            <code className="bg-primary-foreground/20 px-1.5 py-0.5 rounded text-[12px] font-mono">{children}</code>
                                                        ),
                                                    }}
                                                >
                                                    {msg.content}
                                                </ReactMarkdown>
                                            </div>
                                        ) : (
                                            (() => {
                                                const content = msg.content
                                                const parts = []
                                                let remaining = content

                                                // Extract Plan
                                                const planMatch = remaining.match(/<plan>([\s\S]*?)<\/plan>/)
                                                if (planMatch) {
                                                    parts.push({ type: 'plan', content: planMatch[1].trim() })
                                                    remaining = remaining.replace(planMatch[0], '')
                                                }

                                                const regex = /(<thought>[\s\S]*?<\/thought>)|(<tool_call>[\s\S]*?<\/tool_call>)|(<tool_result>[\s\S]*?<\/tool_result>)/g
                                                let match
                                                let lastIndex = 0
                                                const tags = []

                                                while ((match = regex.exec(remaining)) !== null) {
                                                    if (match.index > lastIndex) {
                                                        const text = remaining.substring(lastIndex, match.index)
                                                        if (text.trim()) tags.push({ type: 'text', content: text })
                                                    }

                                                    if (match[1]) {
                                                        tags.push({ type: 'thought', content: match[1].replace(/<\/?thought>/g, '').trim() })
                                                    } else if (match[2]) {
                                                        tags.push({ type: 'tool_call', content: match[2].replace(/<\/?tool_call>/g, '').trim() })
                                                    } else if (match[3]) {
                                                        tags.push({ type: 'tool_result', content: match[3].replace(/<\/?tool_result>/g, '').trim() })
                                                    }
                                                    lastIndex = regex.lastIndex
                                                }
                                                if (lastIndex < remaining.length) {
                                                    const text = remaining.substring(lastIndex)
                                                    if (text.trim()) tags.push({ type: 'text', content: text })
                                                }

                                                return (
                                                    <div className="flex flex-col gap-2 max-w-[85%]">
                                                        {parts.map((p, i) => (
                                                            <div key={`plan-${i}`} className="bg-blue-50/50 dark:bg-blue-900/10 border border-blue-200 dark:border-blue-800 rounded-lg p-3 text-xs mb-2 shadow-sm">
                                                                <div className="font-semibold text-blue-700 dark:text-blue-400 mb-1 flex items-center gap-1">
                                                                    <IconBulb className="h-3 w-3" /> Plan
                                                                </div>
                                                                <div className="whitespace-pre-wrap text-blue-900 dark:text-blue-300 font-mono text-[11px] leading-relaxed">
                                                                    {p.content}
                                                                </div>
                                                            </div>
                                                        ))}

                                                        {tags.map((tag, i) => {
                                                            if (tag.type === 'thought') {
                                                                return (
                                                                    <Collapsible.Root
                                                                        key={i}
                                                                        defaultOpen={idx === messages.length - 1 && sending}
                                                                        className="bg-background/50 rounded-lg overflow-hidden border border-primary/10 mb-1"
                                                                    >
                                                                        <Collapsible.Trigger asChild>
                                                                            <button className="flex items-center justify-between w-full px-3 py-1.5 text-[11px] font-medium text-muted-foreground hover:bg-primary/5 transition-colors">
                                                                                <div className="flex items-center gap-1.5 line-clamp-1">
                                                                                    <IconBulb className="h-3 w-3 text-primary/60" />
                                                                                    <span>Thinking Process...</span>
                                                                                </div>
                                                                                <IconChevronDown className="h-3 w-3 transition-transform duration-200" />
                                                                            </button>
                                                                        </Collapsible.Trigger>
                                                                        <Collapsible.Content className="px-3 py-2 text-[11px] text-muted-foreground/80 border-t border-primary/5 italic whitespace-pre-wrap leading-relaxed">
                                                                            {tag.content}
                                                                        </Collapsible.Content>
                                                                    </Collapsible.Root>
                                                                )
                                                            } else if (tag.type === 'tool_call') {
                                                                let callData = { name: 'Unknown', arguments: {} }
                                                                try { callData = JSON.parse(tag.content) } catch {
                                                                    // ignore
                                                                }
                                                                return (
                                                                    <div key={i} className="bg-muted/30 rounded-lg border border-primary/10 overflow-hidden mb-1">
                                                                        <div className="px-3 py-1.5 text-[11px] font-medium flex items-center justify-between text-muted-foreground bg-muted/50">
                                                                            <span className="font-mono">Running: {callData.name}</span>
                                                                        </div>
                                                                        <div className="px-3 py-2 text-[10px] font-mono text-muted-foreground overflow-x-auto">
                                                                            {JSON.stringify(callData.arguments)}
                                                                        </div>
                                                                    </div>
                                                                )
                                                            } else if (tag.type === 'tool_result') {
                                                                return (
                                                                    <Collapsible.Root
                                                                        key={i}
                                                                        defaultOpen={idx === messages.length - 1 && sending}
                                                                        className="bg-muted/30 rounded-lg border border-primary/10 overflow-hidden mb-1"
                                                                    >
                                                                        <Collapsible.Trigger asChild>
                                                                            <button className="flex items-center justify-between w-full px-3 py-1.5 text-[11px] font-medium text-muted-foreground hover:bg-primary/5 transition-colors">
                                                                                <div className="flex items-center gap-1.5">
                                                                                    <span>Tool Output</span>
                                                                                </div>
                                                                                <IconChevronDown className="h-3 w-3 transition-transform duration-200" />
                                                                            </button>
                                                                        </Collapsible.Trigger>
                                                                        <Collapsible.Content className="px-3 py-2 text-[10px] font-mono whitespace-pre-wrap max-h-[200px] overflow-y-auto border-t border-primary/5">
                                                                            {tag.content}
                                                                        </Collapsible.Content>
                                                                    </Collapsible.Root>
                                                                )
                                                            } else {
                                                                return (
                                                                    <div key={i} className="markdown-content prose prose-sm dark:prose-invert max-w-none bg-muted px-3 py-2 rounded-2xl rounded-tl-none shadow-sm">
                                                                        <ReactMarkdown
                                                                            remarkPlugins={[remarkGfm]}
                                                                            components={{
                                                                                table: ({ children }) => (
                                                                                    <div className="overflow-x-auto my-2 rounded-lg border border-primary/10">
                                                                                        <table className="w-full text-left border-collapse">{children}</table>
                                                                                    </div>
                                                                                ),
                                                                                thead: ({ children }) => (
                                                                                    <thead className="bg-muted/50 font-semibold">{children}</thead>
                                                                                ),
                                                                                th: ({ children }) => (
                                                                                    <th className="px-3 py-2 border-b border-primary/10">{children}</th>
                                                                                ),
                                                                                td: ({ children }) => (
                                                                                    <td className="px-3 py-2 border-b border-primary/5">{children}</td>
                                                                                ),
                                                                                p: ({ children }) => <p className="mb-2 last:mb-0">{children}</p>,
                                                                                code: ({ children }) => (
                                                                                    <code className="bg-muted px-1.5 py-0.5 rounded text-[12px] font-mono">{children}</code>
                                                                                ),
                                                                            }}
                                                                        >
                                                                            {tag.content}
                                                                        </ReactMarkdown>
                                                                    </div>
                                                                )
                                                            }
                                                        })}
                                                    </div>
                                                )
                                            })()
                                        )}
                                        {msg.role === 'user' && (
                                            <div className="h-7 w-7 rounded-full bg-primary flex items-center justify-center flex-shrink-0 mt-1">
                                                <IconUser className="h-4 w-4 text-primary-foreground" />
                                            </div>
                                        )}
                                    </div>
                                ))}

                                {sending && (
                                    <div className="flex gap-2 justify-start">
                                        <div className="h-7 w-7 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0">
                                            <IconRobot className="h-4 w-4 text-primary" />
                                        </div>
                                        <div className="bg-muted rounded-2xl rounded-tl-none px-3 py-2 shadow-sm">
                                            <div className="flex gap-1">
                                                <span className="w-1.5 h-1.5 bg-primary/40 rounded-full animate-bounce" />
                                                <span className="w-1.5 h-1.5 bg-primary/40 rounded-full animate-bounce [animation-delay:0.2s]" />
                                                <span className="w-1.5 h-1.5 bg-primary/40 rounded-full animate-bounce [animation-delay:0.4s]" />
                                            </div>
                                        </div>
                                    </div>
                                )}
                                <div ref={messagesEndRef} />
                            </div>
                        </ScrollArea>
                    </div>

                    {/* Input */}
                    <div className="p-3 border-t bg-background">
                        <div className="relative">
                            <Textarea
                                value={inputValue}
                                onChange={(e) => setInputValue(e.target.value)}
                                onKeyDown={handleKeyDown}
                                placeholder={configMessage || t('aiChat.placeholder', 'Type a message...')}
                                disabled={availableModels.length === 0}
                                className="min-h-[100px] w-full pr-12 pb-10 resize-none text-xs rounded-xl focus-visible:ring-primary/30"
                            />
                            <div className="absolute left-2 bottom-2 flex items-center gap-2">
                                <Select value={selectedModel} onValueChange={setSelectedModel}>
                                    <SelectTrigger className="h-7 w-fit text-[10px] bg-muted/50 border-none shadow-none focus:ring-0">
                                        <SelectValue placeholder="Model" />
                                    </SelectTrigger>
                                    <SelectContent className="z-[10000]">
                                        {availableModels.map((model) => (
                                            <SelectItem key={model} value={model}>
                                                {model}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </div>
                            <Button
                                size="icon"
                                onClick={handleSend}
                                disabled={!inputValue.trim() || sending}
                                className="absolute right-2 bottom-2 h-8 w-8 rounded-lg"
                            >
                                <IconSend className="h-4 w-4" />
                            </Button>
                        </div>
                        <div className="text-[10px] text-muted-foreground mt-2 flex justify-between items-center px-1">
                            <span>Kubernetes Assistant</span>
                            <div className="flex gap-2">
                                <span className="hover:text-primary cursor-pointer transition-colors" onClick={() => setInputValue("List pods")}>Pods</span>
                                <span className="hover:text-primary cursor-pointer transition-colors" onClick={() => setInputValue("Security scan")}>Scan</span>
                            </div>
                        </div>
                    </div>
                </>
            )}
        </Card>
    )
}
