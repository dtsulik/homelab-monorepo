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
- All services are missing metrics
- All services are missing tracing

## MISC

```bash
# upload data
curl -v --data-binary "@pun.jpg" -H "filename: pun.jpg" -X POST https://apigw.local/upload

# request
curl -v -X POST \
    -H "Content-Type: application/json" \
    --data '{"images": ["doge.jpg", "pun.jpg"], "delays": [100,100], "output":"doggo.gif"}' \
    https://apigw.local/request
```



