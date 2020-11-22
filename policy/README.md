# Zeus

In the energy-aware datacenter, zeus is responsible to Policy, e.g. receiving energy signal from client, maintaining Policy.

## How to build Docker image

For binary, run:

```make zeus```

For docker image, run:

``` docker build --tag <tag> .```

You are supposed to see the built image when run:

```docker images```

To run it:

```docker run -d -it -p 8080:9999 --name zeus -v my-vol:/app <tag> --rm```

And now you can access it with `localhost:8080` on your browser

## How to deploy it to Minikube
1. Tunnel the docker-env to Minikube

`eval $(minikube -p minikube docker-env)`

2. Build the image into minikube's docker:

``` docker build --tag <tag> .```

3. Create Deployment

```kubectl create -f deployment.yml```

4. Create Service

```kubectl expose deployment zeus --type=LoadBalancer --port=9090```

5. Check out the service

```minikube service zeus```

## How to deploy it to GKE
1. configure your Docker with gcloud:

`gcloud auth configure-docker`

2. Push the image to GCR(Google Cloud Registry):

`docker push us.gcr.io/kube-flux/kube-flux-zeus:0.0.3`

3. Deploy the container of that image

`kubectl apply -f low.yaml
kubectl apply -f medium.yaml
kubectl apply -f top.yaml`

`kubectl create deployment test --image=us.gcr.io/kube-flux/kube-flux-zeus:0.0.3`

4. Expose the Service

`kubectl expose deployment zeus --name=zeus-service --type=LoadBalancer --port 80 --target-port 9999`

Now you'd see the external Ip by calling `kubectl get service`!
