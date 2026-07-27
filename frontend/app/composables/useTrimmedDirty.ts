// Trim-aware dirty state for vee-validate forms: whitespace-only edits
// (leading/trailing) do not count as changes, inner whitespace does.
// Empty string and undefined/null are treated as the same "no value".

import type { FormMeta } from 'vee-validate'
import type { ComputedRef } from 'vue'

const normalize = (value: unknown): unknown => {
  if (value === undefined || value === null || value === '') return null
  if (typeof value === 'string') {
    const trimmed = value.trim()
    return trimmed === '' ? null : trimmed
  }
  if (value instanceof Date) return value.getTime()
  if (Array.isArray(value)) return value.map(normalize)
  if (typeof value === 'object') {
    return Object.fromEntries(
      Object.entries(value as Record<string, unknown>).map(([key, entry]) => [key, normalize(entry)]),
    )
  }
  return value
}

const deepEqual = (a: unknown, b: unknown): boolean => {
  if (a === b) return true
  if (Array.isArray(a) && Array.isArray(b)) {
    return a.length === b.length && a.every((entry, index) => deepEqual(entry, b[index]))
  }
  if (a && b && typeof a === 'object' && typeof b === 'object') {
    const keys = new Set([...Object.keys(a), ...Object.keys(b)])
    return [...keys].every(key =>
      deepEqual((a as Record<string, unknown>)[key], (b as Record<string, unknown>)[key]),
    )
  }
  return false
}

export default function useTrimmedDirty<T extends Record<string, unknown>>(
  meta: ComputedRef<FormMeta<T>>,
  values: T,
): ComputedRef<boolean> {
  return computed(() => {
    // Fast path: untouched form is never dirty
    if (!meta.value.dirty) return false
    return !deepEqual(normalize({ ...values }), normalize({ ...meta.value.initialValues }))
  })
}
