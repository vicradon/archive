"""
Creating a Kubernetes Deployment
"""
import pulumi
import pulumi_docker as docker
from pulumi_kubernetes.apps.v1 import Deployment
from pulumi_kubernetes.core.v1 import Service
import subprocess

# Minikube does not implement services of type `LoadBalancer`; require the user to specify if we're
# running on minikube, and if so, create only services of type NodePort (not cluster ip lol).
config = pulumi.Config()
is_minikube = config.require_bool("isMinikube")


registry_prefix = config.get("registryPrefix") or "localhost:5000"

registry = docker.Container(
    "local-registry",
    image="registry:2",
    ports=[{"internal": 5000, "external": 5000}],
    restart="always",
)



images = ["app-ui", "app-orders", "app-catalog", "app-cart", "ui-ui", "app-checkout"]

pushed_images = []
for img in images:
    full_tag = f"{registry_prefix}/{img}:latest"
    pushed = docker.Tag(
        f"{img}-push",
        source_image=f"{img}:latest",
        target_image=full_tag,
        opts=pulumi.ResourceOptions(depends_on=[registry]),
    )
    pushed_images.append(pushed)

pulumi.export("pushed_images", [p.target_image for p in pushed_images])

app_name = "app-ui"
app_labels = { "app": app_name }

deployment = Deployment(
    app_name,
    spec={
        "selector": { "match_labels": app_labels },
        "replicas": 1,
        "template": {
            "metadata": { "labels": app_labels },
            "spec": { "containers": [{ "name": app_name, "image": f"vicradon/timeapi" }] }
        }
    })

# Allocate an IP to the Deployment.
frontend = Service(
    app_name,
    metadata={
        "labels": deployment.spec["template"]["metadata"]["labels"],
    },
    spec={
        "type": "NodePort" if is_minikube else "LoadBalancer",
        # port is 2500, services within kubernetes use it. target port is 4500, the app runs there
        "ports": [{ "port": 2500, "target_port": 4500, "protocol": "TCP" }],
        "selector": app_labels,
    })

# When "done", this will print the public IP.
result = None
if is_minikube:
    result = frontend.spec.apply(lambda v: v["cluster_ip"] if "cluster_ip" in v else None)

    try:
        minikube_ip = subprocess.run(
            ["minikube", "ip"], capture_output=True, text=True, check=True
        ).stdout.strip()
    except Exception:
        minikube_ip = "unknown"

    # Extract nodePort dynamically
    node_port = frontend.spec.apply(
        lambda spec: spec["ports"][0]["node_port"] if "ports" in spec else None
    )

    # Build full URL like: http://192.168.49.2:31501
    url = node_port.apply(lambda p: f"http://{minikube_ip}:{p}" if p else "unknown")
    pulumi.export("url", url)
else:
    ingress = frontend.status.load_balancer.apply(lambda v: v["ingress"][0] if "ingress" in v else "output<string>")
    result = ingress.apply(lambda v: v["ip"] if v and "ip" in v else (v["hostname"] if v and "hostname" in v else "output<string>"))

pulumi.export("ip", result)
