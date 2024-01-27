FROM node:20.3.1 as dev

WORKDIR /app
COPY ./frontend /app
RUN npm install

CMD ["npm", "run", "dev"]

FROM dev as prod

RUN npm run build
CMD ["node", ".output/server/index.mjs"]
