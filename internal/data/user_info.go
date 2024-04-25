package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"golangHW.darkhanomirbay/internal/validator"
	"time"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

// Declare a new AnonymousUser variable.
var AnonymousUser = &UserInfo{}

func (u *UserInfo) IsAnonymous() bool {
	return u == AnonymousUser
}

type UserInfo struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Surname   string    `json:"surname"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Role      string    `json:"role"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}
type password struct {
	plaintext *string
	hash      []byte
}
type UserInfoModel struct {
	DB *sql.DB
}

func (m UserInfoModel) Insert(user *UserInfo) error {
	query := `INSERT INTO user_info(fname,sname,email,password_hash,user_role,activated) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id,created_at,updated_at,version`

	args := []any{user.Name, user.Surname, user.Email, user.Password.hash, user.Role, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}
func (m UserInfoModel) GetByEmail(email string) (*UserInfo, error) {
	query := `SELECT id, created_at, updated_at,fname,sname, email, password_hash, user_role,activated, version
FROM user_info WHERE email=$1`
	var user UserInfo
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Name,
		&user.Surname,
		&user.Email,
		&user.Password.hash,
		&user.Role,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}
func (m UserInfoModel) Update(user *UserInfo) error {
	query := `UPDATE user_info SET fname=$1,sname=$2,email=$3,password_hash=$4,user_role=$5,activated=$6,version=version + 1 WHERE id=$7 AND version=$8`
	args := []any{
		user.Name,
		user.Surname,
		user.Email,
		user.Password.hash,
		user.Role,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}

	}
	return nil
}
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPassword
	p.hash = hash
	return nil
}
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}
func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}
func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")

}
func ValidateUser(v *validator.Validator, user *UserInfo) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(v, user.Email)
	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}
func (m UserInfoModel) GetForToken(tokenScope, tokenPlaintext string) (*UserInfo, error) {
	// Calculate the SHA-256 hash of the plaintext token provided by the client.
	// Remember that this returns a byte *array* with length 32, not a slice.
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))
	// Set up the SQL query.
	query := `
SELECT user_info.id, user_info.created_at, user_info.updated_at,user_info.fname, user_info.sname,user_info.email, user_info.password_hash, user_info.user_role,user_info.activated,user_info.version
FROM user_info
INNER JOIN tokens
ON user_info.id = tokens.user_id
WHERE tokens.hash = $1
AND tokens.scope = $2
AND tokens.expiry > $3`
	// Create a slice containing the query arguments. Notice how we use the [:] operator
	// to get a slice containing the token hash, rather than passing in the array (which
	// is not supported by the pq driver), and that we pass the current time as the
	// value to check against the token expiry.
	args := []any{tokenHash[:], tokenScope, time.Now()}
	var user UserInfo
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Execute the query, scanning the return values into a User struct. If no matching
	// record is found we return an ErrRecordNotFound error.
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Name,
		&user.Surname,
		&user.Email,
		&user.Password.hash,
		&user.Role,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// Return the matching user.
	return &user, nil
}
func (m *UserInfoModel) Get(id int64) (*UserInfo, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := `SELECT id,created_at,updated_at,fname,sname,email,password_hash,user_role,activated,version FROM user_info WHERE id=$1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var user UserInfo
	err := m.DB.QueryRowContext(ctx, query, id).Scan(&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Name,
		&user.Surname,
		&user.Email,
		&user.Password.hash,
		&user.Role,
		&user.Activated,
		&user.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}
func (m *UserInfoModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `DELETE FROM user_info WHERE id=$1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
func (m *UserInfoModel) GetAll(Fname string, Sname string, filters Filters) ([]*UserInfo, Metadata, error) {
	query := fmt.Sprintf(`SELECT count(*) OVER(), id,created_at,updated_at,fname,sname,email,password_hash,user_role,activated,version
	FROM user_info
	WHERE (to_tsvector('simple', fname) @@ plainto_tsquery('simple', $1) OR $1 = '')
	AND (to_tsvector('simple', sname) @@ plainto_tsquery('simple', $2) OR $2 = '')
	ORDER BY %s %s,id ASC
	LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, Fname, Sname, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()
	userInfos := []*UserInfo{}
	totalRecords := 0
	for rows.Next() {
		var user UserInfo

		err := rows.Scan(&totalRecords, &user.ID,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.Name,
			&user.Surname,
			&user.Email,
			&user.Password.hash,
			&user.Role,
			&user.Activated,
			&user.Version)
		if err != nil {
			return nil, Metadata{}, err
		}
		userInfos = append(userInfos, &user)
	}
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return userInfos, metadata, nil
}
func (m *UserInfoModel) GetAllNonActivated(Fname string, Sname string, filters Filters) ([]*UserInfo, Metadata, error) {
	query := fmt.Sprintf(`SELECT count(*) OVER(), id,created_at,updated_at,fname,sname,email,password_hash,user_role,activated,version
	FROM user_info
	WHERE activated=false AND 
	    (to_tsvector('simple', fname) @@ plainto_tsquery('simple', $1) OR $1 = '')
	AND (to_tsvector('simple', sname) @@ plainto_tsquery('simple', $2) OR $2 = '')
	ORDER BY %s %s,id ASC
	LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, Fname, Sname, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()
	userInfos := []*UserInfo{}
	totalRecords := 0
	for rows.Next() {
		var user UserInfo

		err := rows.Scan(&totalRecords, &user.ID,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.Name,
			&user.Surname,
			&user.Email,
			&user.Password.hash,
			&user.Role,
			&user.Activated,
			&user.Version)
		if err != nil {
			return nil, Metadata{}, err
		}
		userInfos = append(userInfos, &user)
	}
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return userInfos, metadata, nil
}
func (m *UserInfoModel) GetAllNoActiv() ([]*UserInfo, error) {
	//	query := `SELECT id, created_at, updated_at, fname, sname, email, password_hash, user_role, activated, version
	//FROM user_info
	//WHERE activated = false';
	query2 := `SELECT u.id, u.created_at, u.updated_at, u.fname, u.sname, u.email, u.password_hash, u.user_role, u.activated, u.version
FROM user_info u
JOIN tokens t ON u.id = t.user_id
WHERE u.activated = false
AND t.expiry < NOW();
`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := m.DB.QueryContext(ctx, query2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	userInfos := []*UserInfo{}
	//totalRecords := 0
	for rows.Next() {
		var user UserInfo
		//&totalRecords,
		err := rows.Scan(&user.ID,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.Name,
			&user.Surname,
			&user.Email,
			&user.Password.hash,
			&user.Role,
			&user.Activated,
			&user.Version)
		if err != nil {
			return nil, err
		}
		userInfos = append(userInfos, &user)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return userInfos, nil
}
