package scaffold

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarshalFlatWithoutCommentYaml(t *testing.T) {
	type Subject struct {
		Age   int    `yaml:"age"`
		Name  string `yaml:"name"`
		Title string `yaml:"title"`
	}

	s1 := Subject{
		Age:  22,
		Name: "John",
	}

	expected := stripLeadingNewline(`
age: 22
name: "John"
# title: ""
`)
	actual := generateYaml(s1)
	require.Equal(t, expected, actual)
}

func TestMarshalFlatWithSomeComment(t *testing.T) {
	type Subject struct {
		Age    int    `yaml:"age" comment:"the age of the user"`
		Name   string `yaml:"name" comment:"the name of the user"`
		Title  string `yaml:"title"`
		Active bool   `yaml:"active"`
	}

	s1 := Subject{
		Name:  "John",
		Title: "Dr",
	}

	expected := stripLeadingNewline(`
# age: 0 # the age of the user
name: "John" # the name of the user
title: "Dr"
# active: false
`)
	actual := generateYaml(s1)
	require.Equal(t, expected, actual)
}

func TestMarshalFlatWithSomeRequiredCommentYaml(t *testing.T) {
	type Subject struct {
		Age    int    `yaml:"age" comment:"required, the age of the user"`
		Name   string `yaml:"name" comment:"required, the name of the user"`
		Title  string `yaml:"title" comment:"the title of the user"`
		Active bool   `yaml:"active" comment:"the active status of the user"`
	}

	s1 := Subject{
		Name:  "John",
		Title: "Dr",
	}

	// note: "age" is rendered to yaml because of the word "required" in the comment
	expected := stripLeadingNewline(`
age: 0 # required, the age of the user
name: "John" # required, the name of the user
title: "Dr" # the title of the user
# active: false # the active status of the user
`)
	actual := generateYaml(s1)
	require.Equal(t, expected, actual)
}

func TestMarshalNestedStruct(t *testing.T) {
	type VeryNested struct {
		ID             string `yaml:"id" comment:"the id of the user"`
		ExpirationDate string `yaml:"expirationDate"` // attribute with no comment
		Department     string `yaml:"department" comment:"one of: business, development, research, AI"`
	}
	type Nested struct {
		Details VeryNested `yaml:"details"` // struct with no comment
		Name    string     `yaml:"name" comment:"the name of the user"`
	}
	type Subject struct {
		User   Nested `yaml:"user" comment:"required, user data"`
		Active bool   `yaml:"active" comment:"required, the active flag of the user"`
	}

	s1 := Subject{
		User: Nested{
			Details: VeryNested{
				ID: "123",
			},
			Name: "John",
		},
		Active: true,
	}

	expected := stripLeadingNewline(`
user: # required, user data
  details:
    id: "123" # the id of the user
#     expirationDate: ""
#     department: "" # one of: business, development, research, AI
  name: "John" # the name of the user
active: true # required, the active flag of the user
`)
	actual := generateYaml(s1)
	require.Equal(t, expected, actual)
}

func TestMarshalSlice(t *testing.T) {
	type Subject struct {
		Age             int      `yaml:"age" comment:"required, the age of the user"`
		Name            string   `yaml:"name" comment:"required, the name of the user"`
		Nicknames       []string `yaml:"nicknames"`
		Friends         []string `yaml:"friends"`
		FavouriteFoods  []string `yaml:"favouriteFoods" comment:"favourite foods of the user"`
		FavouriteColors []string `yaml:"favouriteColors" comment:"favourite colors of the user"`
	}
	s1 := Subject{
		Name:            "John",
		Friends:         []string{"Bob", "Alice"},
		FavouriteColors: []string{"Red", "Green", "Blue"},
	}

	expected := stripLeadingNewline(`
age: 0 # required, the age of the user
name: "John" # required, the name of the user
# nicknames:
#   -
friends:
  - "Bob"
  - "Alice"
# favouriteFoods: # favourite foods of the user
#   -
favouriteColors: # favourite colors of the user
  - "Red"
  - "Green"
  - "Blue"
`)
	actual := generateYaml(s1)
	require.Equal(t, expected, actual)
}

func TestMarshalNestedSlice(t *testing.T) {
	type Nested struct {
		Nicknames       []string `yaml:"nicknames"`
		Friends         []string `yaml:"friends"`
		FavouriteFoods  []string `yaml:"favouriteFoods" comment:"favourite foods of the user"`
		FavouriteColors []string `yaml:"favouriteColors" comment:"favourite colors of the user"`
	}
	type Subject struct {
		Age     int    `yaml:"age" comment:"required, the age of the user"`
		Name    string `yaml:"name" comment:"required, the name of the user"`
		Details Nested `yaml:"details" comment:"the details of the user"`
	}
	s1 := Subject{
		Name: "John",
		Details: Nested{
			Friends:         []string{"Bob", "Alice"},
			FavouriteColors: []string{"Red", "Green", "Blue"},
		},
	}

	expected := stripLeadingNewline(`
age: 0 # required, the age of the user
name: "John" # required, the name of the user
details: # the details of the user
#   nicknames:
#     -
  friends:
    - "Bob"
    - "Alice"
#   favouriteFoods: # favourite foods of the user
#     -
  favouriteColors: # favourite colors of the user
    - "Red"
    - "Green"
    - "Blue"
`)
	actual := generateYaml(s1)
	require.Equal(t, expected, actual)
}

func TestMarshalDoublyNestedSlice(t *testing.T) {
	type DoublyNested struct {
		Nicknames       []string `yaml:"nicknames"`
		Friends         []string `yaml:"friends"`
		FavouriteFoods  []string `yaml:"favouriteFoods" comment:"favourite foods of the user"`
		FavouriteColors []string `yaml:"favouriteColors" comment:"favourite colors of the user"`
	}
	type Nested struct {
		Public  DoublyNested `yaml:"public"`
		Private DoublyNested `yaml:"private" comment:"private details of the user"`
	}
	type Subject struct {
		Age     int    `yaml:"age" comment:"required, the age of the user"`
		Name    string `yaml:"name" comment:"required, the name of the user"`
		Details Nested `yaml:"details" comment:"the details of the user"`
	}
	s1 := Subject{
		Name: "John",
		Details: Nested{
			Public: DoublyNested{
				Friends:         []string{"Bob", "Alice"},
				FavouriteColors: []string{"Red", "Green", "Blue"},
			},
			/*
				Private: DoublyNested{
					Nicknames:      []string{"LazyCoyote", "FunnyBunny"},
					FavouriteFoods: []string{"Pizza", "Cola", "Nachos"},
				},
			*/
		},
	}

	expected := stripLeadingNewline(`
age: 0 # required, the age of the user
name: "John" # required, the name of the user
details: # the details of the user
  public:
#     nicknames:
#       -
    friends:
      - "Bob"
      - "Alice"
#     favouriteFoods: # favourite foods of the user
#       -
    favouriteColors: # favourite colors of the user
      - "Red"
      - "Green"
      - "Blue"
#   private: # private details of the user
#     nicknames:
#       -
#     friends:
#       -
#     favouriteFoods: # favourite foods of the user
#       -
#     favouriteColors: # favourite colors of the user
#       -
`)
	actual := generateYaml(s1)
	require.Equal(t, expected, actual)
}

func stripLeadingNewline(val string) string {
	return strings.TrimLeft(val, "\n")
}
