// Copyright 2021 Orange SA
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package apis

package k8s

import (
	"testing"

	"github.com/stretchr/testify/require"
	oktres "gitlab.tech.orange/dbmsprivate/operators/okt/resources"
	okterr "gitlab.tech.orange/dbmsprivate/operators/okt/results"
	okthash "gitlab.tech.orange/dbmsprivate/operators/okt/tools/hash"
	k8sres "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
}

// ==== BEGINNING OF STUB (TO GENERATE WITH CLI COMMAND)
// StatefulSetResourceStub an OKT extended StatefulSet resource
type StatefulSetResourceStub struct {
	Expected              k8sres.StatefulSet
	MutableResourceObject // OKT K8S resource
	oktres.MutationHelper
}

// blank assignment to verify that ReconcileCockroachDB implements reconcile.Reconciler
var _ oktres.MutableResource = &StatefulSetResourceStub{}

// GetResourceObject Implement a Stub interface function to get the Mutable Object
func (r *StatefulSetResourceStub) GetResourceObject() *ResourceObject {
	return &r.ResourceObject
}

// GetExpected Implements a Stub interface function to get the Expected object
func (r *StatefulSetResourceStub) GetExpected() *k8sres.StatefulSet {
	return &r.Expected
}

// Init Initialize OKT resource with K8S runtime object in the same Namespace of the Custom Resource
func (r *StatefulSetResourceStub) Init(client k8sclient.Client, namespace, name string) error {
	r.Expected.APIVersion = "apps/v1"
	r.Expected.Kind = "StatefulSet"
	r.MutationHelper = &DefaultMutationHelper{Expected: &r.Expected}

	return r.MutableResourceObject.Init(client, &r.Expected, namespace, name)
}

// PreMutate xx
func (r *StatefulSetResourceStub) PreMutate(scheme *runtime.Scheme) error {

	return nil
}

// PostMutate xx
func (r *StatefulSetResourceStub) PostMutate(cr k8sclient.Object, scheme *runtime.Scheme) error {
	if scheme != nil {
		if err := r.SetOwnerReference(cr, scheme); err != nil {
			return err
		}
	}

	return nil
}

// GetHashableRefHelper provide an helper for the HashableRef interface
// This will help to defines which object(s) data has to be used to detect modifications thanks to the Hash computation
func (r *StatefulSetResourceStub) GetHashableRefHelper() *HashableRefHelper {
	hr := &HashableRefHelper{}
	hr.Init(r.MutationHelper)

	return hr
}

func (r *StatefulSetResourceStub) getHelper() *StatefulSetHelper {
	helper := StatefulSetHelper{}
	helper.StatefulSetStub = r
	return &helper
}

// ==== END OF STUB (TO GENERATE WITH CLI COMMAND)

type myMutableStatefulSet struct {
	StatefulSetResourceStub
}

// Blank assigment to check type
var _ oktres.Mutator = &myMutableStatefulSet{}

func initTestStatefulSet(t *testing.T, tObj oktres.MutableResourceType) {
	require.False(t, tObj.IsCreation(), "Obj creation flag must be False while no verification to the existence on Cluster has been done for this test")
	require.False(t, tObj.NeedResync(), "Must be false by default")
	require.False(t, tObj.LastSyncState(), "Must be false by default")

	tObj.UpdateSyncStatus(tObj.GetHashableRef())
	require.True(t, tObj.LastSyncState(), "Creation time. Must be True, rigth after the first update of the hash key in annotations")
	require.True(t, tObj.NeedResync(), "Creation time. Must be True, rigth after the first update of the hash key in annotations")

	tObj.UpdateSyncStatus(tObj.GetHashableRef()) // Called twice without new modification!
	require.False(t, tObj.LastSyncState(), "A second call without any modification must show that the object has its hash key up to date")
	require.True(t, tObj.NeedResync(), "The object remains unsynched, even after a second call of UpdateSynchStatus()")
}

