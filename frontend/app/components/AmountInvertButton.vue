<template>
  <Button
    v-tooltip.top="isAmountNegative ? 'Betrag positiv machen' : 'Betrag negativ machen'"
    :severity="isAmountNegative ? 'info' : 'danger'"
    icon="pi"
    :icon-class="{ 'pi-minus-circle': !isAmountNegative, 'pi-plus-circle': isAmountNegative }"
    :disabled="normalizedAmount == 0"
    @click="$emit('invert-amount')"
  />
</template>

<script lang="ts" setup>
const props = defineProps({
  amount: {
    type: [Number, String],
    required: true,
  },
})
defineEmits(['invert-amount'])

const normalizedAmount = computed(() => typeof props.amount == 'string' ? Number(props.amount) : props.amount)
const isAmountNegative = computed(() => normalizedAmount.value < 0)
</script>
