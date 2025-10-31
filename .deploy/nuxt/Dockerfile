FROM node:22.19.0 as dev

WORKDIR /app
COPY ./frontend /app
RUN npm install

CMD ["npm", "run", "dev"]

FROM dev as prod

RUN npm run build

FROM node:22.19-alpine3.22 AS deploy

WORKDIR /app

COPY --from=prod /app/.output /app/output

CMD ["node", "output/server/index.mjs"]
