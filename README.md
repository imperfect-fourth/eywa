# eywa
A flexible ORM-style GraphQL client for building graphql queries dynamically.

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
    ID   uuid.UUID `graphql:"id"`
    Name string    `graphql:"name"`
    Age  int       `graphql:"age"`
}

// to satisfy the Model interface
func (u *User) ModelName() string {
    return "user"
}

q := Query(&User{}).Select("name").Where(
    Comparisons: Comparison{
        "id": {Eq: uuid.New()},
    },
)
resp, err := q.Exec(client)
```

For eg, creating a new query to get 5 users by `age` who are older than, say, 35 but
younger than 50, and selecting the field `id` is as easy as:
```
resp, err := Query(&User{}).Select("id").Where(
    Comparisons: Comparison{
        "age": {
            Gt: 35,
            Lt: 50,
        },
    },
).Limit(5).Exec(client)
```

## Hasura support

|    | queries |mutations|order_by|distinct_on|limit|where|offset|relationships in queries|
|:---|:-------:|:-------:|:------:|:---------:|:---:|:---:|:----:|:-----------:|
| v2 |✅ **\***|    ❌   |   ✅   |    ✅     | ✅  | ✅  |  ✅  |    ❌       |
| v3 |    ✅   |   -     |   -    |     -     | ✅  | ✅  |  ✅  |    ❌       |

**\*** aggregate type queries not supported  
