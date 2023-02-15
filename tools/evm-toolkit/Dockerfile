FROM node:18 AS build

WORKDIR /app

COPY package.json ./
COPY package-lock.json ./

RUN npm install

COPY . ./

RUN npm run build --prod

FROM nginx:1.23-alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY --from=build /app/networks.json /usr/share/nginx/html