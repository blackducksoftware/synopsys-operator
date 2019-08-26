FROM nginx

RUN mkdir -p /opt/dist/

COPY dist /opt/dist/

ADD ./docker/nginx.conf /etc/nginx/conf.d/default.conf

CMD ["nginx", "-g", "daemon off;"]
