package utils

import (
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GRPCtoHTTPStatus(err error) (int, string) {
	if err == nil {
		return http.StatusOK, "OK"
	}

	st, ok := status.FromError(err)
	if !ok {
		return http.StatusInternalServerError, "Internal Server Error"
	}

	switch st.Code() {

	case codes.InvalidArgument:
		return http.StatusBadRequest, st.Message()

	case codes.Unauthenticated:
		return http.StatusUnauthorized, st.Message()

	case codes.PermissionDenied:
		return http.StatusForbidden, st.Message()

	case codes.NotFound:
		return http.StatusNotFound, st.Message()

	case codes.AlreadyExists:
		return http.StatusConflict, st.Message()

	case codes.Unavailable:
		return http.StatusServiceUnavailable, "Service is temporarily unavailable"

	case codes.Internal, codes.Unknown:
		return http.StatusInternalServerError, "An unexpected error occurred"

	default:
		return http.StatusInternalServerError, "An unexpected error occurred"
	}
}
