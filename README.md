# Restarter

A simple program that watches a kubernetes deployment and triggers a restart
when more than x% of pods are not in the `ready` state for a given period of time.

## Usage

```shell
restarter --namespace default --name bongo --interval 1m --restart-when 75 --restart-after 5m
```

Running this will watch the deployment `bongo` in the `default` namespace.
It will check the percentage of pods that are "ready" every minute and
trigger a restart if more than 75% of them are not ready for 5 consecutive minutes.
