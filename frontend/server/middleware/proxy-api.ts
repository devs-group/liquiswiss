export default defineEventHandler(async (event) => {
  if (!event.path.startsWith('/api/') || (event.path.startsWith('/api/global'))) {
    return
  }

  const config = useRuntimeConfig()
  console.log(
    `${getMethod(event)}: ${event.path}\n\t${config.apiHost}${event.path}`,
  )
  try {
    return await proxyRequest(event, `${config.apiHost}${event.path}`, {
      headers: getRequestHeaders(event),
    })
  }
  catch (error) {
    console.error(error)
  }
})
