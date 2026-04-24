interface ChatMessage {
  role: 'user' | 'assistant'
  content: string
}

interface ChatResponse {
  message: string
  session_id: string
}

interface ChatbotStatusResponse {
  active: boolean
}

const pageMap: Record<string, string> = {
  'index': 'forecast',
  'employees': 'employees',
  'employees-id': 'employees',
  'transactions': 'transactions',
  'bank-accounts': 'bank-accounts',
  'settings': 'settings',
  'settings-profile': 'settings',
  'settings-organisations': 'settings',
  'settings-app': 'settings',
  'settings-automation': 'settings',
  'organisation': 'settings',
}

export default function useChat() {
  const messages = useState<ChatMessage[]>('chatMessages', () => [])
  const sessionId = useState<string | null>('chatSessionId', () => null)
  const isLoading = useState('chatIsLoading', () => false)
  const isOpen = useState('chatIsOpen', () => false)
  const chatbotActive = useState<boolean | null>('chatbotActive', () => null)

  const getPageName = (): string => {
    const route = useRoute()
    const name = String(route.name ?? '')
    return pageMap[name] ?? 'forecast'
  }

  const checkChatbotStatus = async () => {
    if (chatbotActive.value !== null) return
    try {
      const data = await $fetch<ChatbotStatusResponse>('/api/organisations/chatbots/status')
      chatbotActive.value = data.active
    }
    catch {
      chatbotActive.value = false
    }
  }

  const sendMessage = async (message: string) => {
    const page = getPageName()
    messages.value.push({ role: 'user', content: message })
    isLoading.value = true

    try {
      const data = await $fetch<ChatResponse>('/api/chat', {
        method: 'POST',
        body: {
          page,
          message,
          session_id: sessionId.value ?? '',
        },
      })
      sessionId.value = data.session_id
      messages.value.push({ role: 'assistant', content: data.message })
    }
    catch {
      messages.value.push({ role: 'assistant', content: 'Es ist ein Fehler aufgetreten. Bitte versuchen Sie es erneut.' })
    }
    finally {
      isLoading.value = false
    }
  }

  const clearChat = () => {
    messages.value = []
    sessionId.value = null
  }

  return {
    messages,
    sessionId,
    isLoading,
    isOpen,
    chatbotActive,
    checkChatbotStatus,
    sendMessage,
    clearChat,
    getPageName,
  }
}
