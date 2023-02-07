# Backdoor

It is a very simple service meant to be used as a web hook for deploying something via web hooks from CI in a *TRUSTED ENVIRONMENT*.
It has only basic security built in, so it should not be used in anything you are not willing to sacrifice.

## Use case
You have CI job that builds image "username/service" and pushes it to some registry. After push is done, CI fires up curl like this:

```shell
curl https://your.url/deploy/deploymentname?secret=NotSoSecretAtAll
```

Which *hopefully* pulls and deploys your newly built service.

## Configuration

You need to configure at least one deployment for service to be usable. Deployment is specified in configuration file as an object with a name and two properties (script name to be executed and a secret to authorize deployment).

Here is an example of `config.json` file:

```json
{
  "deploymentname": {
    "secret": "NotSoSecretAtAll",
    "script": "deploy.sh"
  }
}
```

When started service will read this config and and begin listening on port 8080 for requests. It will handle only requests to `/deploy/{deployment}` paths, where `{deployment}` is one of the deployment names configured in `config.json`. In our example above, the only working url would be `/deploy/deploymentname`. Then service will check secret token you passed as url value to match the one configured.
If it matches, then the configured script will be run. It should exist in `./scripts` directory (relative to service working dir) and be executable.
If script's return code is zero, service will return HTTP code 200, otherwise code will be 500.

