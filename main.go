package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	sparta "github.com/mweagle/Sparta"
	spartaREST "github.com/mweagle/Sparta/archetype/rest"
	spartaAccessor "github.com/mweagle/Sparta/aws/accessor"
	spartaCF "github.com/mweagle/Sparta/aws/cloudformation"
	spartaDecorator "github.com/mweagle/Sparta/decorator"
	todoResources "github.com/mweagle/SpartaTodoBackend/service"
	gocf "github.com/mweagle/go-cloudformation"
	"github.com/sirupsen/logrus"
)

func restResources(s3BucketResourceName string) []spartaREST.Resource {
	return []spartaREST.Resource{
		&todoResources.TodoCollectionResource{
			S3Accessor: spartaAccessor.S3Accessor{
				S3BucketResourceName: s3BucketResourceName,
			},
		},
		&todoResources.TodoItemResource{
			S3Accessor: spartaAccessor.S3Accessor{
				S3BucketResourceName: s3BucketResourceName,
			},
		},
	}
}

func spartaTodoBackendFunctions(api *sparta.API,
	s3BucketResourceName string) []*sparta.LambdaAWSInfo {
	lambdaFns := make([]*sparta.LambdaAWSInfo, 0)

	for _, eachResource := range restResources(s3BucketResourceName) {
		// Register the resources and lambda functions
		registeredFuncs, registeredFuncsErr := spartaREST.RegisterResource(api, eachResource)
		if registeredFuncsErr != nil {
			panic("Failed to initialize resource: " + registeredFuncsErr.Error())
		}
		lambdaFns = append(lambdaFns, registeredFuncs...)
	}
	// Tell each lambda function about the RestAPI
	for _, eachFunc := range lambdaFns {
		eachFunc.DependsOn = []string{s3BucketResourceName}
		eachFunc.Options.Environment = map[string]*gocf.StringExpr{
			"REST_API": api.RestAPIURL(),
		}
	}
	return lambdaFns
}

/*
================================================================================
╔╦╗╔═╗╔═╗╔═╗╦═╗╔═╗╔╦╗╔═╗╦═╗╔═╗
 ║║║╣ ║  ║ ║╠╦╝╠═╣ ║ ║ ║╠╦╝╚═╗
═╩╝╚═╝╚═╝╚═╝╩╚═╩ ╩ ╩ ╚═╝╩╚═╚═╝
================================================================================
*/

func workflowHooks(lambdaFuncs []*sparta.LambdaAWSInfo,
	s3BucketResourceName string) *sparta.WorkflowHooks {
	s3BucketHook := func(context map[string]interface{},
		serviceName string,
		cfTemplate *gocf.Template,
		S3Bucket string,
		S3Key string,
		buildID string,
		awsSession *session.Session,
		noop bool,
		logger *logrus.Logger) error {

		// Add the dynamic S3 bucket, orphan it...
		s3Bucket := &gocf.S3Bucket{}
		s3Resource := cfTemplate.AddResource(s3BucketResourceName,
			s3Bucket)
		s3Resource.DeletionPolicy = "Retain"
		return nil
	}

	// Setup the DashboardDecorator lambda hook
	workflowHooks := &sparta.WorkflowHooks{
		ServiceDecorators: []sparta.ServiceDecoratorHookHandler{
			sparta.ServiceDecoratorHookFunc(s3BucketHook),
			spartaDecorator.DashboardDecorator(lambdaFuncs, 60),
		},
	}
	return workflowHooks
}

////////////////////////////////////////////////////////////////////////////////
// Main
func main() {
	// Register the function with the API Gateway
	apiStage := sparta.NewStage("v1")
	apiGateway := sparta.NewAPIGateway("SpartaTodoBackend", apiStage)
	// Enable CORS s.t. the test harness can access the resources
	apiGateway.CORSOptions = &sparta.CORSOptions{
		Headers: map[string]interface{}{
			"Access-Control-Allow-Headers": "Content-Type,X-Amz-Date,Authorization,X-Api-Key,Location",
			"Access-Control-Allow-Methods": "*",
			"Access-Control-Allow-Origin":  "https://www.todobackend.com",
		},
	}
	// Lambda functions
	s3BucketResourceName := sparta.CloudFormationResourceName("S3Bucket", "S3Bucket")
	lambdaFns := spartaTodoBackendFunctions(apiGateway, s3BucketResourceName)

	// S3BucketResourceName
	hooks := workflowHooks(lambdaFns, s3BucketResourceName)

	// Deploy it
	stackName := spartaCF.UserScopedStackName("SpartaTodoBackend")
	sparta.MainEx(stackName,
		fmt.Sprintf("Provision a serverless TodoBackend service (https://todobackend.com)"),
		lambdaFns,
		apiGateway,
		nil,
		hooks,
		false)
}
