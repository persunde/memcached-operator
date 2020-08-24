## Files to look at:

How it works:
An Operator watches one or more CustomResource(s) (CR). The job of the Operator is to make sure that the state of the Cluster, matches what is defined in the CR.
So if a CR says there should be 12 Replicas of Deployment A, then the Operator will take the appropriate actions to make sure the Cluster has 12 replicas/Pods of A.
A CR contains some data that describes the desired state of some parts of the cluster, as well as metadata, logs etc. For example, it can contain the reason for why something failed on the last try, like a Node being unresponsive and failing the health check, and therefore the operation failed. Such information can be used when the Operator will retry to Reconcile() again after a backoff period if it failed the last time.

An Operator watches for changes in the CRs. So whenever there is a change in CR the Operator watches, it will trigger an event and the Reconcile() function will be called, so that the Operator will take the appropriate actions to make sure the cluster is in a state as described by the CR.

! IMPORTANT NOTE:
The CR must ALWAYS be applied on the cluster, before an Operator can be run. As the Operator depends on and expect that the CR(s) it will watch are already in the cluster.

NOTE:
Normally any updates to the CR will cause a new Reconcice() event, but you can update the CR without causing a new event. You can see this in both the *_controller.go files. Since they use:
```go
ctx := context.Background()
```
By using context.Backgrund() any updates to the CR will not cause a new Reconcile() event, unless you force it by returning an error or explicitlty returning with Requeue: True like this:
```go
return ctrl.Result{Requeue: true}, nil
```
This is normally used when updating metadata status that does not directly impact the state of the cluster. Such as what the latest latency is, or why something failed or if it succeded. This is nice info to have, both for the Operator, but also if a person want to manually check the logs and status regarding the Operator and the objects it watches.

NOTE:
Unless I am not clear enough. A CR only defines a small part of the state of the whole cluster (eg. the number of replicas in a given deployment, or the image version of a Pod/Deployment). So the Cluster can be in many different states, but still be in a state that is in correspondence with the given CR.

### 1. main.go

This is the entry point of the Operator. The starting point.
It defines the metadata.
Most importantly it binds the functionality to this operator. In this case it uses "SetupWithManager()" function to watch for changes in one CustomResource (CR). 
It calls SetupWithManager() two times, since this Operator watches two different CustomResouce's, and then STARTS the Operator at the end of the main.go file.

Things to note:
SyncPeriod: 
Normally an Operator would only be activated once a change happens to one of the CRs it watches. But since we want to periodically ping/poll a webserver,
we set a SyncPeriod of X seconds. By default, this is set to once every 10 day or so.

Namspace:
If the namespace is empty, then it defaults to the default namespace, in this case that is "default".
You can limit the Operator to a namespace or it could work across all namespaces.
For example, if you limit the namespace(s) it can read from, then it will only get events from CRs in that namespace.

### 2. memcached_controller.go && webserver_controller.go

In all controllers, the Reconcile() function is the main one.
All Operators, no matter the language, tries to reconcile the Kubernetes Cluster with the CustomResource it watches.
So whenever a change happens in the CR, a Reconcile() event will be triggered, and the Operator will try and change (part of) the state of the Cluster to be as defined in the CR.

In both files, the Reconcile() function is the entry point, the main controll-loop.
The function SetupWithManager() is used by main.go to bind this controller to the Operator.
All other functions in the files are helper functions for Reconcile().

#### Flow of Reconcile() in memcached_controller.go:
1. Get/fetch the memchached CustomResource from the cluster and put the data into the `memcached` object
2. Check if the memchached Deployment exists
    * If not create a new one and update the CR to cause a new event and return
    * You can see the Deployment configuration defined in function deploymentForMemcached(), you could just as well fetch the definition from a yaml file as well.
3. Ensure the deployment replica count is the same as defined in the spec of the CR
    * If not, update the Deployment with the correct replica count and return
4. Get a list of the pods for this CRs deployment and update the CR's value "status.Nodes" with the list of pod names if it diffes from the current one
5. Return successfully

#### Flow of Reconcile() in webserver_controller.go:
The flow of webserver_controller.go is similar.
1. Get/fetch the webserver CustomResource from the cluster and put the data into the `webserver` object
2. Check if the webserver Deployment exists
    * If not create a new one and update the CR to cause a new event and return
    * You can see the Deployment configuration defined in function deploymentForWebserver(), you could just as well fetch the definition from a yaml file as well.
