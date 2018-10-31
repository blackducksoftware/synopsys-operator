function clone() {
    if [ -d docker-compose ]; then
        pushd ./docker-compose/
            git pull
        popd
    else
        git clone http://github.com/blackducksoftware/hub
    fi 
}

function printEnv() {
    pushd hub/docker-compose
        for f in `ls *env` ; do
            echo "#${f}"
            cat $f
        done
    popd
}
function printImage() {
   cat hub/docker-compose/*.yml | grep black | grep \. | sort | uniq | grep image
}


function clean(){
    rm -rf ./docker-compose/
}



clone
printEnv | sort > hubenv
printImage | sort > image



cat hubenv 
cat image