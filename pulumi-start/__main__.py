"""
Creating a Kubernetes Deployment
"""
import pulumi
import pulumi_docker as docker
from pulumi_kubernetes.apps.v1 import Deployment
from pulumi_kubernetes.core.v1 import Service
import subprocess
import pulumi_awsx as awsx
import pulumi_eks as eks
import pulumi_kubernetes as k8s

container_repository = awsx.ecr.Repository("retailstore_repository")
pulumi.export("repository_url", container_repository.url)

images = ["app-orders", "app-catalog", "app-cart", "ui-ui", "app-checkout"]

# Build and push all images
built_images = {}
for image in images:
    app_dir = "../src/" + image.split("-")[1]

    built_image = awsx.ecr.Image(
        image,
        repository_url=container_repository.url,
        context=app_dir,
        platform="linux/amd64",
        opts=pulumi.ResourceOptions(
            ignore_changes=["provider"]
        )
    )
    built_images[image] = built_image

# Minikube configuration
config = pulumi.Config()
is_minikube = config.require_bool("isMinikube")

# EKS configuration
min_cluster_size = config.get_int("minClusterSize", 1)
max_cluster_size = config.get_int("maxClusterSize", 3)
desired_cluster_size = config.get_int("desiredClusterSize", 2)
eks_node_instance_type = config.get("eksNodeInstanceType", "t3.medium")
vpc_network_cidr = config.get("vpcNetworkCidr", "10.0.0.0/16")

# Create EKS cluster and Kubernetes provider (only if not using minikube)
if not is_minikube:
    eks_vpc = awsx.ec2.Vpc("eks-vpc",
        enable_dns_hostnames=True,
        cidr_block=vpc_network_cidr)

    eks_cluster = eks.Cluster("eks-cluster",
        vpc_id=eks_vpc.vpc_id,
        authentication_mode=eks.AuthenticationMode.API,
        public_subnet_ids=eks_vpc.public_subnet_ids,
        private_subnet_ids=eks_vpc.private_subnet_ids,
        instance_type=eks_node_instance_type,
        desired_capacity=desired_cluster_size,
        min_size=min_cluster_size,
        
        max_size=max_cluster_size,
        node_associate_public_ip_address=False,
        endpoint_private_access=False,
        endpoint_public_access=True
    )
    
    # Create Kubernetes provider from EKS cluster
    k8s_provider = k8s.Provider("k8s-provider",
        kubeconfig=eks_cluster.kubeconfig)
    
    # Export the kubeconfig for EKS
    pulumi.export("kubeconfig", eks_cluster.kubeconfig)

# Deploy each microservice
deployments = {}
services = {}

# Service configurations (adjust ports based on your actual app requirements)
service_configs = {
    "app-orders": {"port": 8080, "target_port": 8080},
    "app-catalog": {"port": 8080, "target_port": 8080},
    "app-cart": {"port": 8080, "target_port": 8080},
    "app-checkout": {"port": 8080, "target_port": 8080},
    "ui-ui": {"port": 8080, "target_port": 8080},
}

for image_name in images:
    app_labels = {"app": image_name}
    
    # Set provider based on environment
    deployment_opts = pulumi.ResourceOptions(depends_on=[built_images[image_name]])
    service_opts = pulumi.ResourceOptions()
    
    if not is_minikube:
        deployment_opts = pulumi.ResourceOptions(
            provider=k8s_provider,
            depends_on=[built_images[image_name]]
        )
        service_opts = pulumi.ResourceOptions(provider=k8s_provider)
    
    # Create deployment
    deployment = Deployment(
        image_name,
        spec={
            "selector": {"match_labels": app_labels},
            "replicas": 1,
            "template": {
                "metadata": {"labels": app_labels},
                "spec": {
                    "containers": [{
                        "name": image_name,
                        "image": built_images[image_name].image_uri,
                    }]
                }
            }
        },
        opts=deployment_opts
    )
    deployments[image_name] = deployment
    
    # Create service
    service = Service(
        f"{image_name}-service",
        metadata={
            "labels": app_labels,
        },
        spec={
            "type": "NodePort" if is_minikube else "LoadBalancer",
            "ports": [{
                "port": service_configs[image_name]["port"],
                "target_port": service_configs[image_name]["target_port"],
                "protocol": "TCP"
            }],
            "selector": app_labels,
        },
        opts=service_opts
    )
    services[image_name] = service

# Export URLs for each service
if is_minikube:
    try:
        minikube_ip = subprocess.run(
            ["minikube", "ip"], capture_output=True, text=True, check=True
        ).stdout.strip()
    except Exception:
        minikube_ip = "unknown"
    
    for image_name, service in services.items():
        node_port = service.spec.apply(
            lambda spec: spec["ports"][0]["node_port"] if "ports" in spec else None
        )
        url = node_port.apply(lambda p: f"http://{minikube_ip}:{p}" if p else "unknown")
        pulumi.export(f"{image_name}_url", url)
else:
    for image_name, service in services.items():
        # Use apply() to access the Output value
        url = service.status.apply(
            lambda status: status["load_balancer"]["ingress"][0]["hostname"] 
            if status and "load_balancer" in status 
            and status["load_balancer"] 
            and "ingress" in status["load_balancer"]
            and len(status["load_balancer"]["ingress"]) > 0
            and "hostname" in status["load_balancer"]["ingress"][0]
            else None
        )
        pulumi.export(f"{image_name}_url", url)