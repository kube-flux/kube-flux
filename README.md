# Kube-flux
This document shows you how to re-build the whole project.
## 1. Terraform:
Our system runs on Kubernetes cluster, the first step is to deploy a cluster on Google Kubernetes Engine.

Terraform is an open-source infrastructure as code software tool to provision data center infrastructure.
For installation, please refer to [install-guide](https://learn.hashicorp.com/tutorials/terraform/install-cli)
+ `terraform/cluster/` directory contains the terraform code to provision a GKE cluster.
+ Create a service account under `IAM & Admin` for our GCP project if not already created.
+ Fetch the secret-key and store it as `tf-key.json` under `terraform/cluster/`
#### Commands:
+ `terraform init` 
+ `terraform plan` 
+ `terraform apply`
## 2. Authenticate Kubectl with GKE:
+ Easiest way to authenticate is using `gcloud`. For `gcloud` installation, please refer to [install-guide](https://cloud.google.com/sdk/docs/install)
+ Under GKE cluster, click on the connect button and copy-past `gcloud <command>` and run it in your local terminal.
+ `kubectl config current-context` to cross-reference to currently deployed GKE cluster.
+ `kubectl get pods -A` to display the pods as a sanity check.
## Authenticate Go-client:
+ Under GKE cluster, copy & paste the certificate file and store it as `ca.pem` under `prod/`
## Kubectl:
Deploying apps via kubectl:
+ `app/` directory contains `deployment.yaml`, `service.yaml`, `ingress.yaml`.
#### Commands:
+ `kubectl apply -f deployment.yaml`
+ `kubectl apply -f service.yaml`
+ `kubectl apply -f ingress.yaml`

# Deploy Zeus

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

3. Deploy the container

`kubectl apply -f low.yaml
kubectl apply -f medium.yaml
kubectl apply -f top.yaml`

`kubectl create deployment test --image=us.gcr.io/kube-flux/kube-flux-zeus:0.0.3`

4. Expose the Service

`kubectl expose deployment zeus --name=zeus-service --type=LoadBalancer --port 80 --target-port 9999`

Now you'd see the external Ip by calling `kubectl get service`!

## Starting up the UI
1. Enter the frontend directory:

`cd frontend`

2. Install dependencies:

`npm install`

3. Run app and open http://localhost:9000 to view it in the browser:

`npm start`
