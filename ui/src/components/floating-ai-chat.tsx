import { useEffect, useRef, useState } from 'react'
import {
    IconRobot,
    IconSend,
    IconUser,
    IconX,
    IconMinimize,
    IconMaximize,
} from '@tabler/icons-react'
import { clsx } from 'clsx'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { AIChatMessage, AIModelsResponse } from '@/types/ai'
import { sendAIChatMessage, fetchAIModels } from '@/lib/api'
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

    useEffect(() => {
        if (isOpen && !isMinimized) {
            messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
        }
    }, [messages, isOpen, isMinimized])

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

        try {
            const clusterName = localStorage.getItem('current-cluster') || undefined
            const response = await sendAIChatMessage(
                {
                    sessionID: sessionId || '',
                    message: userMsg.content,
                    model: selectedModel,
                },
                clusterName
            )

            if (!sessionId) {
                setSessionId(response.sessionID)
            }

            const assistantMsg: AIChatMessage = {
                role: 'assistant',
                content: response.message,
                createdAt: new Date().toISOString(),
            }
            setMessages((prev) => [...prev, assistantMsg])
        } catch (error) {
            console.error(error)
            toast.error(t('aiChat.errors.send', 'Failed to send message'))
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

    if (!isOpen || !config?.is_ai_chat_enabled) return null

    return (
        <Card
            className={clsx(
                'fixed right-6 bottom-6 w-96 shadow-2xl z-[9999] flex flex-col transition-all duration-300 overflow-hidden border border-primary/20 bg-background p-0 gap-0',
                isMinimized ? 'h-14' : 'h-[600px] max-h-[80vh]'
            )}
        >
            {/* Header */}
            <div className="p-3 bg-primary text-primary-foreground flex items-center justify-between cursor-pointer" onClick={() => setIsMinimized(!isMinimized)}>
                <div className="flex items-center gap-2">
                    <IconRobot className="h-5 w-5" />
                    <span className="font-semibold text-sm">Cloud Sentinel AI</span>
                </div>
                <div className="flex items-center gap-1" onClick={(e) => e.stopPropagation()}>
                    <Button variant="ghost" size="icon" className="h-8 w-8 text-primary-foreground hover:bg-primary-foreground/10" onClick={() => setIsMinimized(!isMinimized)}>
                        {isMinimized ? <IconMaximize className="h-4 w-4" /> : <IconMinimize className="h-4 w-4" />}
                    </Button>
                    <Button variant="ghost" size="icon" className="h-8 w-8 text-primary-foreground hover:bg-primary-foreground/10" onClick={() => setIsOpen(false)}>
                        <IconX className="h-4 w-4" />
                    </Button>
                </div>
            </div>

            {!isMinimized && (
                <>
                    {/* Messages */}
                    <div className="flex-1 overflow-hidden">
                        <ScrollArea className="h-full p-4">
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
                                    <div
                                        key={idx}
                                        className={clsx(
                                            'flex gap-2',
                                            msg.role === 'user' ? 'justify-end' : 'justify-start'
                                        )}
                                    >
                                        {msg.role !== 'user' && (
                                            <div className="h-7 w-7 rounded-full bg-primary/10 flex items-center justify-center flex-shrink-0 mt-1">
                                                <IconRobot className="h-4 w-4 text-primary" />
                                            </div>
                                        )}

                                        <div
                                            className={clsx(
                                                'rounded-2xl px-3 py-2 max-w-[85%] text-sm shadow-sm',
                                                msg.role === 'user'
                                                    ? 'bg-primary text-primary-foreground rounded-tr-none'
                                                    : 'bg-muted rounded-tl-none'
                                            )}
                                        >
                                            {msg.content}
                                        </div>

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
