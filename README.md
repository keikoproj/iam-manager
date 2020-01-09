# iam-manager
AWS IAM role management for K8s namespaces inside cluster using CRD(Custom Resource Definitions)


### Custom Resource Controller managing the IAM Role life cycle

#### Abstract:
This module can help other organizations who are looking for namespace IAM role management with the enough security boundaries defined around the solution. The idea of this approach is to build custom resource controller which can securely manage IAM role management independently i n kubernetes environment.


#### Design:

Here is the high level design diagram

![Arch](docs/images/IAM_CRD_DESIGN.jpeg)

