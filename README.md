# Selfservice

Run kubernetes workloads through user friendly forms.

## Why does this exist

This project puts version controlled html forms and markdown docs between a user and the workloads in a kubernetes cluster.

![selfservice](https://user-images.githubusercontent.com/849403/219342375-d6798267-4eee-4c5b-b877-e163dbf012cf.jpg)

## Install

```bash
kubectl create namespace selfservice
kubectl apply -n selfservice -f https://raw.githubusercontent.com/infor-design/selfservice/main/manifests/install.yaml
```

## Port Forwarding

```bash
kubectl port-forward svc/selfservice-server -n selfservice 8080:8080
```

## Ingress

```yaml
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: selfservice-server-ingress
  namespace: selfservice
  annotations:
    ingress.kubernetes.io/app-root: "/"
    kubernetes.io/ingress.class: nginx
spec:
  rules:
    - host: domain.com
      http:
        paths:
          - backend:
              service:
                name: selfservice-server
                port:
                  number: 8080
            path: "/"
            pathType: ImplementationSpecific
```
