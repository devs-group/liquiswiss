<template>
  <div class="flex flex-col gap-4">
    <div class="flex justify-between items-center gap-2">
      <hr class="h-0.5 bg-black flex-1">
      <p class="text-xl">
        App Einstellungen
      </p>
      <hr class="h-0.5 bg-black flex-1">
    </div>

    <p class="text-xs text-right">
      Hinweis: Diese Einstellungen werden pro Browser gespeichert
    </p>

    <div class="grid grid-cols-2 gap-4">
      <div class="flex flex-col justify-center h-full col-span-full md:col-span-1 bg-zinc-100 dark:bg-zinc-800 p-2">
        <div class="flex items-center gap-2 ">
          <Checkbox
            v-model="skipOrganisationSwitchQuestion"
            binary
            input-id="skip-organisation-switch-question"
            name="skip-organisation-switch-question"
          />
          <label
            class="cursor-pointer"
            for="skip-organisation-switch-question"
          >Nicht nachfragen beim Wechseln der Organisation</label>
        </div>
      </div>

      <div class="flex flex-col justify-center h-full col-span-full md:col-span-1 bg-zinc-100 dark:bg-zinc-800 p-2">
        <div class="flex items-center gap-2 col-span-full md:col-span-1">
          <label for="dark-mode">Farbmodus:</label>
          <ClientOnly>
            <SelectButton
              v-model="colorMode.preference"
              :options="darkModeOptions"
              option-label="label"
              option-value="value"
            />
          </ClientOnly>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Config } from '~/config/config'

useHead({
  title: 'App Einstellungen',
})

const toast = useToast()

const colorMode = useColorMode()
const { skipOrganisationSwitchQuestion } = useSettings()

watch([skipOrganisationSwitchQuestion], (value, oldValue) => {
  if (value !== oldValue) {
    toast.add({
      summary: 'Erfolg',
      detail: `Einstellung gespeichert`,
      severity: 'info',
      life: Config.TOAST_LIFE_TIME_SHORT,
    })
  }
})

const darkModeOptions = computed(() => {
  return DarkModeOptions.map((value) => {
    let label
    switch (value) {
      case 'dark':
        label = 'Dunkel'
        break
      case 'light':
        label = 'Hell'
        break
      default:
        label = 'System'
    }
    return {
      label: label,
      value: value,
    }
  })
})
</script>