const yaml string = `---
# Source: orange-cockroachdb-insecure--k8s-chart/templates/8-StatefulSet-tpl.yaml
kind: StatefulSet
apiVersion: apps/v1
spec:
  #serviceName: crdb-cockroachdb-dev
  replicas: 4
  # Do not play to remove this selector/matchLabels part or you'll get a nil pointer exception 
  # on resource.Spec.Template.SetLabels(labelSelectors)
  selector:
    matchLabels:
#      setname: crdb-cockroachdb-dev
  template:
    metadata:
      labels:
        setname: crdb-cockroachdb-dev
    spec:
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 100
            podAffinityTerm:
              labelSelector:
                matchExpressions:
                - key: app
                  operator: In
                  values:
                  - cockroachdb
              topologyKey: kubernetes.io/hostname
      containers:
      - name: cockroachdb-sidecar
        image: dockerfactory-unstable-iva.si.francetelecom.fr/dbmsprivate/docker/orange-cockroachdb-sidecar/1.0.5-debian9-test
        imagePullPolicy: IfNotPresent
        ports:
        - name: crdb-sidecar
          containerPort: {{ .apiPort }}
          protocol: TCP
        resources:
          requests:
            cpu: 10m
            memory: 128Mi
          limits:
            cpu: 100m
            memory: 128Mi
        volumeMounts:
        - name: dbms-cockroachdb-run
          mountPath: /data/run
        livenessProbe:
          httpGet:
            path: /ping
            port: 8080
          initialDelaySeconds: 10
          timeoutSeconds: 3
          periodSeconds: 10
          successThreshold: 1
          failureThreshold: 3
        terminationMessagePath: /dev/termination-log


      - name: cockroachdb
        image: dockerfactory-playground-iva.si.francetelecom.fr/cockroachdb/cockroach:v19.2.5
        imagePullPolicy: Always
        ports:
        - containerPort: {{ .grpcPort }}
          name: grpc
        - containerPort: {{ .uiPort }}
          name: http
        livenessProbe:
          httpGet:
            path: "/health"
            port: http
            #scheme: HTTPS
          initialDelaySeconds: 30
          periodSeconds: 5
        readinessProbe:
          httpGet:
            path: "/health?ready=1"
            port: http
            #scheme: HTTPS
          initialDelaySeconds: 10
          periodSeconds: 5
          failureThreshold: 2
        volumeMounts:
        - name: dbms-cockroachdb-data
          mountPath: /cockroach/cockroach-data
        env:
        - name: COCKROACH_CHANNEL
          value: kubernetes-insecure
        resources:
          requests:
            memory: "500Mi"
            cpu: 128m
          limits:
            memory: "500Mi"
            cpu: "500m"
        command:
          - "/bin/bash"
          - "-ecx"
          # The use of qualified $(hostname -f) is crucial:
          # Other nodes aren't able to look up the unqualified hostname.
          # On Openshift
          #- "exec /cockroach/cockroach start --logtostderr --insecure --advertise-host $(hostname -f) --http-addr 0.0.0.0:8088 --join cockroachdb-0.crdb-cockroachdb-dev,cockroachdb-1.crdb-cockroachdb-dev,cockroachdb-2.crdb-cockroachdb-dev --cache 25% --max-sql-memory 25%"
          # On MicroK8S, because hostname -f return a longer name
          #- "exec /cockroach/cockroach start --logtostderr --insecure --advertise-host $(hostname -f) --http-addr 0.0.0.0:8088 --join crdb-cockroachdb-dev-0.crdb-cockroachdb-dev.default.svc.cluster.local,crdb-cockroachdb-dev-1.crdb-cockroachdb-dev.default.svc.cluster.local,crdb-cockroachdb-dev-2.crdb-cockroachdb-dev.default.svc.cluster.local --cache 25% --max-sql-memory 25%"
          # Finaly the DNS resolution with short names works...
          - "exec /cockroach/cockroach start --logtostderr --insecure --advertise-host $(hostname -f) --http-addr 0.0.0.0:8088 --join crdb-cockroachdb-dev-0.crdb-cockroachdb-dev,crdb-cockroachdb-dev-1.crdb-cockroachdb-dev,crdb-cockroachdb-dev-2.crdb-cockroachdb-dev --cache 25% --max-sql-memory 25%"
      # No pre-stop hook is required, a SIGTERM plus some time is all that's
      # needed for graceful shutdown of a node.
      terminationGracePeriodSeconds: 60

      volumes:
      - name: dbms-cockroachdb-data
        persistentVolumeClaim:
          claimName: dbms-cockroachdb-data
      - name: dbms-cockroachdb-run

#        FROM MongoDB template
#        restartPolicy: Always
#        terminationGracePeriodSeconds: 120
#        dnsPolicy: ClusterFirst
#        securityContext: {}


  podManagementPolicy: Parallel
  updateStrategy:
    type: RollingUpdate

  volumeClaimTemplates:
  - metadata:
      name: dbms-cockroachdb-data
    spec:
      accessModes:
        - "ReadWriteOnce"
      resources:
        requests:
          storage: "1Gi"
#      storageClassName: "microk8s-hostpath"`

