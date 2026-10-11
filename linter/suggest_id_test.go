package linter

import (
	"testing"
)

func TestExtractIDRoot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want string
	}{
		{name: "userId", want: "User"},
		{name: "order_id", want: "Order"},
		{name: "order_ids", want: "Order"},
		{name: "requestID", want: "Request"},
		{name: "product_id_v2", want: "Product"},
		{name: "id_parent", want: "Parent"},
		{name: "PID", want: "P"},
		{name: "userID", want: "User"},
		{name: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := extractIDRoot(tt.name); got != tt.want {
				t.Errorf("extractIDRoot(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestCapitalizeFirst(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{in: "", want: ""},
		{in: "user", want: "User"},
		{in: "User", want: "User"},
		{in: "pID", want: "PID"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()

			if got := capitalizeFirst(tt.in); got != tt.want {
				t.Errorf("capitalizeFirst(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestExtractDomainFromFunction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		funcName string
		want     string
	}{
		{funcName: "AddRule", want: "Rule"},
		{funcName: "GetUser", want: "User"},
		{funcName: "DeleteOrder", want: "Order"},
		{funcName: "ProcessPayment", want: "Payment"},
		{funcName: "DataPoint", want: "DataPoint"},
		{funcName: "Finding", want: "Finding"}, // "Find"+"ing": lowercase remainder is not a noun
		{funcName: "Clone", want: ""},          // verb-only: GenericID fallback
		{funcName: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			t.Parallel()

			if got := extractDomainFromFunction(tt.funcName); got != tt.want {
				t.Errorf("extractDomainFromFunction(%q) = %q, want %q", tt.funcName, got, tt.want)
			}
		})
	}
}

func TestSuggestIDTypeName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		param      string
		parent     string
		want       string
	}{
		{param: "userId", parent: "", want: "UserID"},
		{param: "product_id", parent: "", want: "ProductID"},
		{param: "product_id_v2", parent: "", want: "ProductID"},
		{param: "id", parent: "GetUser", want: "UserID"},
		{param: "id", parent: "AddRule", want: "RuleID"},
		{param: "id", parent: "Clone", want: "GenericID"},
		{param: "id", parent: "", want: "GenericID"},
		{param: "userIDs", parent: "", want: "UserID"},
	}

	for _, tt := range tests {
		t.Run(tt.param+"@"+tt.parent, func(t *testing.T) {
			t.Parallel()

			if got := suggestIDTypeName(tt.param, tt.parent); got != tt.want {
				t.Errorf("suggestIDTypeName(%q, %q) = %q, want %q", tt.param, tt.parent, got, tt.want)
			}
		})
	}
}

func TestSuggestIDBrand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		param  string
		parent string
		want   string
	}{
		{param: "userId", parent: "", want: "UserBrand"},
		{param: "id", parent: "GetUser", want: "UserBrand"},
		{param: "id", parent: "Clone", want: "GenericBrand"},
	}

	for _, tt := range tests {
		t.Run(tt.param+"@"+tt.parent, func(t *testing.T) {
			t.Parallel()

			if got := suggestIDBrand(tt.param, tt.parent); got != tt.want {
				t.Errorf("suggestIDBrand(%q, %q) = %q, want %q", tt.param, tt.parent, got, tt.want)
			}
		})
	}
}

func TestBuildTypeDefinition(t *testing.T) {
	t.Parallel()

	want := "type UserID = id.ID[UserBrand, string]\n\ntype UserBrand struct{}"
	if got := buildTypeDefinition("UserID", "UserBrand", "string"); got != want {
		t.Errorf("buildTypeDefinition() = %q, want %q", got, want)
	}
}

func TestBuildCodeSnippet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		typeStr    string
		kind       primitiveLocation
		parentName string
		want       string
	}{
		{
			name: "param with parent", typeStr: "string", kind: locationParam, parentName: "getUser",
			want: "func getUser(userId string)",
		},
		{
			name: "field with parent", typeStr: "int", kind: locationField, parentName: "Order",
			want: "type Order struct { orderId int }",
		},
		{
			name: "return with parent", typeStr: "string", kind: locationReturn, parentName: "nextOrder",
			want: "func nextOrder() (orderId string, ...)",
		},
		{
			name: "return without parent", typeStr: "string", kind: locationReturn,
			want: "(orderId string)",
		},
		{
			name: "param without parent falls back to bare declaration", typeStr: "string",
			kind: locationParam, want: "userId string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := buildCodeSnippet("orderId", tt.typeStr, tt.kind, tt.parentName); got != tt.want {
				t.Errorf("buildCodeSnippet() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPrimitiveIDRank(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kind primitiveLocation
		want severityRank
	}{
		{name: "userId", kind: locationParam, want: rankHigh},
		{name: "userId", kind: locationField, want: rankMedium},
		{name: "orderId", kind: locationReturn, want: rankHigh},
		{name: "configID", kind: locationParam, want: rankMedium},
		{name: "configID", kind: locationField, want: rankLow},
		{name: "widgetID", kind: locationParam, want: rankLow},
		{name: "tokenID", kind: locationParam, want: rankHigh},
	}

	for _, tt := range tests {
		t.Run(tt.name+"@"+string(tt.kind), func(t *testing.T) {
			t.Parallel()

			if got := primitiveIDRank(tt.name, tt.kind); got != tt.want {
				t.Errorf("primitiveIDRank(%q, %s) = %v, want %v", tt.name, tt.kind, got, tt.want)
			}
		})
	}
}

func TestSeverityRankMapping(t *testing.T) {
	t.Parallel()

	tests := []struct {
		rank severityRank
		want string
	}{
		{rank: rankHigh, want: "error"},
		{rank: rankMedium, want: "warning"},
		{rank: rankLow, want: "info"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()

			if got := string(tt.rank.severity()); got != tt.want {
				t.Errorf("severity() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCreatePrimitiveIDMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		want   string
	}{
		{
			name: "compound param with parent",
			want: "Replace parameter 'userId' (string) with strong ID type UserID in getUser",
		},
		{
			name: "standalone id return without parent",
			want: "Change type of return value 'id' from string to strong ID type GenericID",
		},
		{
			name: "field kind",
			want: "Replace field 'orderId' (int) with strong ID type OrderID in Order",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var got string
			switch tt.name {
			case "compound param with parent":
				got = createPrimitiveIDMessage("userId", "string", "UserID", "getUser", locationParam)
			case "standalone id return without parent":
				got = createPrimitiveIDMessage("id", "string", "GenericID", "", locationReturn)
			case "field kind":
				got = createPrimitiveIDMessage("orderId", "int", "OrderID", "Order", locationField)
			}

			if got != tt.want {
				t.Errorf("createPrimitiveIDMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}
