FROM node:carbon

WORKDIR /

# FRONTEND
ADD ./frontend/client/ /app/
RUN cd /app/ && npm i && npm run build
RUN mv /app/build/ /views/

COPY ./protoform-installer ./protoform

# Run protoform
CMD ["./protoform","install-operator"]
