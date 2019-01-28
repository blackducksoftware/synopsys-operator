if [[ $GOPATH == "" ]] ; then
	echo "Need gopath to proceed ! exiting!!!"
	exit 1
fi

echo "cloning generators, might fail, no big deal"
mkdir -p $GOPATH/src/k8s.io

pushd $GOPATH/src/k8s.io
  for REPO in code-generator apimachinery api
  do
    git clone git@github.com:kubernetes/${REPO}.git
    pushd $REPO
      git pull
    popd
  done
popd 

pushd $GOPATH/src/k8s.io/code-generator
  crds=( opssight alert blackduck )
  crdVersions=( v1 v1 v1 )
  j=0
  for i in "${crds[@]}" ; do
    set +x
    ./generate-groups.sh "deepcopy,client,informer,lister" "github.com/blackducksoftware/synopsys-operator/pkg/${i}/client" "github.com/blackducksoftware/synopsys-operator/pkg/api" ${i}:${crdVersions[j]}
    let "j++"
  done
popd
