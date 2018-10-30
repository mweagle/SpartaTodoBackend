# SpartaTodoBackend
[Sparta](https://github.com/mweagle/Sparta) application that demonstrates provisioning an REST style API
that satisfies the https://todobackend.com spec tests

## Instructions

1. `git clone https://github.com/mweagle/SpartaTodoBackend`
1. `cd SpartaTodoBackend`
1. `go get -u -v ./...`
1. `S3_BUCKET=<MY_S3_BUCKET_NAME> mage provision`
1. In the _Stack output_ section of the log, look for the **API Gateway URL** key and copy it (eg: _https://XXXXXXXXXX.execute-api.us-west-2.amazonaws.com/v1_).
1. Visit https://www.todobackend.com/specs/ and test your new API


## Result

<div align="center"><img src="https://raw.githubusercontent.com/mweagle/SpartaTodoBackend/master/site/test_results.jpg" />
</div>

