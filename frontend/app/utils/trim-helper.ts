// Trims all top-level string values of a form payload before submit.
// yup's .trim() only validates, it does not transform submitted values.
// Never use on password fields: whitespace there is significant.
export const trimStringValues = <T extends Record<string, unknown>>(values: T): T => {
  return Object.fromEntries(
    Object.entries(values).map(([key, value]) => [
      key,
      typeof value === 'string' ? value.trim() : value,
    ]),
  ) as T
}
