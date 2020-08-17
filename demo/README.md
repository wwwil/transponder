# Transponder Istio Demo

## Prerequisites

This demo requires a Google Cloud Platform (GCP) account.

The following applications are required:
- `istioctl` version 1.6.5 or newer.
- `kubectl` with a version matching the clusters.
- `gcloud` configured with your GCP account.

All commands assume you are working from the `demo/` directory and use the same shell throughout the demo to reuse
environment variables.

## 00. Set Up Some Clusters With Istio

Set environment variables for the demo:

```bash
# The ID of the GCP project to use.
PROJECT=jetstack-wil
# The GCP zones for the two clusters.
ZONE_1=europe-west2-a
ZONE_2=europe-west2-b
```

Set up the first cluster:

```bash
gcloud container clusters create transponder-demo-1 --release-channel=regular --num-nodes=3 --zone $ZONE_1 --project $PROJECT
gcloud container clusters get-credentials transponder-demo-1 --zone $ZONE_1 --project $PROJECT
kubectl label namespace kube-node-lease istio-injection=disabled
kubectl label namespace kube-system istio-injection=disabled
kubectl label namespace kube-public istio-injection=disabled
kubectl label namespace default istio-injection=enabled
kubectl create namespace istio-system
kubectl create secret generic cacerts -n istio-system --from-file=00-certs/ca-cert.pem \
    --from-file=00-certs/ca-key.pem --from-file=00-certs/root-cert.pem --from-file=00-certs/cert-chain.pem
istioctl install -f 00-istiooperator.yaml
ISTIOCOREDNS_CLUSTERIP=$(kubectl get svc -n istio-system istiocoredns -o jsonpath={.spec.clusterIP})
cp 00-kubedns-config.template.yaml 00-kubedns-config-1.yaml
sed -i '' 's/istiocoredns_clusterip/'"$ISTIOCOREDNS_CLUSTERIP"'/g' 00-kubedns-config-1.yaml
kubectl apply -f 00-kubedns-config-1.yaml
kubectl delete pods -n kube-system -l k8s-app=kube-dns
kubectl apply -f 00-peerauthentication.yaml
```

Then set up another cluster:

```bash
gcloud container clusters create transponder-demo-2 --release-channel=regular --num-nodes=3 --zone $ZONE_2 --project $PROJECT
gcloud container clusters get-credentials transponder-demo-2 --zone $ZONE_2 --project $PROJECT
kubectl label namespace kube-node-lease istio-injection=disabled
kubectl label namespace kube-system istio-injection=disabled
kubectl label namespace kube-public istio-injection=disabled
kubectl label namespace default istio-injection=enabled
kubectl create namespace istio-system
kubectl create secret generic cacerts -n istio-system --from-file=00-certs/ca-cert.pem \
    --from-file=00-certs/ca-key.pem --from-file=00-certs/root-cert.pem --from-file=00-certs/cert-chain.pem
istioctl install -f 00-istiooperator.yaml
ISTIOCOREDNS_CLUSTERIP=$(kubectl get svc -n istio-system istiocoredns -o jsonpath={.spec.clusterIP})
cp 00-kubedns-config.template.yaml 00-kubedns-config-2.yaml
sed -i '' 's/istiocoredns_clusterip/'"$ISTIOCOREDNS_CLUSTERIP"'/g' 00-kubedns-config-2.yaml
kubectl apply -f 00-kubedns-config-2.yaml
kubectl delete pods -n kube-system -l k8s-app=kube-dns
kubectl apply -f 00-peerauthentication.yaml
```

Note that these clusters are using the same example certificates. This is a quick way to ensure they share a root of
trust without generating individual certificates. This approach is not suitable for a production environment.

## 01. Deploy Transponder To The Cluster

Change `kubectl` to use the first cluster:

```bash
kubectl config use-context gke_${PROJECT}_${ZONE_1}_transponder-demo-1
```

Deploy the Transponder server and scanner into the cluster: 

```bash
kubectl apply -f ../deployments/transponder-server.yaml
kubectl apply -f ../deployments/transponder-scanner.yaml
```

Wait for the `Pods` to start, then check the logs of the scanner to verify that it is working:

```bash
POD=$(kubectl get pod -l app=transponder-scanner -o jsonpath="{.items[0].metadata.name}")
kubectl logs -f $POD transponder-scanner
```

The output should look something like this:

