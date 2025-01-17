# `go-double-e`

A simple double-entry web API written in Go.

## Notes

Handling nullable parameters

The [sqlc documentation explains how to handle the nullable parameters](https://docs.sqlc.dev/en/stable/howto/named_parameters.html#nullable-parameters) using `sqlc.narg` and gives an example.

```sql
-- name: UpdateAuthor :one
UPDATE author
SET
 name = coalesce(sqlc.narg('name'), name),
 bio = coalesce(sqlc.narg('bio'), bio)
WHERE id = sqlc.arg('id')
RETURNING *;
```

However, this only works with default sql driver.

To make it work with pgx, we need to do two things:
- add postgres type casting additionally to the `sqlc.narg` parameter
- set the `emit_result_struct_pointers: true` in the `sqlc.yaml` file

The change in the syntax is as follows:
```go
// from this
type UpdateParams struct {
	Name string `json:"name"`
}

params := UpateParams{
	Name: req.Name,
}

// to this

type UpdateParams struct {
    Name *string `json:"name,omitempty"`
}

name := pgtype.Text{
	Valid: req.Name != nil,
}

if req.Name != nil {
    name = *req.Name
}

params := UpdateParams{
	Name: name,
}

// using the `ptrValue` helper function for convenience 
type UpdateParams struct {
    Name *string `json:"name,omitempty"`
}

val, valid := ptrValue(req.Name)
    params := UpdateParams{
        Name: pgtype.Text{
        String: val,
        Valid: valid,
    },
}
```
