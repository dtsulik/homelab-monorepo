# gif-doggo

Microservice to generate a gif from series of images. It is based on [workshop code here](https://github.com/dtsulik/workshop-from-idea-to-mvp).

This one has more to it:
- Tracing
- ApiGW (this is here to maybe in future add A/B deployments)
- Structured Logging
- Metrics

## service architecture

Idea is to expand on the services from the workshop and add apigw to it add more cloud agnostic spin to it. (The workshop stack was aws native).