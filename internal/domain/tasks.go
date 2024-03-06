package domain

// Task 领域对象，是 DDD 中的 entity
// BO(business object)
type Task struct {
	Id       int64
	Email    string
	Password string
	Phone    string
}
