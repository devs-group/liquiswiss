<template>
  <IconField>
    <InputText
      :model-value="modelValue"
      :placeholder="placeholder"
      @update:model-value="onInput"
    />
    <InputIcon
      v-show="modelValue"
      class="pi pi-times cursor-pointer"
      @click="onClear"
    />
  </IconField>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  modelValue: string
  placeholder?: string
}>(), {
  placeholder: 'Suchen',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'clear': []
}>()

const onInput = (value: string) => {
  emit('update:modelValue', value)
}

const onClear = () => {
  emit('update:modelValue', '')
  emit('clear')
}
</script>
