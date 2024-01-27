FROM node:20.11.0 as dev

WORKDIR /app
COPY ./frontend /app
RUN npm install

CMD ["npm", "run", "dev"]

FROM dev as prod

RUN npm run build
CMD ["node", ".output/server/index.mjs"]
