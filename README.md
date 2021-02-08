# Resource Replication Operator


This operator is designed to copy resources from one namespace to
another.  Sometimes there is a single secret like a certificate key
pair that you want to copy to many namespaces.  This operator allows
you to have a single secret that exists in one namespace.  And upon
declaring you want a copy in another namespace, make a copy there.

``` yaml
apiVersion: utils.simopolis.xyz/v1alpha1
kind: ReplicatedResource
metadata:
  name: test-rr
spec:
  source:
    namespace: certificates
    kind: Secret
    name: my-secret
```
