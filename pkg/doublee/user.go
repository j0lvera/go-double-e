package doublee

// NOTE: doesn't make sense to have this here, but it's a good example of how to use the SQLC queries
//func (c *db.Client) CreateUser(ctx context.Context, email, password string) (db.User, error) {
//	user, err := c.queries.CreateUser(ctx, db.CreateUserParams{
//		Email:    email,
//		Password: password,
//	})
//	if err != nil {
//		return db.User{}, fmt.Errorf("error creating user: %w", err)
//	}
//
//	return user, nil
//}
