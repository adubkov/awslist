# awslist
Tool to get list of aws instances from all regions and accounts.

## configuration
awslist use `~/.aws/config` to load credentials, so it should be something like:
```
[default]
aws_access_key_id=ACCESS_KEY1
aws_secret_access_key=SECRET_ACCESS_KEY1

[profile prod]
aws_access_key_id=ACCESS_KEY2
aws_secret_access_key=SECRET_ACCESS_KEY2

[profile qa]
aws_access_key_id=ACCESS_KEY3
aws_secret_access_key=SECRET_ACCESS_KEY3
```

You can run it one time to get list of instances:
```
$ ./awslist
i-444bf444,qa-mongo-444bf444,10.10.10.10,m3.medium,None,us-west-1,qa
...
```

or run in as a service to pull data over http with something like `curl http://127.0.0.1:8080`:
```
$ ./awslist -service=true
2015/07/06 23:07:31 [INFO] Runing http server on port: 8080
2015/07/06 23:07:43 [INFO][127.0.0.1:8080]: GET / request from 127.0.0.1:52825. 568 instances was returned.
...
```
