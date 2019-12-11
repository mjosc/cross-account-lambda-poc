package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/sts"
)

type Input struct {
	ParameterName  string `json:"parameterName"`
	ParameterValue string `json:"parameterValue"`
}

type Output interface {
}

type SuccessResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func putParameter(svc *ssm.SSM, input *Input) error {
	_, err := svc.PutParameter(&ssm.PutParameterInput{
		Description: aws.String("Test input to validate using a lambda to update SSM across multiple accounts"),
		Name:        &input.ParameterName,
		Overwrite:   aws.Bool(true),
		Type:        aws.String("String"),
		Value:       &input.ParameterValue,
	})
	return err
}

func errorResponse(msg string, err error) (Output, error) {
	return &ErrorResponse{
		Message: msg,
		Error:   err.Error(),
	}, err
}

func handle(input *Input) (Output, error) {

	assumeRoleAccount := os.Getenv("ASSUME_ROLE_ACCOUNT")
	if assumeRoleAccount == "" {
		msg := "ASSUME_ROLE_ACCOUNT not set"
		return errorResponse(msg, errors.New(msg))
	}
	assumeRoleArn := fmt.Sprintf("arn:aws:iam::%v:role/gurgler-lambda", assumeRoleAccount)

	sess := session.Must(session.NewSession())
	stsSvc := sts.New(sess)
	ssmSvc := ssm.New(sess)

	identity, err := stsSvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return errorResponse("unable to retrieve caller identity", err)
	} else if identity.Account == nil {
		return errorResponse("unable to retreive current account", errors.New("empty account"))
	}

	fmt.Printf("updating ssm parameter %v in account %v\n", input.ParameterName, *identity.Account)
	if err := putParameter(ssmSvc, input); err != nil {
		message := fmt.Sprintf("unable to put parameter %v in account %v", input.ParameterName, identity.Account)
		return errorResponse(message, err)
	}

	creds := stscreds.NewCredentials(sess, assumeRoleArn)
	ssmSvc = ssm.New(sess, &aws.Config{Credentials: creds})

	fmt.Printf("updating ssm parameter %v in account %v\n", input.ParameterName, assumeRoleAccount)
	if err := putParameter(ssmSvc, input); err != nil {
		message := fmt.Sprintf("unable to put parameter %v in account %v", input.ParameterName, assumeRoleAccount)
		return errorResponse(message, err)
	}

	fmt.Println("success")

	return &SuccessResponse{
		Message: "success",
	}, nil
}

func main() {
	lambda.Start(handle)
}
