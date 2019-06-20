#!/bin/bash

# Copyright (C) 2019 Synopsys, Inc.
#
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements. See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership. The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License. You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied. See the License for the
# specific language governing permissions and limitations
# under the License.

PROJECT_ID=$1
BRANCH_NAME=$2
DOCKER_PROJECT=$3
VERSION=$(cat build.properties | cut -d- -f 1 | cut -d= -f 2)

if [ "$DOCKER_PROJECT" = "" ];
then
    echo "skipping......."
else
    docker tag "gcr.io/$PROJECT_ID/$DOCKER_PROJECT/synopsys-operator:$BRANCH_NAME" "gcr.io/$PROJECT_ID/$DOCKER_PROJECT/synopsys-operator:$VERSION"
    docker push "gcr.io/$PROJECT_ID/$DOCKER_PROJECT/synopsys-operator:$VERSION"
fi