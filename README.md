# Pod Network Annotator

The podnetworkannotator applies annotations to the running Pod with its public
network values.

The annotations are:
 - `netdiscover.cycore.io/public_ipv4`
 - `netdiscover.cycore.io/public_ipv6`
 - `netdiscover.cycore.io/public_hostname`

Required environment variables are:
  - `POD_NAMESPACE` - the namespace of the Pod
  - `POD_NAME` - the name of the Pod

Both of these are available from the downward API of Kubernetes.

Optional environment variables are:
  - `CLOUD` - the cloud environment in which the Pod is running.  See
    [CyCoreSystems/netdiscover](github.com/CyCoreSystems/netdiscover) for details about supported cloud
    environments.

The network data is run at start and once every minute thereafter, so it will
pick up any changes to the Pod's public network information over time.

