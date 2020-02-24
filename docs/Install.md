##Installation:

Simplest way to install iam-manager along with the role required for it to do the job is to run [install.sh](hack/install.sh) command.  

Update the allowed policies in [allowed_policies.txt](hack/allowed_policies.txt) and config map properties [config_map](hack/iammanager.keikoproj.io_iamroles-configmap.yaml) as per your environment before you run install.sh.

Note: You must be cluster admin and have exported KUBECONFIG and also has Administrator access to underlying AWS account and have the credentials exported.

example:
```bash
export KUBECONFIG=/Users/myhome/.kube/admin@eks-dev2-k8s  
export AWS_PROFILE=admin_123456789012_account
./install.sh [cluster_name] [aws_region] [aws_profile]
./install.sh eks-dev2-k8s us-west-2 aws_profile
```

#### Enable Webhook?
iam-manager uses Dynamic Admission control (admission webhooks) to validate the requests against the whitelisted policies and rejects the requests before it gets inserted into etcd. This is the cleanest approach to avoid any more invalid/junk data into etcd.  
To enable webhooks,
1. You must be completed the installation section before you proceed further.
2. You must have [cert-manager](https://cert-manager.io/docs/) installed on the cluster to  manage the certificates.  
    ```kubectl apply -f cert-manager/cert-manager-v0.12.0.yaml --validate=false```
3. Apply webhook spec
    ```kubectl apply -f hack/iam-manager_with_webhook.yaml```
4. Update the "webhook.enabled" property in config map to true.

##### iam-manager with kiam
This installation is where pods can assume the role only via kiam. Kiam server runs on master nodes and any role created must be trusted by master node instance profile to be assumed by kiam.  

example:
```bash
export KUBECONFIG=/Users/myhome/.kube/admin@eks-dev2-k8s  
export AWS_PROFILE=admin_123456789012_account
./update_with_kiam.sh [cluster_name] [aws_region] [aws_profile] [masters_nodes_instance_profile]
./update_with_kiam.sh eks-dev2-k8s us-west-2 aws_profile arn:aws:iam::123456789012:role/masters.eks-dev2-k8s

```

