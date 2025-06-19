package share

type CreateBody struct {
	Password string `json:"password"`
	Expires  string `json:"expires"`
	Unit     string `json:"unit"`
}
type Link struct {
	Hash         string `json:"hash" storm:"id,index"`
	Path         string `json:"path" storm:"index"`
	Source       string `json:"source" storm:"index"`
	UserID       uint   `json:"userID"`
	Expire       int64  `json:"expire"`
	PasswordHash string `json:"password_hash,omitempty"`
	Token        string `json:"token,omitempty"`
}
