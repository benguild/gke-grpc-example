# gke-grpc-example
Example of how to deploy gRPC (or a plain HTTP/2 server) on Google Kubernetes Engine on Google Cloud Platform that's compatible with Let's Encrypt for HTTPS/TLS at the load balancer level.

https://benguild.com/2018/11/11/quickstart-golang-kubernetes-grpc-tls-lets-encrypt/

## Setup instructions

Clone the repo within your `GOPATH`. — If you don't have one, Google what that is and setup your Go environment first! 😅

Locally, generate your protos using the following command:

```bash
protoc -I=./protos-src --go_out=plugins=grpc:protos ./protos-src/*.proto
```

Then, be sure that you have Go Dep, and ensure your dependencies are available:

```bash
dep ensure
```

... Cool. 👍🏻 — Then, continue to follow the instructions in the blog post linked above!
