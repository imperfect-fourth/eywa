# eywa
A flexible ORM-style GraphQL client for building graphql queries dynamically.

**_This module is in pre-alpha phase._**

## Motivation
Go GraphQL client libraries exist on two extremes, neither of which allows you
to elegantly build queries dynamically. One extreme will have you define 
graphql queries as plain strings:

```go
...
query := `query GetUser($id: uuid!) { user(id: $id) { name } }`
variables := map[string]interface{}{"id": "1"}
var resp struct {
	 struct User {
		Name string
	}
}
client.Query(ctx, query, variables, &resp)
...
```
Now, if you want to get user by `name` instead of `id`, and/or select more 
fields, you either have to repeat the code with a new query string, or do some 
ugly string manipulation.

The second extreme is where instead of strings, you define queries as structs.
```go
...
var query struct {
    User struct {
        Name uuid.UUID `graphql:"name"`
    } `graphql:"user(id: $id)"`
}
...
```
This provides some additional type safety, but has a similar drawback—every 
query needs a new struct.

These concerns are non-issues for projects with only a handful queries, but 
they can seriously bloat larger codebases. 

## How does Eywa help?
With **Eywa**, you define your models **ONCE** as structs(as you do normally),
and then **flexibly** and **dynamically** build whatever query you need using 
ORM-style method chaining. The above query using eywa would look like:
```go
type User struct {
    ID   uuid.UUID `json:"id"`
    Name string    `json:"name"`
    Age  int       `json:"age"`
}

// to satisfy the Model interface
func (u User) ModelName() string {
    return "user"
}

q := GetUnsafe[User]().Where(
    Eq[User]("id", uuid.New()),
).Select("name")
resp, err := q.Exec(client)
```

For eg, creating a new query to get 5 users by `age` who are older than, say,
35 but younger than 50, and selecting the field `id` is as easy as:
```go
resp, err := GetUnsafe[User]().Where(
    And(
        Gt[User]("age", 35),
        Lt[User]("age", 50),
    ),
).Limit(5).Select("id", "age").Exec(client)
```

## `eywagen` and death to raw string literals
In the examples above, you may have noticed that the `Select` method takes raw 
strings as field names. This is prone to typos. `eywa` has a codegen tool for 
generating constants and functions for selecting fields.  
Install eywagen
```bash
go install github.com/imperfect-fourth/eywa/cmd/eywagen
```
Add `go:generate` comments to your code.
```go
//go:generate eywagen -types User -output user_fields.go
type User struct {
    ...
}
```
Run
```go 
go generate .
```
This will create a file `user_fields.go` in the same package
```go
//user_fields.go
package <package>

const User_Name string = "name"
const User_Age string = "age"
const User_ID string = "id"
```
Now, you can make the same query as:
```go
resp, err := Get[User]().Where(
    And(
        Gt(User_Age, 35),
        Lt(User_Age, 50),
    ),
).Limit(5).Select(
    User_ID,
    User_Age,
).Exec(client)
```

If a model has a relationship with another model, `eywagen` will generate a
function to select fields for that relationship. Eg.
```go
//go:generate eywagen -types User,Order -output model_fields.go
type User struct {
    ...
    Orders []Order `json:"orders"`
}
type Order struct {
    ID  uuid.UUID `json:"id"`
}

...

resp, err := Select[User]().Limit(5).Select(
    User_ID,
    User_Name,
    User_Orders(
        Order_ID,
    ),
).Exec(client)

//query GetUser {
//  user(limit: 5) {
//    id
//    name
//    orders {
//      id
//    }
//  }
//}
```


## Hasura support

|    | queries |mutations|order_by|distinct_on|limit|where|offset|relationships in queries|
|:---|:-------:|:-------:|:------:|:---------:|:---:|:---:|:----:|:-----------:|
| v2 |✅ **\***|    ❌   |   ✅   |    ✅     | ✅  | ✅  |  ✅  |    ❌       |
| v3 |    ✅   |   -     |   -    |     -     | ✅  | ✅  |  ✅  |    ❌       |

**\*** aggregate type queries not supported  
