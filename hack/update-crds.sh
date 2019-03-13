if [[ $GOPATH == "" ]] ; then
	echo "Need gopath to proceed ! exiting!!!"
	exit 1
fi

echo "cloning generators, might fail, no big deal"
mkdir -p $GOPATH/src/k8s.io

pushd $GOPATH/src/k8s.io
  commits=( b1289fc74931d4b6b04bd1a259acfc88a2cb0a66 94ebb086c69b9fec4ddbfb6a1433d28ecca9292b d216743eed4c3242b85d094d2a589f41a793652d )
  j=0
  for REPO in code-generator apimachinery api
  do
    git clone git@github.com:kubernetes/${REPO}.git
    pushd $REPO
      git checkout ${commits[j]}
    popd
    let "j++"
  done
popd 

pushd $GOPATH/src/k8s.io/code-generator
  crds=( blackduck opssight alert rgp )
  crdVersions=( v1 v1 v1 v1 )
  j=0
  for i in "${crds[@]}" ; do
    set +x
    ./generate-groups.sh "deepcopy,client,informer,lister" "github.com/blackducksoftware/synopsys-operator/pkg/${i}/client" "github.com/blackducksoftware/synopsys-operator/pkg/api" ${i}:${crdVersions[j]}
    let "j++"
  done
popd
