FROM node:carbon

WORKDIR /

# FRONTEND
ADD ./frontend/client/ /app/
RUN cd /app/ && npm i && npm run build
RUN mv /app/build/ /views/

COPY ./operator-build ./operator

# Run protoform
CMD ["./operator", "install-operator"]
