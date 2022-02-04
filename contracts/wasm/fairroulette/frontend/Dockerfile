FROM node:lts-buster as build

COPY package.json .
COPY package-lock.json .

RUN npm install

COPY . .

RUN mv docker/config.dev.docker.js config.dev.js

RUN npm run build_worker
RUN npm run build

FROM nginx:alpine

ARG WASP_URL=https://127.0.0.1
ARG WASP_WS_URL=wss://127.0.0.1
ARG GOSHIMMER_URL=https://127.0.0.1
ARG CHAIN_ID=qawsedrftgzhujikolp
ARG CONTRACT_NAME=fairroulette
ARG GOOGLE_ANALITICS_ID

COPY --from=build public /usr/share/nginx/html
COPY docker/nginx.conf /etc/nginx/conf.d/default.conf

RUN sed -i "s|#WASP_URL#|${WASP_URL}|g" /usr/share/nginx/html/build/bundle.js
RUN sed -i "s|#WASP_WS_URL#|${WASP_WS_URL}|g" /usr/share/nginx/html/build/bundle.js
RUN sed -i "s|#GOSHIMMER_URL#|${GOSHIMMER_URL}|g" /usr/share/nginx/html/build/bundle.js
RUN sed -i "s|#CHAIN_ID#|${CHAIN_ID}|g" /usr/share/nginx/html/build/bundle.js
RUN sed -i "s|#CONTRACT_NAME#|${CONTRACT_NAME}|g" /usr/share/nginx/html/build/bundle.js
RUN sed -i "s|#GOOGLE_ANALITICS_ID#|${GOOGLE_ANALITICS_ID}|g" /usr/share/nginx/html/build/bundle.js

RUN find /usr/share/nginx/html -type d -exec chmod 755 {} \;
RUN find /usr/share/nginx/html -type f -exec chmod 644 {} \;
