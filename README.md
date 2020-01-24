## NOTE

This repo is no longer being maintained. Users are welcome to fork it, but we make no warranty of its functionality.

## Kubernetes Cluster Deployment on CenturyLink Cloud

[![](https://badge.imagelayers.io/centurylink/kube-cluster-deploy.svg)](https://imagelayers.io/?images=centurylink/kube-cluster-deploy:latest 'Get your own badge on imagelayers.io')

This image is for use in [Panamax](http://panamax.io) to create a cluster in CLC using the [Dray](https://registry.hub.docker.com/u/centurylink/dray/) tool. It takes environment variables related to one's CLC account and creates a cluster based on your sepecifications/values you pass. It DOES NOT install kubernetes if ran standalone. It only creates the nodes in order to then install kubernetes. The output of the ran container are the IPs of the created nodes: Master and any Minions specified.

Environment Variables include:
* USERNAME
* PASSWORD
* CPU
* MEMORY_GB
* GROUP_ID
* MINION_COUNT

### Standalone Usage
`docker run -d --name=cluster -e "USERNAME=username" -e "PASSWORD=password" -e "CPU=2" -e "MEMORY_GB=2" -e "GROUP_ID=wa1-1234" -e "MINION_COUNT=2" centurylink/kube-cluster-deploy`