```bash
2020/07/29 20:09:57 Transponder version: v0.0.1 linux/amd64
2020/07/29 20:09:57 Transponder scanner is starting.
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

## 02. Expose Transponder Externally and Connect Locally

Deploy additional resources to make the Transponder server externally accessible via the Istio ingress gateway:

```bash
kubectl apply -f 02-transponder-server-external.yaml
```

This externally exposes insecure ports and should not be used in a production environment.

Get the external IP of the Istio ingress gateway, copy the example scanner config and change it to point at this
address:

```bash
INGRESS_GATEWAY_IP_1=$(kubectl get -n istio-system service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
cp ../scanner.yaml 02-scanner.yaml
sed -i '' 's/localhost/'"$INGRESS_GATEWAY_IP_1"'/g' 02-scanner.yaml
```

Download the Transponder binary for your platform and run the scanner it with the modified config file:

```bash
OS=darwin
ARCH=amd64
curl -L https://github.com/wwwil/transponder/releases/download/v0.0.1/transponder-v0.0.1-$OS-$ARCH.zip -O
unzip -o transponder-v0.0.1-$OS-$ARCH.zip
rm transponder-v0.0.1-$OS-$ARCH.zip
./transponder scanner --config-file=02-scanner.yaml
```

The output should look like this:

```bash
2020/08/07 18:05:35 Transponder version: v0.0.1 darwin/amd64
2020/08/07 18:05:35 Transponder scanner is starting.
2020/08/07 18:05:35 Using config file from: 02-scanner.yaml
2020/08/07 18:05:35 34.89.60.18:8080 HTTP: Successfully made request: Hello from transponder-server-78454877fb-lw25v
2020/08/07 18:05:35 34.89.60.18:8443 HTTPS: Successfully made request: Hello from transponder-server-78454877fb-lw25v
2020/08/07 18:05:35 34.89.60.18:8081 GRPC: Successfully made request: Hello from transponder-server-78454877fb-lw25v
2020/08/07 18:05:40 34.89.60.18:8080 HTTP: Successfully made request: Hello from transponder-server-78454877fb-lw25v
2020/08/07 18:05:40 34.89.60.18:8443 HTTPS: Successfully made request: Hello from transponder-server-78454877fb-lw25v
2020/08/07 18:05:40 34.89.60.18:8081 GRPC: Successfully made request: Hello from transponder-server-78454877fb-lw25v
```

The local Transponder scanner is connecting to the server via the Istio ingress gateway's external IP address.

## 03. Multi Cluster

Change `kubectl` to use the second cluster:

```bash
kubectl config use-context gke_${PROJECT}_${ZONE_2}_transponder-demo-2
```

Deploy the Transponder server to the cluster:

```bash
kubectl apply -f ../deployments/transponder-server.yaml
```

Get the ingress gateway IP of the new cluster and add it to the `ServiceEntry` manifests for the first cluster:

```bash
INGRESS_GATEWAY_IP_2=$(kubectl get -n istio-system service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
cp 03-transponder-server-multi-cluster.template.yaml 03-transponder-server-multi-cluster.yaml
sed -i '' 's/ingress_gateway_ip/'"$INGRESS_GATEWAY_IP_2"'/g' 03-transponder-server-multi-cluster.yaml
```

Switch back to the first cluster to create a `ServiceEntry` for the Transponder server in the new cluster and update
the Transponder scanner config:

```bash
kubectl config use-context gke_${PROJECT}_${ZONE_1}_transponder-demo-1
kubectl apply -f 03-transponder-server-multi-cluster.yaml
POD=$(kubectl get pod -l app=transponder-scanner -o jsonpath="{.items[0].metadata.name}")
kubectl delete pod $POD
POD=$(kubectl get pod -l app=transponder-scanner -o jsonpath="{.items[0].metadata.name}")
kubectl logs $POD transponder-scanner
```

The output should look like this:

```bash
2020/08/07 16:47:15 Transponder version: v0.0.1 linux/amd64
2020/08/07 16:47:15 Transponder scanner is starting.
2020/08/07 16:47:16 Using config file from: /etc/transponder/scanner.yaml
2020/08/07 16:47:16 transponder-server:8080 HTTP: Error making request: "Get http://transponder-server:8080: dial tcp 10.31.253.217:8080: connect: connection refused"
2020/08/07 16:47:16 transponder-server:8443 HTTPS: Error making request: "Get https://transponder-server:8443: dial tcp 10.31.253.217:8443: connect: connection refused"
2020/08/07 16:47:16 transponder-server:8081 GRPC: Error making request: "rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing dial tcp 10.31.253.217:8081: connect: connection refused\""
2020/08/07 16:47:21 transponder-server.default.global:8080 HTTP: Successfully made request: Hello from transponder-server-78454877fb-rh59p
2020/08/07 16:47:21 transponder-server.default.global:8443 HTTPS: Successfully made request: Hello from transponder-server-78454877fb-rh59p
2020/08/07 16:47:21 transponder-server.default.global:8081 GRPC: Successfully made request: Hello from transponder-server-78454877fb-rh59p
2020/08/07 16:47:26 transponder-server:8080 HTTP: Successfully made request: Hello from transponder-server-78454877fb-lw25v
2020/08/07 16:47:26 transponder-server:8443 HTTPS: Successfully made request: Hello from transponder-server-78454877fb-lw25v
2020/08/07 16:47:26 transponder-server:8081 GRPC: Successfully made request: Hello from transponder-server-78454877fb-lw25v
2020/08/07 16:47:31 transponder-server.default.global:8080 HTTP: Successfully made request: Hello from transponder-server-78454877fb-rh59p
2020/08/07 16:47:31 transponder-server.default.global:8443 HTTPS: Successfully made request: Hello from transponder-server-78454877fb-rh59p
2020/08/07 16:47:31 transponder-server.default.global:8081 GRPC: Successfully made request: Hello from transponder-server-78454877fb-rh59p
```

The scanner is now be connecting to both `transponder-server` in the local cluster and
`transponder-server.default.global` in the remote cluster.

Again there may be some requests that fail initially but this is normal.

## 04. Per Cluster DNS

Create an `EnvoyFilter` to rewrite hostnames in the second cluster:

```bash
kubectl config use-context gke_${PROJECT}_${ZONE_2}_transponder-demo-2
kubectl apply -f 04-envoy-filter.yaml
```

Create a `ServiceEntry` in the first cluster that points at the second cluster's Transponder server using the new
cluster specific name:

```bash
cp 04-transponder-server-multi-cluster-dns.template.yaml 04-transponder-server-multi-cluster-dns.yaml
sed -i '' 's/ingress_gateway_ip/'"$INGRESS_GATEWAY_IP_2"'/g' 04-transponder-server-multi-cluster-dns.yaml
kubectl config use-context gke_${PROJECT}_${ZONE_1}_transponder-demo-1
kubectl apply -f 04-transponder-server-multi-cluster-dns.yaml
POD=$(kubectl get pod -l app=transponder-scanner -o jsonpath="{.items[0].metadata.name}")
kubectl delete pod $POD
POD=$(kubectl get pod -l app=transponder-scanner -o jsonpath="{.items[0].metadata.name}")
kubectl logs $POD transponder-scanner
```

## 05. Service Aliases

Add new stub zone to Kube DNS in the first cluster. It will update configuration without needing to restart:

```bash
kubectl config use-context gke_${PROJECT}_${ZONE_1}_transponder-demo-1
COREDNS_SERVICE_IP_1=$()
cp 05-kube-dns-configmap.template.yaml 05-kube-dns-configmap-transponder-demo-1.yaml
sed -i '' 's/coredns_service_ip/'"$COREDNS_SERVICE_IP_1"'/g' 05-kube-dns-configmap-transponder-demo-1.yaml
kubectl apply 05-kube-dns-configmap-transponder-demo-1.yaml
```

Configure CoreDNS rewriting and restart to reload the configuration:

```bash
kubectl apply 05-coredns-configmap-transponder-demo-1.yaml
POD=$(kubectl get pod -n istio-system -l app=istiocoredns -o jsonpath="{.items[0].metadata.name}")
kubectl delete pod $POD
```

Add `ServiceEntry`, `DestinationRule` and new Transponder scanner configuration:

```bash
kubectl apply -f 05-service-alias-transponder-demo-1.yaml
POD=$(kubectl get pod -l app=transponder-scanner -o jsonpath="{.items[0].metadata.name}")
kubectl delete pod $POD
wait 5
POD=$(kubectl get pod -l app=transponder-scanner -o jsonpath="{.items[0].metadata.name}")
kubectl logs $POD transponder-scanner
```

## 06. Service Aliases with Multi-Cluster

TODO

## Clean Up

Delete the clusters:

```bash
gcloud container clusters delete transponder-demo-1 --zone $ZONE_1 --project $PROJECT
gcloud container clusters delete transponder-demo-2 --zone $ZONE_2 --project $PROJECT
```

Remove temporary files:

```bash
rm transponder
rm 00-kubedns-config-1.yaml
rm 00-kubedns-config-2.yaml
rm 02-scanner.yaml
rm 03-transponder-server-multi-cluster.yaml
rm 04-transponder-server-multi-cluster-dns.yaml
rm 05-kube-dns-configmap-transponder-demo-1.yaml
rm 06-kube-dns-configmap-transponder-demo-2.yaml
```
