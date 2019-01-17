# Runtime

### Development

#### Test

```
make test
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
