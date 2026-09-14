package rider_handler

type UpdatePasswordRequest struct {
	CurrentPassword    string `json:"curr_password"`
	NewPassword        string `json:"new_password"`
	ConfirmNewPassword string `json:"confirm_new_password"`
}
