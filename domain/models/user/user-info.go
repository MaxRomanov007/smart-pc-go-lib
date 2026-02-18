package user

type Info struct {
	AuthTime int        `json:"auth_time"`
	IAT      int        `json:"iat"`
	RAT      int        `json:"rat"`
	Sub      string     `json:"sub"`
	Traits   InfoTraits `json:"traits"`
}

type InfoTraits struct {
	Email   string         `json:"email"`
	Name    InfoTraitsName `json:"name"`
	Picture string         `json:"picture"`
}

type InfoTraitsName struct {
	First string `json:"first"`
	Last  string `json:"last"`
}
