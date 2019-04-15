# Runtime

### Development

#### Test

```
make test
```
### Setup development environment

- Create a secret for docker-registry

```
apiVersion: v1
kind: Secret
metadata:
  name: docker-reg-credential
  annotations:
    build.knative.dev/docker-0: https://index.docker.io/v1/
type: kubernetes.io/basic-auth
data:
  # Use 'echo -n "username" | base64' to generate this string
  username: <>
  # Use 'echo -n "password" | base64' to generate this string
  password: <>
```

- Create a service account for the above secret

```
apiVersion: v1
kind: ServiceAccount
metadata:
  name: build-bot
secrets:
- name: docker-reg-credential
```

- Create configmaps for Dockerfiles supporting the programming languages which are nodejs6 and nodejs8 as of now

```
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    function: runtime-function-controller
  name: dockerfileNodejs6
data:
  Dockerfile: |-
    FROM kubeless/nodejs@sha256:5c3c21cf29231f25a0d7d2669c6f18c686894bf44e975fcbbbb420c6d045f7e7
    USER root
    RUN export KUBELESS_INSTALL_VOLUME='/kubeless' && \
        mkdir /kubeless && \
        cp /src/handler.js /kubeless && \
        cp /src/package.json /kubeless && \
        /kubeless-npm-install.sh
    USER 1000
```

```
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    function: runtime-function-controller
  name: dockerfile-nodejs-8
data:
  Dockerfile: |-
    FROM kubeless/nodejs@sha256:5c3c21cf29231f25a0d7d2669c6f18c686894bf44e975fcbbbb420c6d045f7e7
    USER root
    RUN export KUBELESS_INSTALL_VOLUME='/kubeless' && \
        mkdir /kubeless && \
        cp /src/handler.js /kubeless && \
        cp /src/package.json /kubeless && \
        /kubeless-npm-install.sh
    USER 1000
```

- Create the following configmap which serves as configuration for the runtime controller

```
apiVersion: v1
data:
  dockerRegistry: <>  ### dockerhub handle
  runtimes: |
    - ID: nodesjs8
      dockerFileName: dockerfile-nodejs-8
    - ID: nodejs6
      dockerFileName: dockerfile-nodejs-6
  serviceAccountName: runtime-controller
kind: ConfigMap
metadata:
  labels:
    app: kubeless
  name: runtime-controller-config
```


#### Install the CRD to a local Kubernetes cluster

```
make install
```
#### Build and run the manager
```
make run
```

# Create a docker image

```
make docker-build IMG=<img-name>
```

# Push the docker image to a configured container registry

```
make docker-push IMG=<img-name>
```
