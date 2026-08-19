package dto

type UpdateUsuarioRequest struct {
	Nome    *string `json:"nome,omitempty"`
	Email   *string `json:"email,omitempty"`
	Bio     *string `json:"bio,omitempty"`
	Version int     `json:"version" binding:"required"`
}
