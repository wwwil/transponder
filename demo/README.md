# Transponder Istio Demo

## Prerequisites

This demo requires a Google Cloud Platform (GCP) account.

The following applications are used:
- `istioctl` version 1.6.5
- `kubectl` version 1.15
- `gcloud` configured with your GCP account

All commands assume you are working from the `demo/` directory.

## Set Up

Set up the demo environment:

```bash
REGION=europe-west2
PROJECT=jetstack-wil
gcloud container clusters create transponder-demo --cluster-version 1.15.12-gke.6 --region $REGION --project $PROJECT
gcloud container clusters get-credentials transponder-demo --region $REGION --project $PROJECT
kubectl label namespace kube-node-lease istio-injection=disabled
kubectl label namespace kube-system istio-injection=disabled
kubectl label namespace kube-public istio-injection=disabled
kubectl label namespace default istio-injection=enabled
istioctl install --set profile=default --set addonComponents.prometheus.enabled=false
kubectl apply -f peerauthentication.yaml
```

## Deploy Transponder

Set the correct image tag for example Transponder server and scanner `Deployment` manifests and apply them to the
cluster:

```bash
kubectl apply -f ../deployments/transponder-server.yaml
kubectl apply -f ../deployments/transponder-scanner.yaml
```

## Check The Scanner

Check the logs of the scanner to verify that it is working:

```bash
POD=$(kubectl get pod -l app=transponder-scanner -o jsonpath="{.items[0].metadata.name}")
kubectl logs $POD transponder-scanner
```

The output should look something like this:

```bash
2020/07/29 20:09:57 Using config file from: /etc/transponder/scanner.yaml
2020/07/29 20:09:57 Get http://transponder-server:8080: dial tcp 10.15.253.36:8080: connect: connection refused
2020/07/29 20:09:57 Get https://transponder-server:8443: dial tcp 10.15.253.36:8443: connect: connection refused
2020/07/29 20:09:57 gRPC: Error making request: rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial tcp 10.15.253.36:8081: connect: connection refused"
2020/07/29 20:10:02 HTTP: Successfully made request: Hello from transponder-server-75b9c9c87b-hf7g9
2020/07/29 20:10:02 HTTPS: Successfully made request: Hello from transponder-server-75b9c9c87b-hf7g9
2020/07/29 20:10:02 gRPC: Successfully made request: Hello from transponder-server-75b9c9c87b-hf7g9
2020/07/29 20:10:07 HTTP: Successfully made request: Hello from transponder-server-75b9c9c87b-hf7g9
2020/07/29 20:10:07 HTTPS: Successfully made request: Hello from transponder-server-75b9c9c87b-hf7g9
2020/07/29 20:10:07 gRPC: Successfully made request: Hello from transponder-server-75b9c9c87b-hf7g9
```

There may be some requests that fail initially but this is normal.

## Expose Transponder

Deploy additional resources to make the Transponder server externally accessible:

```bash
kubectl apply -f transponder-server-external.yaml
```

This externally exposes insecure ports and should not be used in a production environment.

## Run Transponder Scanner Locally

Get the external IP of the Istio ingress gateway, copy the example scanner config and change it to point at this
address, then run `transponder scanner`.

```bash
INGRESS_IP=$(kubectl get -n istio-system service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
cp ../scanner.yaml .
sed -i '' 's/localhost/'"$INGRESS_IP"'/g' scanner.yaml
../transponder scanner --config-file=scanner.yaml
```

The output should look like this:

```bash
2020/07/29 22:03:35 Using config file from: scanner.yaml
2020/07/29 22:03:35 HTTP: Successfully made request: Hello from transponder-server-75b9c9c87b-hf7g9
2020/07/29 22:03:35 HTTPS: Successfully made request: Hello from transponder-server-75b9c9c87b-hf7g9
2020/07/29 22:03:35 gRPC: Successfully made request: Hello from transponder-server-75b9c9c87b-hf7g9
2020/07/29 22:03:40 HTTP: Successfully made request: Hello from transponder-server-75b9c9c87b-hf7g9
2020/07/29 22:03:40 HTTPS: Successfully made request: Hello from transponder-server-75b9c9c87b-hf7g9
2020/07/29 22:03:40 gRPC: Successfully made request: Hello from transponder-server-75b9c9c87b-hf7g9
```
