# OKT State Machine


## Principle

An application can take several states and exposing the current state it has during a run is like pointing the location of a mobile object on a map. 

Having a Graph to describe what is managed by the K8S Operator should be helpful for a human operator or external components in charge to bring some "observability" features to a K8S operator.

Traversing a Graph has also the advantage to maintain a consistent path of all traversed nodes. This can help too to better understand or investigate on what is happening during a run.

Offering a developement framework to build this graph in a consistent way for a K8S operator should allow to normalize the view on the application's lifecycle as long as it is operated by this operator implemented with this framework (using a standard like CNCF CloudEvent could enforce this normalization and interoperability).

A Graph contains nodes and leaf nodes. Evolving between nodes is constraint by defined transitions to pass on known verified events.

The principle is not to traverse all the graph during 1 reconciliation but, at each reconciliation, to:
  + get the current state of the application (a database or whatever application), 
  + collect events that happened, and generate, if needed, the event telling to go to the state accordingly to the expected state in CR (add it to the collected events),
  + trigger (or not) a next state in regards to the collected events
  + in case of new state change, perform asynchroniously the actions related to this state
  + if the expected state is not the current state, requeue a controller runtime event for further Reconiliation...

The proposition is to implement an OKT resource to manage the application life cycle like we have resource of different kind in Kubernetes: 
+ The application's state is evolving through the multiple Reconciliations. 
+ The expected state is specified in the CR, 
+ The current state is got from a specific Client call to pick up the information and map it into the corresponding state representation.

It based on:
+ A state machine based on what is offered by the OKT's GO module `tools/statemachine`
+ A Client implementing OKT's Client interface (CRUD) that communicate with the application 
+ An OKT application resource that can be registered by the OKT Reconciler


## Purpose

This utility allows to build a state machine from the base of a Live Cycle Graph you provide and that fit your application's needs.

A graph is a list of oriented nodes with 1 or more leafs. 

1 action to do is attached to each node. During the action 1 or more events can be triggered and provided to the statemachine as an events list. 

What is an event ? It is simply the name of the node to reach triggered at a moment in a specific context. 

The State Machine  will decide to go to the next possible node by testing sequentially each event from the current state. 




## One thing to know

A special state exists. It is `DefaultState` of type LCGSTATE as defined in the OKT state machine module. 

It is used as a facility to designate the Next step without knowing its name during the execution of a node action.

So a LCGGraph can be browsed through a "normal" path or branch (from start to the end always going to the "Next" node) and there's some "debranching" events that will deroute from the normal path to go to antoher branch of the LCGraph. These "debranching" events are called "Triggers" and are specified at each node description when they exists (debranch on ErrorManagement for example).

## LC Graph types

LCGState - A state node

LCGGraph - A graph (see below) with state nodes. Each state node can have several children. The children list is ordered. This allows to 

LCGEvents - Events list generated during a the application lifecycle. Envents are named exactly as a state node. The event "2" (i.e. itoa(Running)) is the event raised to "go-to-running" state. Actually an event is a LCGState.

LCGChildren - A state node children list. The list is ordered by priority. The child with the higher priority will be triggered first if an event exist. 

Note that Leaf nodes and the graph's entry node, have NO action attached to them. 

Example:

```
const (
	Start oktsm.LCGState = iota  /// Is equal to 0
	Running     // Is equal to 1
	Servicing
	Stopping
	End         // The state End is a uniq leaf node of this Graph (described below)
)

var graph = oktsm.LCGGraph{
	Start: oktsm.LCGChildren{
		Running,
	},
	Running: oktsm.LCGChildren{
        Servicing,  // It has the higher priority in case of dilemne between Servicing or Stopping or DefaultState events
        Stopping,   // The last child is the default node (raised by both Stopping or DefaultState events) 
	},
	Servicing: oktsm.LCGChildren{
		Running,
 		End,  // Default
	},
	Stopping: oktsm.LCGChildren{
 		End,  // Default
	},
    End: oktsm.LCGChildren{},  // A leaf node without any action, that will close the statemachine run...
}
```

## The story behind this implementation (/!\ not yet completed at this time)

Now, right after diving, with Story 1, into a "simple" implementation, I have to go further in the Operator's capability level and especially, I have to handle a way to treat the different "States" my application (a database for example or any application) will going through. 
For example, beyond the resource infrastucture management seen previously, I want now to deal with the fact that my database life is traversing some specific states as follow:

- start - the database is being started but not yet available (when this action is completed a "go-to-running" event is generated)
- running - now the database is ready to accept client connections (It is a stable state while no "go-to-servicing" nor "go-to-stopping" events are raised)
- servicing - a service operation is in progress (a backup, a configuration change) that can affect user experience. Once done, a "go-to-running" or "go-to-end" can be generated
- stopping - the database will stop its service, all client must disconnect
- ended - the service is no longer available

For these steps, I wish an easy way to manage them thanks to change in my CR, and I'd like to have the CR status updated as well while they occurs.
However, these steps are happening at the application level, not at infrastucture level (actually not completely, as we can imagine some dependancies between both).
Here we are plenty in the need to drive the application lifecycle through my operator. But how will we manage that ?

In Story 1 we described a Reconciliation cycle triggered at each event and trying to traverse a list of steps (a branch) as follow :

    CRChecker->ObjectsCreator->Mutator->Updator->ManageSuccess  (+ 2 "debranching" steps to ManageError & CRFinalizer)
Going from 1 step to the other is conditionned by the success of all actions taken during the step. Else we debranch to the `ManageError` step. All of this happen during **1 Reconciliation cycle**.

For my application lifecycle, I have 1 graph (name it **App LC Graph**) of steps representing the applications states I want to manage. At each step some actions have to be done, that may take a while: 

     Start->Running<->Servicing -> End
                    ->Stopping  -> End
                   
Going from 1 step to the other is conditionned by some conditions that may be met  **over N Reconciliation cycles**.

I like the idea to have a clear view on the steps I defined previously, so I'll complete my work with the OperatorSDK and the OKT addon.

OKT comes with a statemachine feature that should help in defining these steps and let me focus on the code I need to implement at each step. 
To allow this, OKT provides:
  -  a sidecar for my application to help me to get my database status and launch actions on it asynchronously.
  -  an utility to modelize my graph of appication states into my CRD
  -  a GO type to implement this graph and transition rules that condition how I validate the transition from one step to another

In my CR I set the wished state (i.e. Servicing) I want to reach, while the current application state (i.e. is maintained in the CR status with a new Condition).

Once the application added to the OKT registry (like any other resource), the OKT Reconciler knows that it has to manage this  resource as follow:

  - on Start: Create() it!
  - on End: Delete() it!
  - on any other state: Update it!

As any other resource, it put in place an idempotent mecanism and detect changes (and thus will do nothing during a Reconciliation if there's nothing new). Here what will trigger a change:

  - a state change (in App LC Graph) due to a CR modification
  - a state change from the observation of a change at the application level. This observability has to be implemented by the application sidecar. 
 
 A state change (in the App LC  Graph) is handled asynchronously to not impact the Controller with a too long task. On such case (long task) 1 or more requeueing orders are left to wait for the observable change once done. 

It also maintain a Status condition in the CR that reflect the application current state and errors if any.

To sum up:
  - an application lifecycle is managed like an infrastucture resource from OKT's point of view, 
- a clear view on what is implemented in term of application lifecycle is provided thanks to the App LC Graph described by the CRD
  - Having all the operators in an organization built upon the same model should help human (or intelligent automates) operators to deal with several kind of K8S operators.
