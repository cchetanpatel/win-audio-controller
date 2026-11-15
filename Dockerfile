# build env

FROM node:27.0.0-alpha.0.1.0-alpine3.24 AS build
WORKDIR /app/

COPY ./app /app

RUN npm install
RUN npm run build


FROM nginx:1.10.1-alpine

COPY --from=build /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
#COPY akl01vftracktaskswidget.crt /etc/nginx/akl01vftracktaskswidget.crt
#COPY akl01vftracktaskswidget.key /etc/nginx/akl01vftracktaskswidget.key

CMD ["nginx", "-g", "daemon off;"]
