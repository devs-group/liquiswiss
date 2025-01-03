export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()
  return await proxyRequest(event, `${config.apiHost}${event.path}`)
})
