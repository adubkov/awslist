# awslist
Tool to list resources in aws accounts.

# build and deploy

```
$ make all

# First we should build docker container:
docker build -t yourreponame:port/awslist:latest .

# Push container to our registrty
docker push yourreponame:port/awslist:latest

# On docker host we should create init file to keep this container running as a service:
cat > /etc/init/awslist.conf <<EOF
description "awslist service"

start on runlevel [2345]
stop on runlevel [016]

respawn
respawn limit 10 5

script
	exec docker run \
	 -v /root/.aws/:/root/.aws/ \
         -p 8080:8080 \
         adubkov/awslist:latest \
	 2>&1 | logger -t awslist

end script
EOF
```

## configuration
awslist using `~/.aws/credentials`:
```
[default]
aws_access_key_id=ACCESS_KEY1
aws_secret_access_key=SECRET_ACCESS_KEY1

[prod]
aws_access_key_id=ACCESS_KEY2
aws_secret_access_key=SECRET_ACCESS_KEY2

[qa]
aws_access_key_id=ACCESS_KEY3
aws_secret_access_key=SECRET_ACCESS_KEY3
```

# usage
Once container is up, you can perform curl requests to get list of aws resources:
```
curl -s http://127.0.0.1:8080
curl -s http://127.0.0.1:8080/ec2
curl -s http://127.0.0.1:8080/elb
curl -s http://127.0.0.1:8080/elb/{profile_name}
curl -s http://127.0.0.1:8080/elb/{profile_name}/{elb_name}
```

# Install with Helm

* require kiam configured

```
function apply_namespace() {
    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Namespace
metadata:
  name: ${1}
  annotations:
    iam.amazonaws.com/permitted: .*
EOF
}

apply_namespace awslist

helm upgrade -i awslist --namespace awslist ./helm/awslist
```
