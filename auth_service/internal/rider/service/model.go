package rider_service


type RegisterResponse struct {
	Message       string
	Err           error
	Access_token  string
	Refresh_token string
	Status        string
	StatusCode    int16
}

type LoginResponse struct {
	Message       string
	Err           error
	Access_token  string
	Refresh_token string
	Status        string
	StatusCode    int16
}

type GetRiderInfo struct {
	Rider_id      string
	Firstname      string
	Lastname       string
	Phone_number   string
	Email          string
	ImageURL       string
	Rating         float64
}

type GetRiderInfoServiceResponse struct {
	Rider_info GetRiderInfo
	Message     string
	Err         error
	Status      string
	StatusCode  int16
}
