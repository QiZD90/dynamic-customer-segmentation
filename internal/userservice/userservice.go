package userservice

type UserService interface {
	GetRandomUsers(percent int) ([]int, error)
}
