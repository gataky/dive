package ui

// getHelpContent returns the formatted help text for gjson syntax
func getHelpContent() string {
	return `[yellow::b]gjson Path Syntax Reference[::-]

[white::b]BEGINNER - Basic Paths[::-]

[gray]Access object fields:[-]
  name              Get top-level field
  user.name         Get nested field
  user.address.city Deep nested access

[gray]Access array elements:[-]
  users.0           First element (index 0)
  users.1.name      Field from second element
  users.-1          Last element
  users.-2          Second to last element

[gray]Array length/count:[-]
  users.#           Number of elements in array
  items.#           Count items


[white::b]INTERMEDIATE - Queries & Wildcards[::-]

[gray]Wildcard queries (get all):[-]
  users.#.name           All names from users array
  items.#.price          All prices from items
  data.*.value           All values (any key)

[gray]Conditional queries:[-]
  users.#(age>21)#                Count users over 21
  users.#(active==true)#          Count active users
  users.#(age>=18 && age<=65)#    Count in age range
  items.#(price<100)#             Count items under $100

[gray]Get filtered results:[-]
  users.#(age>21)#.name           Names of users over 21
  items.#(inStock==true)#.price   Prices of in-stock items

[gray]Escape special characters:[-]
  user.first\.name       Field literally named "first.name"
  data.key\ with\ spaces Field name with spaces
  obj.key\*special       Escape wildcards in key names


[white::b]ADVANCED - Modifiers & Complex Queries[::-]

[gray]Modifiers (use with @):[-]
  @reverse              Reverse array order
  @ugly                 Compact JSON (no formatting)
  @pretty               Pretty-print JSON with indent
  @this                 Current element in context
  @valid                Check if JSON is valid
  @flatten              Flatten nested arrays
  @join                 Join array elements
  @keys                 Get object keys as array
  @values               Get object values as array

[gray]Multi-path queries (get multiple fields):[-]
  {name,age}                   Get name and age
  {name,email,address.city}    Get multiple, including nested
  users.#.{name,age}           Multiple fields from all users

[gray]Array slicing:[-]
  users.0:3         First 3 elements (0, 1, 2)
  users.2:5         Elements at index 2, 3, 4
  users.-3:         Last 3 elements
  users.:-2         All except last 2

[gray]Query operators:[-]
  ==  !=            Equal, not equal
  <   <=            Less than, less or equal
  >   >=            Greater than, greater or equal
  %                 Pattern match (e.g., name%"*John*")
  !                 Logical NOT

[gray]Complex nested queries:[-]
  users.#(orders.#(total>100)#>0)#    Users with orders over $100
  data.#(tags.#(=="important")#>0)#   Items tagged "important"

[gray]Combining modifiers:[-]
  users.#.age|@reverse           Ages in reverse order
  items.#.name|@join             Join all names
  data.@keys                     Get all keys from object


[white::b]Examples with Real Data[::-]

[gray]Given: {"users":[{"name":"Alice","age":25},{"name":"Bob","age":30}]}[-]

  users.#               → 2
  users.0.name          → "Alice"
  users.#.name          → ["Alice","Bob"]
  users.#(age>26)#      → 1
  users.#(age>26)#.name → ["Bob"]
  {users.0.name,users.1.age} → {"name":"Alice","age":30}
`
}
