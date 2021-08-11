EXE=okt-gen-resource

all: build install

example: clean build gen

install: build
	cd okt-gen-resource; mv ${EXE} ${HOME}/bin/.

build: 
	cd okt-gen-resource; go build .

gen:
	cd okt-gen-resource/example; ../${EXE} -type=ConfigMap -kind=MyOperator -group=demo -version=v1alpha1 -path=okt.op.orange/example ./MyData1.yaml
	cd okt-gen-resource/example; ../${EXE} -type=StatefulSet -kind=MyOperator -group=demo -version=v1alpha1 -path=okt.op.orange/example MyStatefulset
	cd okt-gen-resource/example; ../${EXE} -type=Secret -kind=MyOperator -group=demo -version=v1alpha1 -path=okt.op.orange/example MySecret

clean:
	cd okt-gen-resource; rm -f ${EXE} 
	cd okt-gen-resource/example; rm -f *.go *.bak

