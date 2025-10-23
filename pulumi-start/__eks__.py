import pulumi
import pulumi_awsx as awsx
import pulumi_eks as eks

config = pulumi.Config()
min_cluster_size = config.get_int("minClusterSize", 3)
max_cluster_size = config.get_int("maxClusterSize", 6)
desired_cluster_size = config.get_int("desiredClusterSize", 3)
eks_node_instance_type = config.get("eksNodeInstanceType", "t3.medium")
vpc_network_cidr = config.get("vpcNetworkCidr", "10.0.0.0/16")

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

# Export values to use elsewhere
pulumi.export("kubeconfig", eks_cluster.kubeconfig)
pulumi.export("vpcId", eks_vpc.vpc_id)
