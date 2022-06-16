package data

import (
	"database/sql"
	"errors"
	"time"
)

func NewTestModels(dbPool *sql.DB) Models {
	db = dbPool
	return Models{
		User: &UserTest{},
		Plan: &PlanTest{},
	}
}

type UserTest struct {
	ID        int
	Email     string
	FirstName string
	LastName  string
	Password  string
	Active    int
	IsAdmin   int
	CreatedAt time.Time
	UpdatedAt time.Time
	Plan      *PlanTest
	FailTest  bool
}

type PlanTest struct {
	ID                  int
	PlanName            string
	PlanAmount          int
	PlanAmountFormatted string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	FailTest            bool
}

// GetAll returns a slice of all users, sorted by last name
func (u *UserTest) GetAll() ([]*User, error) {

	if u.FailTest {
		return nil, errors.New("test oops")
	}

	var users []*User

	user := User{
		ID:        1,
		Email:     "killroy@here.com",
		FirstName: "Killroy",
		LastName:  "DeJoy",
		Password:  "not-secret",
		Active:    1,
		IsAdmin:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	users = append(users, &user)

	return users, nil
}

// GetByEmail returns one user by email
func (u *UserTest) GetByEmail(email string) (*User, error) {

	if u.FailTest {
		return nil, sql.ErrNoRows
	}

	user := User{
		ID:        1,
		Email:     email,
		FirstName: "Killroy",
		LastName:  "DeJoy",
		Password:  "not-secret",
		Active:    1,
		IsAdmin:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	plan := Plan{
		ID:                  1,
		PlanName:            "Fake Plan",
		PlanAmount:          1500,
		PlanAmountFormatted: "$15.00",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	user.Plan = &plan

	return &user, nil
}

// GetOne returns one user by id
func (u *UserTest) GetOne(id int) (*User, error) {

	if u.FailTest {
		return nil, sql.ErrNoRows
	}

	user := User{
		ID:        id,
		Email:     "killroy@here.com",
		FirstName: "Killroy",
		LastName:  "DeJoy",
		Password:  "not-secret",
		Active:    1,
		IsAdmin:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	plan := Plan{
		ID:                  1,
		PlanName:            "Fake Plan",
		PlanAmount:          1500,
		PlanAmountFormatted: "$15.00",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	user.Plan = &plan

	return &user, nil
}

// Update updates one user in the database, using the information
// stored in the receiver u
func (u *UserTest) Update(user User) error {
	if u.FailTest {
		return errors.New("test oops")
	}

	return nil
}

// Delete deletes one user from the database, by User.ID
func (u *UserTest) Delete() error {
	if u.FailTest {
		return errors.New("test oops")
	}
	return nil
}

// DeleteByID deletes one user from the database, by ID
func (u *UserTest) DeleteByID(id int) error {
	if u.FailTest {
		return errors.New("test oops")
	}
	return nil
}

// Insert inserts a new user into the database, and returns the ID of the newly inserted row
func (u *UserTest) Insert(user User) (int, error) {
	if u.FailTest {
		return 0, errors.New("test oops")
	}
	return 2, nil
}

// ResetPassword is the method we will use to change a user's password.
func (u *UserTest) ResetPassword(user User, password string) error {
	if u.FailTest {
		return errors.New("test oops")
	}
	return nil
}

// PasswordMatches uses Go's bcrypt package to compare a user supplied password
// with the hash we have stored for a given user in the database. If the password
// and hash match, we return true; otherwise, we return false.
func (u *UserTest) PasswordMatches(user User, plainText string) (bool, error) {
	if u.FailTest {
		return false, errors.New("test oops")
	}
	return true, nil
}

func (p *PlanTest) GetAll() ([]*Plan, error) {
	if p.FailTest {
		return nil, errors.New("test oops")
	}

	var plans []*Plan

	plan := Plan{
		ID:                  1,
		PlanName:            "Fake Plan",
		PlanAmount:          1500,
		PlanAmountFormatted: "$15.00",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	plans = append(plans, &plan)

	return plans, nil
}

// GetOne returns one plan by id
func (p *PlanTest) GetOne(id int) (*Plan, error) {
	if p.FailTest {
		return nil, sql.ErrNoRows
	}

	plan := Plan{
		ID:                  id,
		PlanName:            "Fake Plan",
		PlanAmount:          1500,
		PlanAmountFormatted: "$15.00",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	return &plan, nil
}

// SubscribeUserToPlan subscribes a user to one plan by insert
// values into user_plans table
func (p *PlanTest) SubscribeUserToPlan(user User, plan Plan) error {
	if p.FailTest {
		return errors.New("test ooops")
	}
	return nil
}

// AmountForDisplay formats the price we have in the DB as a currency string
func (p *PlanTest) AmountForDisplay() string {
	return "$15.00"
}
