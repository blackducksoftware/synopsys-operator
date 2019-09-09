if [[ $GOPATH == "" ]] ; then
	echo "Need gopath to proceed ! exiting!!!"
	exit 1
fi

echo "cloning generators, might fail, no big deal"
mkdir -p $GOPATH/src/k8s.io

pushd $GOPATH/src/k8s.io
  commits=( 50b561225d70b3eb79a1faafd3dfe7b1a62cbe73 d7deff9243b165ee192f5551710ea4285dcfd615 40a48860b5abbba9aa891b02b32da429b08d96a0 6ee68ca5fd8355d024d02f9db0b3b667e8357a0f )
  j=0
  for REPO in code-generator apimachinery api client-go
  do
    git clone git@github.com:kubernetes/${REPO}.git
    pushd $REPO
      git checkout master
      git pull
      git checkout ${commits[j]}
    popd
    let "j++"
  done
popd 

pushd $GOPATH/src/k8s.io/code-generator
  crds=( blackduck opssight alert size)
  crdVersions=( v1 v1 v1 v1)
  j=0
  for i in "${crds[@]}" ; do
    set +x
    ./generate-groups.sh "deepcopy,client,informer,lister" "github.com/blackducksoftware/synopsys-operator/pkg/${i}/client" "github.com/blackducksoftware/synopsys-operator/pkg/api" ${i}:${crdVersions[j]}
    let "j++"
  done
popd
