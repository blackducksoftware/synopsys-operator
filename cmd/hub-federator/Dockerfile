FROM node:carbon

WORKDIR /

# FRONTEND
ADD ./frontend/client/ /app/
RUN cd /app/ && npm i && npm run build
RUN mv /app/build/ /views/

COPY ./hub-federator ./hub

# Run the hub federator
CMD ["./hub"]