3. Check the latency from pinging one arbirary Pod by using the Ingress
    * If the latency is to big increase the number of replicas
    * If the latency is to low, lower the number of replicas
4. Update the CR with the latest latency

Things to note:
It uses comments above functions to say what access the Reconcile() function has, like this:

``` go
// +kubebuilder:rbac:groups=cache.example.com,resources=memcacheds,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cache.example.com,resources=memcacheds/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

func (r *MemcachedReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
    ...
}
```


### The Memcached and Webserver CR, where are they defined?
Both CR must be defined in the Code as well as in a yaml file.

IMPORTANT NOTE:
The CR must ALWAYS be applied on the cluster, before an Operator can be run. Therefore it is wise to bundle an Operator with a Chart or similar, so that the CR is applied correctly when a user of it just want to install and run the Operator. With a chart, you just have to run one command, and it will install everything you need on your cluster. If a user instead does all the steps manually, many people will skip some neccesary steps or do them in the wrong order and things will not work.

The data to be described in a CR must also be reflected in the code.
You can find the yaml to be applied to the cluster here:
```
config/crd/bases/cache.example.com_webservers.yaml
```

You can find the CRD definition in golang code here:
```
api/v1alpha1/memcached_types.go
```

By running this command, you will generate the `api/v1alpha1/zz_generated.deepcopy.go` file and the `config/crd/bases/cache.example.com_memcacheds.yaml` file
``` bash
make generate
make manifests
```

You have to manually modify/create the CR yaml files yourselves, they are located here:
```
config/samples/cache_v1alpha_*.yaml
```
Edit it with the values you want them to have, and you can apply them to the Cluster.
Remember, the CRDs define which values are possible (and required?) to have in your CRs. A CRD is like a new K8S "Kind", just like Deployment, Replicaset, HorizontalPodAutoscaler.
And then the CR is an implementation of the CRD, with some spesific values. So a CRD is ONLY a meta object, it describes what kinds you can create in your cluster, while a CR is an actual object in your cluster with some spesific values.
    - You can only be one CRD of a given type, but you can have multiple CR of that CRD kind.
    - Eg: There is only ONE Kind called "Deployment", but you can have multiple deployments in your cluster with different values.

So now your yaml files and golang *_types.go files should match up.


### Java Operator
It is significantly easier to read the code than the golang Operator.
This is because most of the legwork is done by the [java-operator-sdk](https://github.com/ContainerSolutions/java-operator-sdk).
But then again, it is not as versetile and easy to modify compared to writing in Golang.

The `Runner.java` file is the main entry point. 
It starts the Operator.
It watches for a CR called CustomService:
```java
CustomServiceController controller = new CustomServiceController(client);
operator.registerControllerForAllNamespaces(controller, retry);
```
Then it makes a new thread to do the pinging of webservers in the cluster.
    NOTE: Soon the Java-Operator-sdk will have support for looping based on a timeout duration, but it does not have that at the moment, so I made a fast and shitty implementation of it.
At the end, the Operator has a health endpoint that loops indefinitely until the program is forcefully stopped.


The `CustomServiceController.java` file contains the main controller logic. In the file the function `createOrUpdateResource()` is the main controll-loop, and the function that is called when a reconcile event is triggered by a change in the CRs it watches.
It sets the status "AreWeGood" in the CR
```java
status.setAreWeGood("Yes!");
resource.setStatus(status);
```
Then it updates the replica count based on the latency of the webservers. This is done within function 
```java
updateDeploymentReplicaCount(resourceSize);
```

At the end it creates the deployment if it does not already exist, by calling the helper function 
```java
createOrReplaceDeployment();
```


In the forked out thread it runs in an infinite loop the function 
```java
public void checkStatus() {
    ...
}
```
In this function it does these things in order:
1. Fetches the current data in the CR from K8S
2. Gets the latency from one of the webservers
3. Increases or Lowers the CRs "size" value if the latency is to big/low, otherwise no action
    * If it changes the "size" value up or down, it will update the CR with the new desired size aka the desired replica count for the corresponding deployment -> this triggers an event and `createOrUpdateResource()` will be called by the other thread and it will try and reconcile the cluster with the new desired state as defined in the CR. So the operator will here in `createOrUpdateResource()` change the replica count in the deployment to match the "size" value found in the CR.