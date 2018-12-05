if [[ $GOPATH == "" ]] ; then
	echo "Need gopath to proceed ! exiting!!!"
	exit 1
fi

echo "cloning generators, might fail, no big deal"
mkdir -p $GOPATH/src/k8s.io

pushd $GOPATH/src/k8s.io
  git clone git@github.com:kubernetes/code-generator.git 
  git clone git@github.com:kubernetes/apimachinery.git
  git clone git@github.com:kubernetes/api.git
popd 

pushd $GOPATH/src/k8s.io/code-generator
  crds=( hub opssight alert )
  crdVersions=( v1 v1 v1 )
  j=0
  for i in "${crds[@]}" ; do
    set +x
    ./generate-groups.sh "deepcopy,client,informer,lister" "github.com/blackducksoftware/synopsys-operator/pkg/${i}/client" "github.com/blackducksoftware/synopsys-operator/pkg/api" ${i}:${crdVersions[j]}
  	let "j++"
  done
popd
