package cms

import (
	"errors"
	"github.com/asaskevich/govalidator"
	"google.golang.org/appengine/datastore"
)

var PasswordField = &Field{
	Name:    "password",
	NoIndex: true,
	ValidateFunc: func(value interface{}) bool {
		return govalidator.IsByteLength(value.(string), 6, 128)
	},
}

var User = &Entity{
	Name: "user",
	Rules: Rules{
		Add: Guest,
	},
	NameFunc: func(providedFieldValue interface{}, oldName string, failedCount int) (string, error) {
		if failedCount > 0 {
			if valueString, ok := providedFieldValue.(string); ok {
				return "", errors.New("username " + valueString + " already taken")
			}
			return "", errors.New("username of invalid type")
		}
		if len(oldName) > 0 {
			return oldName, nil
		}
		if valueString, ok := providedFieldValue.(string); ok && govalidator.IsEmail(valueString){
			return valueString, nil
		}
		return "", errors.New("invalid provided field value")
	},
	Fields: []*Field{
		{
			Name:           "email",
			IsNameProvider: true,
			Rules: Rules{
				Write: Admin,
			},
			IsRequired: true,
			ValidateFunc: func(value interface{}) bool {
				return govalidator.IsEmail(value.(string))
			},
		},
		{
			Name:    "role",
			NoIndex: true,
			Rules: Rules{
				Write: Admin,
			},
		},
		PasswordField,
		/*{
			Name:       "firstName",
			IsRequired: true,
			Validator: func(value interface{}) bool {
				return govalidator.IsByteLength(value.(string), 1, 64)
			},
		},
		{
			Name:       "lastName",
			IsRequired: true,
			Validator: func(value interface{}) bool {
				return govalidator.IsByteLength(value.(string), 1, 64)
			},
		},
		{
			Name: "companyName",
			Validator: func(value interface{}) bool {
				return govalidator.IsByteLength(value.(string), 0, 64)
			},
		},
		{
			Name: "companyId",
			Validator: func(value interface{}) bool {
				return govalidator.IsByteLength(value.(string), 0, 8)
			},
		},
		{
			Name:       "address",
			IsRequired: true,
			Validator: func(value interface{}) bool {
				return govalidator.IsByteLength(value.(string), 1, 128)
			},
		},
		{
			Name:       "city",
			IsRequired: true,
			Validator: func(value interface{}) bool {
				return govalidator.IsByteLength(value.(string), 1, 128)
			},
		},
		{
			Name:       "zip",
			IsRequired: true,
			Validator: func(value interface{}) bool {
				return govalidator.IsByteLength(value.(string), 1, 12) && govalidator.IsNumeric(value.(string))
			},
		},
		{
			Name: "phone",
		},
		*/
	},
	OnInit: func(c Context, h *DataHolder) error {
		if c.IsAuthenticated {
			return ErrAlreadyAuthenticated
		}
		return nil
	},
}
/*
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(r)

	err = decrypt([]byte(d.Get(ctx, "password").([]uint8)), []byte(do.GetInput("password").(string)))
	if err != nil {
		ctx.PrintError(w, err, http.StatusInternalServerError)
		return
	}

	err = ctx.NewUserToken(d.id, Role(d.Get(ctx, "role").(string)))
	if err != nil {
		ctx.PrintError(w, err, http.StatusInternalServerError)
		return
	}

	ctx.Print(w, data)
}*/

func login(ctx Context, username string, password string) (Token, error) {
	data, err := getUser(ctx, username)
	if err != nil {
		return Token{}, err
	}

	if uintPass, ok := data["password"].([]uint8); ok {
		err = decrypt([]byte(uintPass), []byte(password))
		if err != nil {
			return Token{}, err
		}

		return newToken(username, Role(data["role"].(int64)), tokenKey)
	}

	return Token{}, ErrForbidden
}

func getUser(ctx Context, username string) (map[string]interface{}, error) {

	var dataHolder = User.New(ctx)
	dataHolder.key = User.NewKey(ctx, username)

	err := datastore.Get(ctx.Context, dataHolder.key, dataHolder)
	if err != nil {
		return nil, err
	}

	return dataHolder.UnsafeOutput(), nil
}

/*func LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(r)

	err = decrypt([]byte(d.Get(ctx, "password").([]uint8)), []byte(do.GetInput("password").(string)))
	if err != nil {
		ctx.PrintError(w, err, http.StatusInternalServerError)
		return
	}

	err = ctx.NewUserToken(d.id, Role(d.Get(ctx, "role").(string)))
	if err != nil {
		ctx.PrintError(w, err, http.StatusInternalServerError)
		return
	}

	ctx.Print(w, data)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(r).WithScopes(ScopeAdd)

	if ctx.IsAuthenticated {
		ctx.PrintError(w, ErrAlreadyAuthenticated, http.StatusInternalServerError)
		return
	}

	// Add user
	d, err := userEntity.FromForm(ctx)
	if err != nil {
		ctx.PrintError(w, err, http.StatusInternalServerError)
		return
	}

	ctx, key, err := userEntity.NewKey(ctx, d.Get(ctx, "email"))
	if err != nil {
		ctx.PrintError(w, err, http.StatusInternalServerError)
		return
	}

	key, err = userEntity.Add(ctx, key, d)
	if err != nil {
		ctx.PrintError(w, err, http.StatusInternalServerError)
		return
	}

	err = ctx.NewUserToken(d.id, Role(d.Get(ctx, "role").(string)))
	if err != nil {
		ctx.PrintError(w, err, http.StatusInternalServerError)
		return
	}

	ctx.Print(w, data)
}*/

var (
	ErrAlreadyAuthenticated = errors.New("already authenticated")
)
