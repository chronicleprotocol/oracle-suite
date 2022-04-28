# E.g.
#   docker build -f Dockerfile . -t oracle-suite-ethutil
#   docker run oracle-suite-ethutil compute-contract-address 0xdd2fd4581271e230360230f9337d5c0430bf44c0 15
FROM node
RUN mkdir /app
WORKDIR /app
RUN npm install rlp
RUN npm install keccak
ADD ./main.js main.js
ADD ./package.json package.json
RUN npm install -g .
