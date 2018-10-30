package service

import (
	"context"
	"net/http"

	"github.com/mweagle/Sparta"
	spartaREST "github.com/mweagle/Sparta/archetype/rest"
	"github.com/mweagle/Sparta/archetype/services"
	spartaAPIGateway "github.com/mweagle/Sparta/aws/apigateway"
	"github.com/sirupsen/logrus"
)

////////////////////////////////////////////////////////////////////////////////

////////////////////////////////////////////////////////////////////////////////
//

// TodoItemResource is the /todo/{id} resource
type TodoItemResource struct {
	services.S3Accessor
}

func (svc *TodoItemResource) isValidTodoID(apigRequest TodoRequest) bool {
	_, idParamExists := apigRequest.PathParams[todoIDParam]
	return idParamExists
}

/*
╔═╗╔═╗╔╦╗
║ ╦║╣  ║
╚═╝╚═╝ ╩
*/

// Get the incoming todo
func (svc *TodoItemResource) Get(ctx context.Context,
	apigRequest TodoRequest) (interface{}, error) {
	var todo Todo
	todoErr := svc.S3Accessor.Get(ctx,
		apigRequest.PathParams[todoIDParam],
		&todo)
	if todoErr != nil {
		return nil,
			spartaAPIGateway.NewErrorResponse(http.StatusNotFound,
				todoErr.Error())
	}
	return spartaAPIGateway.NewResponse(http.StatusOK, todo), nil
}

/*
╔═╗╔═╗╔╦╗╔═╗╦ ╦
╠═╝╠═╣ ║ ║  ╠═╣
╩  ╩ ╩ ╩ ╚═╝╩ ╩
*/

// Patch return empty
func (svc *TodoItemResource) Patch(ctx context.Context,
	apigRequest TodoRequest) (interface{}, error) {
	if !svc.isValidTodoID(apigRequest) {
		return nil, spartaAPIGateway.NewErrorResponse(http.StatusBadRequest, "Invalid Id")
	}

	logger, _ := ctx.Value(sparta.ContextKeyLogger).(*logrus.Logger)
	logger.WithField("Body", apigRequest.Body).Debug("TodoItemResource.Patch request body")

	// Merge it
	var existingTodo Todo
	existingTodoErr := svc.S3Accessor.Get(ctx,
		apigRequest.PathParams[todoIDParam],
		&existingTodo)
	if existingTodoErr != nil {
		return nil,
			spartaAPIGateway.NewErrorResponse(http.StatusNotFound,
				existingTodoErr.Error())
	}
	apigRequest.Body.URL = existingTodo.URL

	// Merge it, save it, return it...
	if apigRequest.Body.Order == 0 {
		apigRequest.Body.Order = existingTodo.Order
	}
	if apigRequest.Body.Title == "" {
		apigRequest.Body.Title = existingTodo.Title
	}

	// Save the item, put it back...
	saveErr := svc.Save(ctx,
		apigRequest.PathParams[todoIDParam],
		apigRequest.Body)
	if saveErr != nil {
		return nil,
			spartaAPIGateway.NewErrorResponse(http.StatusNotFound,
				saveErr.Error())
	}

	logger.WithField("Body", apigRequest.Body).
		Debug("TodoItemResource.Patch response body")

	return spartaAPIGateway.NewResponse(http.StatusOK, apigRequest.Body), nil
}

/*
╔╦╗╔═╗╦  ╔═╗╔╦╗╔═╗
 ║║║╣ ║  ║╣  ║ ║╣
═╩╝╚═╝╩═╝╚═╝ ╩ ╚═╝
*/

// Delete return delete
func (svc *TodoItemResource) Delete(ctx context.Context,
	apigRequest TodoRequest) (interface{}, error) {
	if !svc.isValidTodoID(apigRequest) {
		return nil, spartaAPIGateway.NewErrorResponse(http.StatusBadRequest, "Invalid Id")
	}
	deleteOneErr := svc.S3Accessor.Delete(ctx, apigRequest.PathParams[todoIDParam])
	if deleteOneErr != nil {
		return nil,
			spartaAPIGateway.NewErrorResponse(http.StatusInternalServerError,
				deleteOneErr.Error())
	}
	return nil, nil
}

// ResourceDefinition returns the Sparta REST definition for the Todo item
func (svc *TodoItemResource) ResourceDefinition() (spartaREST.ResourceDefinition, error) {
	return spartaREST.ResourceDefinition{
		URL: todoItemURL,
		MethodHandlers: spartaREST.MethodHandlerMap{
			// GET
			http.MethodGet: spartaREST.NewMethodHandler(svc.Get, http.StatusOK).
				StatusCodes(http.StatusInternalServerError).
				Privileges(svc.S3Accessor.KeysPrivilege("s3:GetObject"),
					svc.S3Accessor.BucketPrivilege("s3:ListBucket")),
			// PATCH
			http.MethodPatch: spartaREST.NewMethodHandler(svc.Patch, http.StatusOK).
				StatusCodes(http.StatusInternalServerError,
					http.StatusBadRequest).
				Privileges(svc.S3Accessor.KeysPrivilege("s3:GetObject"),
					svc.S3Accessor.KeysPrivilege("s3:PutObject")),
			// DELETE
			http.MethodDelete: spartaREST.NewMethodHandler(svc.Delete, http.StatusNoContent).
				StatusCodes(http.StatusInternalServerError).
				Privileges(svc.S3Accessor.KeysPrivilege("s3:Delete*", "s3:List*")),
		},
	}, nil
}

// END
////////////////////////////////////////////////////////////////////////////////
