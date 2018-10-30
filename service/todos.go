package service

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/mweagle/Sparta"
	spartaREST "github.com/mweagle/Sparta/archetype/rest"
	"github.com/mweagle/Sparta/archetype/services"
	spartaAPIGateway "github.com/mweagle/Sparta/aws/apigateway"
	"github.com/sirupsen/logrus"
)

////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
//

// TodoCollectionResource is the /todo resource
type TodoCollectionResource struct {
	services.S3Accessor
}

/*
╔═╗╔═╗╔╦╗
║ ╦║╣  ║
╚═╝╚═╝ ╩
*/

// Get the incoming todo
func (svc *TodoCollectionResource) Get(ctx context.Context,
	apigRequest TodoRequest) (interface{}, error) {
	ctor := func() interface{} {
		return &Todo{}
	}
	todos, todosErr := svc.S3Accessor.GetAll(ctx, ctor)
	if todosErr != nil {
		return nil,
			spartaAPIGateway.NewErrorResponse(http.StatusInternalServerError,
				todosErr.Error())
	}
	return spartaAPIGateway.NewResponse(http.StatusOK, todos), nil
}

/*
╔═╗╔═╗╔═╗╔╦╗
╠═╝║ ║╚═╗ ║
╩  ╚═╝╚═╝ ╩
*/

// Post to the collection
func (svc *TodoCollectionResource) Post(ctx context.Context,
	apigRequest TodoRequest) (interface{}, error) {
	logger, _ := ctx.Value(sparta.ContextKeyLogger).(*logrus.Logger)

	apigRequest.Body.ID = apigRequest.Context.RequestID
	todoURL := fmt.Sprintf("%s/v1%s/%s",
		os.Getenv("REST_API"),
		todoRootURL,
		apigRequest.Body.ID)
	apigRequest.Body.URL = todoURL

	logger.WithField("Body", apigRequest.Body).
		Debug("TodoCollectionResource.Post saving body")

	saveErr := svc.S3Accessor.Save(ctx,
		apigRequest.Body.ID,
		apigRequest.Body)
	if saveErr != nil {
		return nil,
			spartaAPIGateway.NewErrorResponse(http.StatusInternalServerError)
	}

	logger.WithField("Body", apigRequest.Body).
		Debug("TodoCollectionResource.Post response body")

	return spartaAPIGateway.NewResponse(http.StatusCreated,
		apigRequest.Body,
		map[string]string{
			"Location": todoURL,
		}), nil
}

/*
╔╦╗╔═╗╦  ╔═╗╔╦╗╔═╗
 ║║║╣ ║  ║╣  ║ ║╣
═╩╝╚═╝╩═╝╚═╝ ╩ ╚═╝
*/

// Delete all the Todos
func (svc *TodoCollectionResource) Delete(ctx context.Context,
	apigRequest TodoRequest) (interface{}, error) {
	deleteAllErr := svc.S3Accessor.DeleteAll(ctx)
	if deleteAllErr != nil {
		return nil,
			spartaAPIGateway.NewErrorResponse(http.StatusInternalServerError,
				deleteAllErr.Error())
	}
	return spartaAPIGateway.NewResponse(http.StatusOK, []string{}), nil
}

// ResourceDefinition returns the Sparta REST definition for the
// todos collection resource
func (svc *TodoCollectionResource) ResourceDefinition() (spartaREST.ResourceDefinition, error) {
	// https://docs.aws.amazon.com/IAM/latest/UserGuide/list_amazons3.html
	return spartaREST.ResourceDefinition{
		URL: todoRootURL,
		MethodHandlers: spartaREST.MethodHandlerMap{
			// GET
			http.MethodGet: spartaREST.NewMethodHandler(svc.Get, http.StatusOK).
				StatusCodes(http.StatusInternalServerError).
				Privileges(svc.S3Accessor.KeysPrivilege("s3:GetObject"),
					svc.S3Accessor.BucketPrivilege("s3:ListBucket")),
			// POST
			http.MethodPost: spartaREST.NewMethodHandler(svc.Post, http.StatusCreated).
				StatusCodes(http.StatusInternalServerError).
				Headers("Location").
				Privileges(svc.S3Accessor.KeysPrivilege("s3:PutObject")),
			// DELETE
			http.MethodDelete: spartaREST.NewMethodHandler(svc.Delete, http.StatusNoContent).
				StatusCodes(http.StatusInternalServerError).
				Privileges(svc.S3Accessor.KeysPrivilege("s3:DeleteObject"),
					svc.S3Accessor.BucketPrivilege("s3:ListBucket")),
		},
	}, nil
}
