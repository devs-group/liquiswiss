// Container health probe: answers without rendering a page, so no profile call hits the backend
export default defineEventHandler(() => 'ok')
