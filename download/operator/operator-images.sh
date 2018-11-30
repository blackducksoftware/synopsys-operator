#!/bin/bash

source `dirname ${BASH_SOURCE}`/args.sh "${@}"

if [[ "$_arg_push"  == "on" && "$_arg_project"  == "" ]]; then
    echo "Please provide the Docker project/repository to push the images!!!"
    exit 1
fi

OPERATOR_IMAGES=("synopsys-operator")

echo "*************************************************************************"
echo "Started pulling all Operator images"
echo "*************************************************************************"
for OPERATOR_IMAGE in "${OPERATOR_IMAGES[@]}"
do
	docker pull docker.io/blackducksoftware/"$OPERATOR_IMAGE":"$_arg_tag"
done
echo "*************************************************************************"
echo "Pulled all Operator images"
echo "*************************************************************************"

OPERATOR_DIR="operator-images"
if [[ "$_arg_push"  == "off" ]]; then
    mkdir -p ./"$OPERATOR_DIR"
    echo "*************************************************************************"
    echo "Started saving all Operator images"
    echo "*************************************************************************"
    for OPERATOR_IMAGE in "${OPERATOR_IMAGES[@]}"
    do
        docker save docker.io/blackducksoftware/"$OPERATOR_IMAGE":"$_arg_tag" -o ./"$OPERATOR_DIR"/"$OPERATOR_IMAGE".tar
    done
    echo "*************************************************************************"
    echo "Saved all Operator images in ./$OPERATOR_DIR"
    echo "*************************************************************************"
else
    echo ""
    echo ""
    echo "********************************************************************************************************************"
    echo "Please provide the Docker credentials of $_arg_registry registry for $_arg_user user..."
    echo "********************************************************************************************************************"
    docker login -u "$_arg_user" "$_arg_registry"

    # Docker tag and push all opssight images
    DOCKER_REPO="$_arg_registry"/"$_arg_project"
    echo "*************************************************************************"
    echo "Started tagging and pushing all Operator images"
    echo "*************************************************************************"
    for OPERATOR_IMAGE in "${OPERATOR_IMAGES[@]}"
    do
        docker tag docker.io/blackducksoftware/"$OPERATOR_IMAGE":"$_arg_tag" "$DOCKER_REPO"/"$OPERATOR_IMAGE":"$_arg_tag"
        docker push "$DOCKER_REPO"/"$OPERATOR_IMAGE":"$_arg_tag"
    done
    echo "*************************************************************************"
    echo "Tagged and pushed all Operator images"
    echo "*************************************************************************"
fi