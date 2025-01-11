package server

import (
	is_ "github.com/matryer/is"
	"testing"
)

func TestMapNonZeroFields(t *testing.T) {
	is := is_.New(t)

	t.Run("interfaces with the same shape", func(t *testing.T) {
		// same shape
		// same amount of fields
		// same order

		type shape struct {
			Id          string
			Name        string
			Description string
		}

		source := shape{
			Id:          "123",
			Name:        "John",
			Description: "A person",
		}

		dest := shape{
			Id:          "456",
			Name:        "Doe",
			Description: "Another person",
		}

		MapNonZeroFields(&source, &dest)

		is.Equal(source.Id, "123")
		is.Equal(source.Name, "John")
		is.Equal(source.Description, "A person")
	})

	t.Run("interfaces with zeros", func(t *testing.T) {
		type shape struct {
			Id          string
			Name        string
			Description string
		}

		source := shape{
			Id:          "",
			Name:        "John",
			Description: "",
		}

		dest := shape{
			Id:          "456",
			Name:        "Doe",
			Description: "Person",
		}

		MapNonZeroFields(&source, &dest)

		is.Equal(dest.Id, "456")
		is.Equal(dest.Name, "John")
		is.Equal(dest.Description, "Person")
	})
}
