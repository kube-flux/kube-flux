# Kube-flux
## Terraform:
Terraform is an open-source infrastructure as code software tool to provision data center infrastructure.
For installation, please refer to [install-guide](https://learn.hashicorp.com/tutorials/terraform/install-cli)
+ `terraform/cluster/` directory contains the terraform code to provision a GKE cluster.
+ Create a service account under `IAM & Admin` for our GCP project if not already created.
+ Fetch the secret-key and store it as `tf-key.json` under `terraform/cluster/`
#### Commands:
+ `terraform init` 
+ `terraform plan` 
+ `terraform apply`
## Authenticate Kubectl with GKE:
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
