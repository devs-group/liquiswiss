// Shared blink helper for real-time updates. Cleanup is timeout-based, NOT
// animationend: animationend bubbles up from child animations (PrimeVue
// transitions) and would strip the class before the blink is visible.

// Keep in sync with the animation in tailwind.css (3 x 0.5s)
export const REALTIME_FLASH_MS = 1600

export function flashRealtimeElement(element: Element, className = 'realtime-flash') {
  if (element.classList.contains(className)) return
  element.classList.remove('realtime-flash', 'realtime-flash-delete')
  // Force reflow so re-adding the class restarts the animation
  void (element as HTMLElement).offsetWidth
  element.classList.add(className)
  setTimeout(() => element.classList.remove(className), REALTIME_FLASH_MS)
}

// Flash all elements matching a selector (e.g. a specific form input)
export function flashRealtimeSelector(selector: string, className = 'realtime-flash') {
  document.querySelectorAll(selector).forEach(element => flashRealtimeElement(element, className))
}
