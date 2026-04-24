<template>
  <!-- FAB Button -->
  <button
    v-if="chatbotActive && !isOpen"
    class="fixed bottom-6 right-6 sm:bottom-6 sm:right-6 bottom-4 right-4
           z-50 w-14 h-14 rounded-full bg-primary text-white
           flex items-center justify-center shadow-lg
           hover:bg-primary/90 transition-all duration-200
           active:scale-95"
    @click="isOpen = true"
  >
    <i class="pi pi-bolt text-xl" />
  </button>

  <!-- Chat Panel -->
  <Transition
    enter-active-class="transition duration-200 ease-out"
    enter-from-class="opacity-0 scale-95"
    enter-to-class="opacity-100 scale-100"
    leave-active-class="transition duration-150 ease-in"
    leave-from-class="opacity-100 scale-100"
    leave-to-class="opacity-0 scale-95"
  >
    <div
      v-if="chatbotActive && isOpen"
      class="fixed z-50 flex flex-col bg-white dark:bg-zinc-900 shadow-2xl
             inset-0 h-[100dvh]
             sm:inset-auto sm:bottom-6 sm:right-6 sm:w-96 sm:max-h-[70vh] sm:h-[600px] sm:rounded-2xl sm:border sm:border-zinc-200 sm:dark:border-zinc-700"
    >
      <!-- Header -->
      <div class="flex items-center justify-between px-4 py-3 border-b border-zinc-200 dark:border-zinc-700 shrink-0">
        <div class="flex items-center gap-2">
          <i class="pi pi-bolt text-primary" />
          <span class="font-semibold text-sm">KI-Assistent</span>
        </div>
        <div class="flex items-center gap-1">
          <button
            v-if="messages.length > 0"
            class="p-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors text-zinc-500"
            title="Chat leeren"
            @click="clearChat"
          >
            <i class="pi pi-trash text-sm" />
          </button>
          <button
            class="p-2 rounded-lg hover:bg-zinc-100 dark:hover:bg-zinc-800 transition-colors text-zinc-500"
            title="Schliessen"
            @click="isOpen = false"
          >
            <i class="pi pi-times text-sm" />
          </button>
        </div>
      </div>

      <!-- Messages -->
      <div
        ref="messagesContainer"
        class="flex-1 overflow-y-auto p-4 space-y-3"
      >
        <div
          v-if="messages.length === 0"
          class="flex flex-col items-center justify-center h-full text-zinc-400 dark:text-zinc-500 text-sm text-center px-4"
        >
          <i class="pi pi-bolt text-3xl mb-3" />
          <p>Stellen Sie mir eine Frage zu Ihren Finanzdaten.</p>
        </div>

        <div
          v-for="(msg, i) in messages"
          :key="i"
          :class="msg.role === 'user' ? 'flex justify-end' : 'flex justify-start'"
        >
          <div
            :class="msg.role === 'user'
              ? 'bg-primary text-white rounded-2xl rounded-br-md px-4 py-2 max-w-[80%]'
              : 'bg-zinc-100 dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 rounded-2xl rounded-bl-md px-4 py-2 max-w-[80%]'"
            class="text-sm whitespace-pre-wrap break-words"
          >
            {{ msg.content }}
          </div>
        </div>

        <!-- Loading indicator -->
        <div
          v-if="isLoading"
          class="flex justify-start"
        >
          <div class="bg-zinc-100 dark:bg-zinc-800 rounded-2xl rounded-bl-md px-4 py-3">
            <ProgressSpinner
              style="width: 20px; height: 20px"
              stroke-width="4"
            />
          </div>
        </div>
      </div>

      <!-- Input -->
      <form
        class="shrink-0 border-t border-zinc-200 dark:border-zinc-700 p-3 pb-[calc(0.75rem+env(safe-area-inset-bottom,0px))]"
        @submit.prevent="handleSend"
      >
        <div class="flex items-center gap-2">
          <InputText
            ref="inputRef"
            v-model="inputMessage"
            class="flex-1"
            placeholder="Nachricht eingeben..."
            inputmode="text"
            :disabled="isLoading"
            @keydown.enter.exact.prevent="handleSend"
          />
          <Button
            icon="pi pi-send"
            :disabled="!inputMessage.trim() || isLoading"
            :loading="isLoading"
            size="small"
            @click="handleSend"
          />
        </div>
      </form>
    </div>
  </Transition>
</template>

<script setup lang="ts">
const { messages, isLoading, isOpen, chatbotActive, checkChatbotStatus, sendMessage, clearChat } = useChat()

const inputMessage = ref('')
const inputRef = ref()
const messagesContainer = ref<HTMLElement | null>(null)

onMounted(() => {
  checkChatbotStatus()
})

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

watch(() => messages.value.length, () => {
  scrollToBottom()
})

watch(isOpen, (open) => {
  if (open) {
    scrollToBottom()
    nextTick(() => {
      inputRef.value?.$el?.focus()
    })
    // Prevent body scroll on mobile when panel is open
    if (window.innerWidth < 640) {
      document.body.style.overflow = 'hidden'
    }
  }
  else {
    document.body.style.overflow = ''
  }
})

const handleSend = () => {
  const msg = inputMessage.value.trim()
  if (!msg || isLoading.value) return
  inputMessage.value = ''
  sendMessage(msg)
}
</script>
