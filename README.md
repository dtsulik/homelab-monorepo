# gif-doggo

Microservice to generate a gif from series of images. It is based on [workshop code here](https://github.com/dtsulik/workshop-from-idea-to-mvp).

This one has more to it:
- Tracing
- ApiGW (this is here to maybe in future add A/B deployments)
- Structured Logging
- Metrics

## service architecture

Idea is to expand on the services from the workshop and add apigw to it add more cloud agnostic spin to it. (The workshop stack was aws native).

## CICD

CICD is done with Dagger in Golang. Deployment will be done with either talking directly to k8s api to ArgoCD api. Since this is a monorepo we can forgo the usual gitops repo and have the cicd code talk to relevant deployment API. (Still not as clean as GitOps, needs more investigation).

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
- `apigw` service should be named bff (backend for frontend)
- errors need to be more structured and relevant types need custom errors/stringers

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