// CockroachDBDefaults defines all the defaults/settings use to initialize resources at deployment time
var tplDefaults = map[string]string{
	"grpcPort":                    "26256",
	"apiPort":                     "8080",
	"uiPort":                      "8088",
	"podDisruptionMaxUnavailable": "1",
	"publicServiceNameSuffix":     "-public",
}

func (r *myMutableStatefulSet) GetHashableRef() okthash.HashableRef {
	helper := r.GetHashableRefHelper()
	helper.AddMetaLabels()
	helper.AddUserData(&r.Expected.Spec)

	return helper
}

// MutateWithDefaults xx
func (r *myMutableStatefulSet) MutateWithInitialData() error {
	if err := r.CopyTpl(yaml, tplDefaults); err != nil {
		return okterr.ErrGiveUpReconciliation.Reason(err)
	}

	return nil
}

// MutateWithCR xx
func (r *myMutableStatefulSet) MutateWithCR() (requeueAfterSeconds uint16, err error) {
	labels := map[string]string{
		"a": "b",
		"c": "d",
	}

	// Apply CR values
	r.Expected.SetLabels(labels)
	r.Expected.Spec.Selector.MatchLabels = labels
	r.Expected.Spec.Template.SetLabels(labels)

	return 0, nil
}

// blank assignment to verify that the resource implements the Mutable interface
var _ oktres.Mutator = &myMutableStatefulSet{}

func TestMutableStatefulSetObject(t *testing.T) {
	/*
		//scheme := runtime.NewScheme()
		//utilruntime.Must(clientgoscheme.AddToScheme(scheme))
		///utilruntime.Must(aggregatorclientsetscheme.AddToScheme(scheme))
		// +kubebuilder:scaffold:scheme

			apischeme := runtime.NewScheme()
			_ = clientgoscheme.AddToScheme(apischeme)
			_ = k8sapp.AddToScheme(apischeme)
	*/
	// Create and Init resource and check Sync status
	tObj := myMutableStatefulSet{}
	_ = tObj.Init(nil, "myns", "myname")
	statefulset := &tObj.Expected
	initTestStatefulSet(t, &tObj)

	statefulset.Spec.ServiceName = "zoz"
	tObj.MutateWithInitialData()
	tObj.MutateWithCR()
	require.True(t, statefulset.Spec.ServiceName == "zoz", "Mutation with Yaml templates should not override unmanaged fields")

	tObj.UpdateSyncStatus(tObj.GetHashableRef())
	require.True(t, tObj.LastSyncState(), "Must be True, rigth after the first update of the hash key in annotations")

	tObj.UpdateSyncStatus(tObj.GetHashableRef()) // Called twice to reset status to False after Init!
	require.False(t, tObj.LastSyncState(), "Now the object is considered as Synched")

	//apischeme.Default(statefulset)
	//scheme.Default(statefulset)
	//kappv1.SetDefaults_StatefulSet(statefullset)
	//require.True(t, tObj.NeedResync(), "Some defaults have been added")
}
