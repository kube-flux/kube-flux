# Zeus

In the energy-aware datacenter, zeus is responsible to Policy, e.g. receiving energy signal from client, maintaining Policy.

# How to build it

For binary, run:

```make zeus```

For docker image, run:

``` docker build --tag kube-flux/zeus:0.0.1 .```

You are supposed to see the built image when run:

```docker images```

To run it:

```docker run -d -it -p 8080:9999 --name zeus -v my-vol:/app kube-flux/zeus:0.0.1 --rm```