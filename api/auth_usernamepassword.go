package api

import (
	"errors"
	"github.com/asaskevich/govalidator"
	"net/http"
)

var PasswordField = &Field{
	Name:       "password",
	IsRequired: true,
	NoIndex:    true,
	Validator: func(value interface{}) bool {
		return govalidator.IsByteLength(value.(string), 6, 128)
	},
	TransformFunc: FuncHashTransform,
}

var User = &Entity{
	Name:      "user",
	Protected: true,
	Rules: Rules{
		Add: Guest,
	},
	Fields: []*Field{
		{
			Name: "email",
			Rules: Rules{
				Write: Admin,
			},
			IsRequired: true,
			Validator: func(value interface{}) bool {
				return govalidator.IsEmail(value.(string))
			},
		},
		{
			Name:    "role",
			NoIndex: true,
			Rules: Rules{
				Write: Admin,
			},
			DefaultValue: 1,
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
}

var (
	ErrAlreadyAuthenticated = errors.New("already authenticated")
)
