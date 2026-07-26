export default defineEventHandler(async (event) => {
  if (!event.path.startsWith('/api/') || (event.path.startsWith('/api/global'))) {
    return
  }

  const config = useRuntimeConfig()
  console.log(
    `${getMethod(event)}: ${event.path}\n\t${config.apiHost}${event.path}`,
  )

  // Strip hop-by-hop headers so clients cannot smuggle upgrade semantics
  const headers = getRequestHeaders(event)
  delete headers.connection
  delete headers.upgrade
  delete headers['keep-alive']
  delete headers['proxy-connection']
  delete headers['transfer-encoding']

  try {
    return await proxyRequest(event, `${config.apiHost}${event.path}`, {
      headers,
    })
  }
  catch (error) {
    console.error(error)
    // Surface backend failures instead of returning an empty 200
    throw createError({ statusCode: 502, statusMessage: 'Bad Gateway' })
  }
})
