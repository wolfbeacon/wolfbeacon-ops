package main

// User struct, will be read from users.json file
type User struct {
	Email       string   `json:"email"`
	Permissions []string `json:"permissions"`
}

func (u *User) Can(permission string) bool {
	for _, p := range u.Permissions {
		if permission == p {
			return true
		}
	}
	return false
}