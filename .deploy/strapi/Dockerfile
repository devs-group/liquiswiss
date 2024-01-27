FROM node:20.11.0 as dev

WORKDIR /app
COPY ./backend /app
RUN npm install
RUN npm install sharp

CMD ["npm", "run", "develop"]

FROM dev as prod

RUN npm run build
CMD ["npm", "run", "start"]
