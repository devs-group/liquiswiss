<script setup lang="ts">
import { Config } from '~/config/config'

const route = useRoute()
const toast = useToast()

const clientId = computed(() => String(route.query.client_id ?? ''))
const clientName = computed(() => String(route.query.client_name ?? 'Unbekannte Anwendung'))
const redirectUri = computed(() => String(route.query.redirect_uri ?? ''))
const codeChallenge = computed(() => String(route.query.code_challenge ?? ''))
const state = computed(() => String(route.query.state ?? ''))
const resource = computed(() => String(route.query.resource ?? ''))

const isValidRequest = computed(() => !!clientId.value && !!redirectUri.value && !!codeChallenge.value)
const isLoading = ref(false)

const respond = async (approve: boolean) => {
  isLoading.value = true
  try {
    const response = await $fetch<{ redirect: string }>('/api/oauth/approve', {
      method: 'POST',
      body: {
        clientId: clientId.value,
        redirectUri: redirectUri.value,
        codeChallenge: codeChallenge.value,
        state: state.value,
        resource: resource.value,
        approve,
      },
    })
    window.location.href = response.redirect
  }
  catch {
    isLoading.value = false
    toast.add({
      summary: 'Fehler',
      detail: 'Die Anfrage konnte nicht verarbeitet werden',
      severity: 'error',
      life: Config.TOAST_LIFE_TIME,
    })
  }
}
</script>

<template>
  <div class="flex flex-col items-center gap-4 max-w-lg mx-auto py-10">
    <Logo class="!text-3xl" />

    <h1 class="text-2xl font-bold text-center">
      Zugriff erlauben?
    </h1>

    <template v-if="isValidRequest">
      <p class="text-center">
        <strong>{{ clientName }}</strong> möchte auf Ihre LiquiSwiss-Daten zugreifen
        (Bankkonten, Transaktionen, Mitarbeitende, Löhne und Prognosen Ihrer aktuellen Organisation).
      </p>

      <Message
        severity="warn"
        class="w-full"
      >
        Mit dem Verbinden stimmen Sie zu, dass Ihre LiquiSwiss-Daten an das von Ihnen
        gewählte KI-Modell (LLM) übertragen werden. Ihre Daten laufen dabei über die
        Server des jeweiligen Anbieters und unterliegen dessen Datenschutzrichtlinien.
      </Message>

      <div class="flex gap-2 w-full">
        <Button
          label="Ablehnen"
          severity="secondary"
          class="flex-1"
          :disabled="isLoading"
          data-testid="oauth-deny-button"
          @click="respond(false)"
        />
        <Button
          label="Erlauben"
          class="flex-1"
          :loading="isLoading"
          data-testid="oauth-approve-button"
          @click="respond(true)"
        />
      </div>

      <p class="text-sm text-center opacity-70">
        Sie können den Zugriff jederzeit in den Profileinstellungen unter
        "Verbundene Anwendungen" widerrufen.
      </p>
    </template>

    <Message
      v-else
      severity="error"
      class="w-full"
    >
      Ungültige Autorisierungsanfrage.
    </Message>
  </div>
</template>
