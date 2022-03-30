# Keyless signature

>This is only for testing purpose.
For pruduction ready install we need to add ingress to Rekor, Fulcio and Dex because the Cosign will talk to them during signing and verification.

## Prerequirements

### 1. Install Rekor server using official helm chart along with trillinan log server and log signer.

```
helm upgrade -i rekor sigstore/rekor --set server.ingress.enabled=false
```

### 2. Define upstream identity provider for dex.

On github developer setings (https://github.com/settings/developers) create a new OAuth app which has `client ID` `client secret`.
Set the `Authorization callback URL` to http://dex:8888/auth/callback
In the `dex-config.yaml` set up the generated `client ID and` and `client secret`:
```
      clientID: "github_client_id"
      clientSecret: "github_client_secret"
      redirectURI: http://dex:8888/auth/callback
```

### 3. Install dex-idp using official chart and passing the configs.

```
helm upgrade -i dex dex/dex -f dex-config.yaml
```

### 4. Generate certificates and keys using OpenSSL

```
cd certs
```
Generate rootCA key for Fulcio
```
openssl genpkey -algorithm ed25519 -aes256 -out rootCAKey.pem
```
Enter password

Creating self signed cert for Fulcio
```
openssl req -x509 -sha256 -new -nodes -key rootCAKey.pem -days 3650 -subj "/C=US/ST=CA/O=MyOrg, Inc./CN=mydomain.com" -extensions v3_ca -config ssl.cnf -out rootCACert.pem
```

Generate public and private key for ct-log server
```
openssl genrsa -aes128 -out priv.pem
```
Enter password


Get public key from private key.
```
openssl rsa -in priv.pem -pubout > pub.pem
```

Quit from cert dir
```
cd ..
```

The passwords will use during `6. Install the keyless-signature chart from dir`.

### 5. Create tree in trillian log server

First we need to install the `createtree` tool:
```
go install github.com/google/trillian/cmd/createtree@latest
```

Port forwarding `trillian-log-server`
```
kubectl port-forward svc/rekor-trillian-log-server 8091:8091
```

```
LOG_ID="$(createtree --admin_server localhost:8091)"
```

Checking the `LOG_ID`
```
echo $LOG_ID
```

### 6. Install the keyless-signature chart from dir.

```
helm upgrade -i keyless --set ctfe.logID=${LOG_ID} --set ctfe.keyPass=<your_ctfe_key_pass>  --set fulcio.keyPass=<your_fulcio_key_pass> . 
```

### 8. Port forwarding the Rekor, Dex and Fulcio you can use it for 

Add dex for your hosts as 127.0.0.1.
It is crucial because the issuer host inside and outside of the cluster should be same if you port forward Dex.
If you are using port-forward on your computer you will have to add dex to the hosts file.

/etc/hosts
```
127.0.0.1 dex
```

```
kubectl port-forward svc/keyless-keyless-signature-fulcio 5555:5555 &
kubectl port-forward svc/rekor-server 3000:3000 &
kubectl port-forward svc/dex 8888:8888 &
```


### 9. After port forwarding Rekor, Dex and Fulcio we can signing images.

 Cosing by defult is using sigstore hosted Rekor, Fulcio and identity provider, you have to override it.

```
COSIGN_EXPERIMENTAL=1 cosign sign --insecure-skip-verify \
    -oidc-issuer "http://dex:8888/auth" \
    -fulcio-url "http://localhost:5555" \
    -rekor-url "http://localhost:3000" \
    image
```

### 10. Verifying the image.

You have to set the root CA for cosign.
You can get it from secret:
```
kubectl get secret keyless-keyless-signature -o jsonpath='{.data.rootca}' | base64 --decod > /tmp/keyless.rootca
```

```
SIGSTORE_ROOT_FILE="/tmp/keyless.rootca" COSIGN_EXPERIMENTAL=1 cosign verify \
    -rekor-url "http://localhost:3000" \
    image
```
