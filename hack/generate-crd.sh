if [[ $# -ne 1 ]]; then
    echo "Error: Invalid number of arguments"
    echo "Usage: ./generate-crd.sh <crd_name>"
    exit 1
fi

crd_name=$1
crd_name_upper="$(tr '[:lower:]' '[:upper:]' <<< ${crd_name:0:1})${crd_name:1}"

# Create directory for CRD definition
echo "Creating Directories for CRD"
mkdir "../pkg/api/$crd_name"
mkdir "../pkg/api/$crd_name/v1"

echo "Copying CRD files from the Sample CRD"
cp "../pkg/api/sample/register.go" "../pkg/api/$crd_name/register.go"
sed -i "" -e "s/sample/$crd_name/g" "../pkg/api/$crd_name/register.go"
sed -i "" -e "s/Sample/$crd_name_upper/g" "../pkg/api/$crd_name/register.go"
echo " > register.go"

cp "../pkg/api/sample/v1/register.go" "../pkg/api/$crd_name/v1/register.go"
sed -i "" -e "s/sample/$crd_name/g" "../pkg/api/$crd_name/v1/register.go"
sed -i "" -e "s/Sample/$crd_name_upper/g" "../pkg/api/$crd_name/v1/register.go"
echo " > v1/register.go"

cp "../pkg/api/sample/v1/doc.go" "../pkg/api/$crd_name/v1/doc.go"
sed -i "" -e "s/sample/$crd_name/g" "../pkg/api/$crd_name/v1/doc.go"
sed -i "" -e "s/Sample/$crd_name_upper/g" "../pkg/api/$crd_name/v1/doc.go"
echo " > v1/doc.go"

cp "../pkg/api/sample/v1/types.go" "../pkg/api/$crd_name/v1/types.go"
sed -i "" -e "s/sample/$crd_name/g" "../pkg/api/$crd_name/v1/types.go"
sed -i "" -e "s/Sample/$crd_name_upper/g" "../pkg/api/$crd_name/v1/types.go"
echo " > v1/types.go"

### THIS CODE IS COPY-PASTA FROM update-crds.sh ####################
echo "Generating Kubernetes Client with update-crds.sh"
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
  crds=( $crd_name )
  crdVersions=( v1 )
  j=0
  for i in "${crds[@]}" ; do
    set +x
    ./generate-groups.sh "deepcopy,client,informer,lister" "github.com/blackducksoftware/synopsys-operator/pkg/${i}/client" "github.com/blackducksoftware/synopsys-operator/pkg/api" ${i}:${crdVersions[j]}
    let "j++"
  done
popd
########################################################################

echo "Copying Controller Files from Sample"
cp "../pkg/sample/crdinstaller.go" "../pkg/$crd_name/crdinstaller.go"
sed -i "" -e "s/sample/$crd_name/g" "../pkg/$crd_name/crdinstaller.go"
sed -i "" -e "s/Sample/$crd_name_upper/g" "../pkg/$crd_name/crdinstaller.go"
echo " > crdinstaller.go"

cp "../pkg/sample/crdcontroller.go" "../pkg/$crd_name/crdcontroller.go"
sed -i "" -e "s/sample/$crd_name/g" "../pkg/$crd_name/crdcontroller.go"
sed -i "" -e "s/Sample/$crd_name_upper/g" "../pkg/$crd_name/crdcontroller.go"
echo " > crdcontroller.go"

cp "../pkg/sample/crdhandler.go" "../pkg/$crd_name/crdhandler.go"
sed -i "" -e "s/sample/$crd_name/g" "../pkg/$crd_name/crdhandler.go"
sed -i "" -e "s/Sample/$crd_name_upper/g" "../pkg/$crd_name/crdhandler.go"
echo " > crdhandler.go"

cp "../pkg/sample/samplecreater.go" "../pkg/$crd_name/${crd_name}creater.go"
sed -i "" -e "s/sample/$crd_name/g" "../pkg/$crd_name/${crd_name}creater.go"
sed -i "" -e "s/Sample/$crd_name_upper/g" "../pkg/$crd_name/${crd_name}creater.go"
echo " > ${crd_name}creater.go"

cp "../pkg/sample/sample.go" "../pkg/$crd_name/${crd_name}.go"
sed -i "" -e "s/sample/$crd_name/g" "../pkg/$crd_name/${crd_name}.go"
sed -i "" -e "s/Sample/$crd_name_upper/g" "../pkg/$crd_name/${crd_name}.go"
echo " > ${crd_name}.go"

cp "../pkg/sample/sampledeployment.go" "../pkg/$crd_name/${crd_name}deployment.go"
sed -i "" -e "s/sample/$crd_name/g" "../pkg/$crd_name/${crd_name}deployment.go"
sed -i "" -e "s/Sample/$crd_name_upper/g" "../pkg/$crd_name/${crd_name}deployment.go"
echo " > ${crd_name}deployment.go"

