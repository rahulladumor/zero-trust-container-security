package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/eks"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// EKS Cluster for container security platform
		cluster, err := eks.NewCluster(ctx, "security-cluster", &eks.ClusterArgs{
			Version: pulumi.String("1.28"),
			VpcConfig: &eks.ClusterVpcConfigArgs{
				SubnetIds: pulumi.StringArray{
					pulumi.String("subnet-xxx"), // Replace with actual subnet IDs
				},
			},
		})
		if err != nil {
			return err
		}

		// Kubernetes provider
		k8sProvider, err := kubernetes.NewProvider(ctx, "k8s", &kubernetes.ProviderArgs{
			Kubeconfig: cluster.KubeconfigJson,
		})
		if err != nil {
			return err
		}

		// Falco for runtime security
		_, err = kubernetes.NewHelm(ctx, "falco", &kubernetes.HelmArgs{
			Chart:     pulumi.String("falco"),
			Version:   pulumi.String("3.8.0"),
			Namespace: pulumi.String("falco"),
			FetchOpts: kubernetes.HelmFetchOptsArgs{
				Repo: pulumi.String("https://falcosecurity.github.io/charts"),
			},
		}, pulumi.Provider(k8sProvider))
		if err != nil {
			return err
		}

		// Trivy Operator for image scanning
		_, err = kubernetes.NewHelm(ctx, "trivy", &kubernetes.HelmArgs{
			Chart:     pulumi.String("trivy-operator"),
			Version:   pulumi.String("0.18.0"),
			Namespace: pulumi.String("trivy-system"),
			FetchOpts: kubernetes.HelmFetchOptsArgs{
				Repo: pulumi.String("https://aquasecurity.github.io/helm-charts/"),
			},
		}, pulumi.Provider(k8sProvider))
		if err != nil {
			return err
		}

		ctx.Export("clusterName", cluster.Name)
		return nil
	})
}
