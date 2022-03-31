
## Generate certificates

First generate a self signed rsa key and certificate that the server can use for TLS.

```sh
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout /tmp/dt.key -out /tmp/dt.crt -subj "/CN=dependency-track-apiserver.dependency-track/O=dependency-track-apiserver.dependency-track"
```

## Create a dependency-track application running in a kubernetes cluster

### Create a secret for ingress.

```sh
kubectl create ns dependency-track
kubectl create secret tls dtsecret --key /tmp/dt.key --cert /tmp/dt.crt -n dependency-track 
```

### Deploy nginx ingress controller

```sh
helm upgrade --install ingress-nginx ingress-nginx \
  --repo https://kubernetes.github.io/ingress-nginx \
  --namespace ingress-nginx --create-namespace
```

### Deploy dependency-track

```sh
helm repo add evryfs-oss https://evryfs.github.io/helm-charts/
helm install dependency-track evryfs-oss/dependency-track --namespace dependency-track --create-namespace -f values.yaml
kubectl apply -f dependency-track.ingress.yaml
```

#### Get dependency-track API server LoadBalancer IP
```sh
API_SERVICE_IP=$(kubectl get svc -n dependency-track dependency-track-apiserver -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
echo $API_SERVICE_IP
34.69.242.184
```

#### Update API_BASE_URL env var in `values.yaml` with `$API_SERVICE_IP` value
example for `API_SERVICE_IP=34.69.242.184`:
```sh
    - name: API_BASE_URL
      value: "http://34.69.242.184:80"
```

#### Upgrade dependency-track to include new values
```sh
helm upgrade dependency-track evryfs-oss/dependency-track --namespace dependency-track --create-namespace -f values.yaml
kubectl apply -f dependency-track.ingress.yaml
```

### Get ingress LoadBalancer IP
```sh
INGRESSGATEWAY_SERVICE_IP=$(kubectl get svc -n ingress-nginx ingress-nginx-controller -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
echo $INGRESSGATEWAY_SERVICE_IP
34.135.8.34
```

### Add a dns record in `/etc/hosts` for the nginx LB IP
example for `INGRESSGATEWAY_SERVICE_IP=34.135.8.34`
```sh
$ cat /etc/hosts
##
# Host Database
#
# localhost is used to configure the loopback interface
# when the system is booting.  Do not change this entry.
##
127.0.0.1	localhost
255.255.255.255	broadcasthost
::1             localhost
# Added by Docker Desktop
# To allow the same kube context to work on the host and the container:
127.0.0.1 kubernetes.docker.internal
# End of section
34.135.8.34 dependency-track-apiserver.dependency-track
```

### Test with curl

```sh
curl -vvv -k https://dependency-track-apiserver.dependency-track/api/version
```

### Test scan command

#### Extract API Key (https://docs.dependencytrack.org/integrations/rest-api/)
1. `kubectl -n dependency-track port-forward svc/dependency-track-frontend 7777:80 &`
1. Browse to http://localhost:7777 (admin:admin)
1. Go to `Administration-->Access Management-->Teams` and get an API Key

#### Ingress (replace `XXX` with your API key)

```sh
SCANNER_DEPENDENCY_TRACK_INSECURE_SKIP_VERIFY=true \
SCANNER_DEPENDENCY_TRACK_DISABLE_TLS=false \
SCANNER_DEPENDENCY_TRACK_HOST=dependency-track-apiserver.dependency-track \
SCANNER_DEPENDENCY_TRACK_API_KEY=XXX \
./bin/kubeclarity scan sbom.cyclonedx  -i sbom -o sbom-result.json
```

#### Localhost (replace `XXX` with your API key)

```sh
k -n dependency-track port-forward svc/dependency-track-apiserver 8081:80

SCANNER_DEPENDENCY_TRACK_DISABLE_TLS=true \
SCANNER_DEPENDENCY_TRACK_HOST=localhost:8081 \
SCANNER_DEPENDENCY_TRACK_API_KEY=XXX \
./bin/kubeclarity scan sbom.cyclonedx  -i sbom -o sbom-result.json
```

### Cleanup
```sh
helm uninstall dependency-track -n dependency-track
helm uninstall ingress-nginx -n ingress-nginx
kubectl delete ns dependency-track ingress-nginx
```