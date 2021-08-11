# Reconciliation in K8S

## K8S resource objects

Resource objects typically have 3 components (ObjectMeta, ResourceSpec, ResourceStatus).

Some of them are modified either by the user or by the K8S (server) or both.

![res-objects](doc/K8SResourceObject.PNG)

## Object mutation

![objects-mutations](doc/ObjectsMutation.png)

## The reconciliation loop
The reconciliation of K8S resources is done thought an idempotent process that tends to converge to the expected state.

Each GO reconciliation function is part of a loop managed by a Controller watching for events of resource modification at the K8S cluster level.

It is not unusual to terminate the reconciliation function with an error while the current state of a resource is viewed through a Cache.

![reconcile-loop](doc/reconcile-loop.png)

To answer the question of the Application's lifecycle and implement an operator with a capability level going beyond the Phase 2, OKT provides some solutions which should help to manage the application as a K8S resource and then to handle its reconciliation with a similar (and idempotent) process as the one we use for the K8S resources.



