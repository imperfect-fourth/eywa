package eywatest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imperfect-fourth/eywa"
	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	requests := map[string]struct {
		query    []*eywa.Field[QueryRoot]
		request  string
		response QueryRoot
	}{
		"simple_query": {
			[]*eywa.Field[QueryRoot]{
				QueryRoot_Users(User_ID),
			},
			"query simple_query {\nusers{\nid\n}\n}",
			QueryRoot{Users: []User{{ID: 1}}},
		},
		"multiple_fields": {
			[]*eywa.Field[QueryRoot]{
				QueryRoot_Users(User_ID, User_Username),
			},
			"query multiple_fields {\nusers{\nid\nusername\n}\n}",
			QueryRoot{Users: []User{{ID: 1, Username: "user1"}, {ID: 2, Username: "user2"}}},
		},
		"query_with_args": {
			[]*eywa.Field[QueryRoot]{
				eywa.WithArgs(
					QueryRoot_Users(User_ID),
					arguments_Limit(3),
					arguments_Offset(2),
				),
			},
			"query query_with_args {\nusers(limit: 3, offset: 2){\nid\n}\n}",
			QueryRoot{Users: []User{{ID: 3}, {ID: 4}, {ID: 5}}},
		},
		"fields_with_args": {
			[]*eywa.Field[QueryRoot]{
				QueryRoot_Users(
					User_ID,
					eywa.WithArgs(
						User_SubField(UserSubField_ID),
						arguments_Limit(1),
						arguments_Offset(2),
					)),
			},
			"query fields_with_args {\nusers{\nid\nsub_field(limit: 1, offset: 2){\nid\n}\n}\n}",
			QueryRoot{Users: []User{{ID: 3, SubField: []UserSubField{{ID: 3}}}}},
		},
		"multiple_queries": {
			[]*eywa.Field[QueryRoot]{
				QueryRoot_Users(User_ID),
				QueryRoot_NotUsers(NotUser_ID),
			},
			"query multiple_queries {\nusers{\nid\n}\nnot_users{\nid\n}\n}",
			QueryRoot{Users: []User{{ID: 1}}, NotUsers: []NotUser{{ID: 1}}},
		},
		"enum_arg": {
			[]*eywa.Field[QueryRoot]{
				eywa.WithArgs(QueryRoot_NotUsers(NotUser_ID),
					arguments_EnumArg(enumValue),
				),
			},
			"query enum_arg {\nnot_users(enum_arg: enum_value){\nid\n}\n}",
			QueryRoot{NotUsers: []NotUser{{ID: 1}}},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rb, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(400)
			w.Write(fmt.Append([]byte("bad request: "), err.Error()))
			return
		}
		var gqlreq eywa.Request
		if err = json.Unmarshal(rb, &gqlreq); err != nil {
			w.WriteHeader(400)
			w.Write(fmt.Append([]byte("bad request: "), err.Error()))
			return
		}

		if _, ok := requests[gqlreq.OperationName]; !ok {
			w.WriteHeader(400)
			fmt.Fprintf(w, "bad request: operation %s not found", gqlreq.OperationName)
			return
		}
		if gqlreq.Query != requests[gqlreq.OperationName].request {
			w.WriteHeader(400)
			w.Write(fmt.Appendf([]byte{}, "invalid query. expected: %s, got: %s", requests[gqlreq.OperationName].request, gqlreq.Query))
			return
		}

		respBytes, err := json.Marshal(eywa.Response[QueryRoot]{
			Data: requests[gqlreq.OperationName].response,
		})
		assert.NoError(t, err)
		w.Write(respBytes)
	}))
	defer server.Close()

	client := eywa.NewClient[QueryRoot, MutationRoot](server.URL, nil)

	for o, r := range requests {
		t.Run(o, func(t *testing.T) {
			resp, err := client.Query(
				t.Context(),
				o,
				r.query...,
			)
			if assert.NoError(t, err) {
				assert.Equal(t, r.response, resp.Data)
			}
		})
	}
}
