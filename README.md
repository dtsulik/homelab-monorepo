# Homelab mono repo

Monorepo for pet projects and personal lab

Services in the project directory are mostly are [rube-goldberg machines](https://en.wikipedia.org/wiki/Rube_Goldberg_machine). They are there only to generate some noise for telemetry or for me to learn various design things (DDD, modular monoliths, event driven architectures, etc).

## CICD

CI is done in Golang most it is done with dagger.io, but I am looking to replace it. CD is done with gitops (ArgoCD).

## TODO
- All services are missing structured reponse
- Add semantic release to dagger code, now it's just pushin to `latest`
- All services are missing metrics (done for one service)
- All services are missing tracing (done to some extent)
- All services need proper models and proper use of interfaces
- Code has mix of snake/camel case, needs to be unified
- Where tests? 
- Srsly where tests?
- Fuzzing?
- `apigw` service should be named bff (done - now helm chart)
- errors need to be more structured and relevant types need custom errors/stringers
- project creation in CI and builder are golang specific

### TODO release specific todos
- add semantic release
- separate build and publish stages
- add helm chart release process

![prod](.docs/test-in-prod.jpg)

## MISC

```bash
# upload data
curl -v --data-binary "@pun.jpg" -H "filename: pun.jpg" -X POST https://k8s-doggo.local/upload

# request
curl -v -X POST \
    -H "Content-Type: application/json" \
    --data '{"images": ["doge.jpg", "pun.jpg"], "delays": [100,100], "output":"doggo.gif"}' \
    https://k8s-doggo.local/request
```

# K8S

```bash
cd build/deployment/cluster
kind create cluster --config=kind.yaml
```

## ALB annotations
```yaml
  annotations:
    alb.ingress.kubernetes.io/scheme: internal
    alb.ingress.kubernetes.io/target-type: instance
    alb.ingress.kubernetes.io/load-balancer-name: 
    alb.ingress.kubernetes.io/backend-protocol: HTTP
    alb.ingress.kubernetes.io/listen-ports: '[{"HTTP":80}]'
```



Gitops repo for homelab, used by ArgoCD