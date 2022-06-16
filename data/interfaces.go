package data

type UserType interface {
	GetAll() ([]*User, error)
	GetByEmail(email string) (*User, error)
	GetOne(id int) (*User, error)
	Update() error
	Delete() error
	DeleteByID(id int) error
	Insert(user User) (int, error)
	ResetPassword(password string) error
	PasswordMatches(plainText string) (bool, error)
}

type PlanType interface {
	GetAll() ([]*Plan, error)
	GetOne(id int) (*Plan, error)
	SubscribeUserToPlan(user User, plan Plan) error
	AmountForDisplay() string
}
