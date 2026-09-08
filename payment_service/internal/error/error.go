package errors

import (
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CreateError(validation map[string]string) error {
	st := status.New(codes.InvalidArgument, "invalid request payload")

	br := &errdetails.BadRequest{}

	for field, value := range validation {
		fmt.Println(field,value)
		br.FieldViolations = append(br.FieldViolations, &errdetails.BadRequest_FieldViolation{
			Field:       field,
			Description: value,
		})
	}

	stWithDetails, err := st.WithDetails(br)

	if err != nil {
		return st.Err()
	}

	return stWithDetails.Err()
}
