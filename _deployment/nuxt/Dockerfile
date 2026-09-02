FROM node:22.19.0 AS dev

WORKDIR /app
COPY ./frontend /app
RUN npm install

CMD ["npm", "run", "dev-host"]

FROM dev AS prod

RUN npm run build

FROM node:22.19-alpine3.22 AS deploy

# Same build identifiers CI stamps into the backend, kept on the image so
# `docker inspect` can answer what this container was built from.
ARG GIT_SHA=dev
ARG VERSION=dev
ARG APP_SHA=
ENV GIT_SHA=${GIT_SHA} VERSION=${VERSION} APP_SHA=${APP_SHA}

WORKDIR /app

COPY --from=prod /app/.output /app/output

# The deploy webhook waits on this before considering the release healthy.
HEALTHCHECK --interval=10s --timeout=5s --start-period=30s --retries=5 \
  CMD ["node", "-e", "fetch('http://127.0.0.1:3000/').then(r=>process.exit(r.ok?0:1)).catch(()=>process.exit(1))"]

CMD ["node", "output/server/index.mjs"]
