## v1.5.0

### Additions

+ Deployment resource added
+ Base CR Status on Conditions as it is the standard way to adopt for managing our own Status data, (see here)[https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties])
+ Status Type managed by OKT is "ReconciliationSuccess"

### Changes

+ Align dependancies with OperatorSDK 1.5.0
+ Side effects of the upgrade `sigs.k8s.io/controller-runtime` from `v0.6.2` to [`v0.7.2`](https://github.com/kubernetes-sigs/controller-runtime/releases/tag/v0.7.2). This version comes with the Object abstraction we missed to group together Runtime (serializable) and Metadata (indentifiable) interfaces (any object which you could write as YAML or JSON, and then `kubectl create`). 
  + The OKT's Kube client creation is no longer done with a Runtime object but with the new sig's runtime Client.Object
  + Reconcile func signature has changed and get the context as first argument (from `Reconcile(ctrl.Request)` to `Reconcile(context.Context, ctrl.Request)`)
  + OKT MutableResource (okt/resources resources.MutableResource) PostMutate() signature changed to take for the CR parameter a client.Object and no longer a runtime.Object
+ Define a dummy Client for tests instead of the dummy oktclient (allows to go further in tests) 
+ OKTStatus structure is removed since we now use standard Status Conditions
+ API break out: 
  + the functions oktresource.ApplyTpl() and oktresource.ApplyGOStruct() are renamed respectively oktresource.CopyTpl() and oktresource.CopyGOStruct()
+ The gen-resource CLI tool generates now a stub resource base on a dictionary definition to determine the resoruce Kind and Version. The stub is generated localy closed to the resource mutator object. This bring more flexibility to handle different kind/version resource and simplify OKT. 

### Bug fixes

+ Fix a naming issue in okt-gen-resource CLI tool
