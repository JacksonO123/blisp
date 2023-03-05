# Function documentation

`print`:
- Prints any number of parameters

`+`:
- Adds any number of parameters
- Parameters must be type `Int` or `Float`

`-`:
- Subtracts any number of parameters
- Parameters must be type `Int` or `Float`
- If one parameter is passed, `-` returns the opposite of the number

`*`:
- Multiplies any number of parameters
- Parameters must be type `Int` or `Float`

`/`:
- Divides any number of parameters
- Parameters must be type `Int` or `Float`

`^`:
- Two parameters be type `Int` or `Float`
- Returns first parameter to the exponent of the second parameter

`%`:
- Returns modulus of the two parameters
- Parameters must be type `Int` or `Float`

`eval`:
- Takes in any number of String parameters, and evaluates each parameter
- If there is only one parameter, the return value will be the return value of the evaluation

`var`:
- Takes two parameters, first is the variable name as `Ident` type, second is the value
- Value can be any type

`const`:
- Defines a variable similar to `var`
- Variable cannot be modified

`set`:
- Two parameters
	- First parameter is a variable
	- Second is the value
	- Value can be any type
- Three parameters
	- First parameter is the variable
	- Second parameter is the index or property of `Struct` or `List`
	- Third parameter is the value
	- Value can be any type

`free`:
- Takes one parameter of type `Ident`
- Frees the variable from memory

`type`:
- Takes one parameter of any type
- Returns the type of the parameter as a `String`

`get`:
- Takes two parameters, first is variable
- Variable is `Struct`
	- Second parameter is `Ident`
	- Returns property of type `Struct`
- Variable is `List`
	- Second parameter is `Int`
	- Returns value at index

`loop`:
- Three parameters
	- First parameter is `List`
		- Second parameter is value iterator of type `Ident`, value of each item in list
	- First parameter is `Int`
		- Second parameter is number iterator of type `Ident`, number from 0 to second parameter
	- Third parameter is `body`
- Four parameters
	- First parameter is `List`
		- Second parameter is value iterator of type `Ident`, value of each item in list
		- Third parameter is index iterator of type `Ident`, index of value in list
	- First parameter is `Int`
		- Second parameter is `Int`
		- Third parameter is number iterator of type `Ident`, number from first parameter to second parameter
	- Fourth parameter is `body`

`scan-line`:
- No parameters
	- Takes input from the command line
	- Returns value
- One parameters
	- First parameter is variable
	- Sets variable to value

`body`:
- Wraps body of code
- Example:
```lisp
(if (eq 2 2) (body
	(print "equal")
))
```

`eq`:
- Takes any number of parameters
- Parameters can be any type
- Returns `Bool` if parameters are equal

`append`:
- Takes two or more parameters
- First parameter is `List`
- Other parameters can be any type
- Appends parameters to list

`prepend`:
- Takes two or more parameters
- First parameter is `List`
- Other parameters can be any type
- Appends parameters to list

`concat`:
- Takes any number of parameters
- Parameters must be type `String`

`exit`:
- Takes one parameter
	- First parameter is exit code
- Default exit code is 0

`break`:
- No parameters
- Breaks from loop

`pop`:
- Takes one parameter of type `List`
- Pops and returns last value of list

`remove`:
- First parameter is type `List`
	- Second parameter is type `Int`
- First parameter is type `Struct`
	- Second parameter is type `Ident` that is a property on the first parameter
- Returns removed item

`len`:
- Takes one parameter of type `List`
- Returns length of list

`and`:
- Takes one or more parameters
- Parameters must be type `Bool`
- Returns weather the parameters are all true

`or`:
- Takes one or more parameters
- Parameters must be type `Bool`
- Returns weather one of the parameters are true

`not`:
- Takes one parameter of type `Bool`
- Inverts the value of the parameter

`func`:
- Takes 2 or more parameters
- First parameter is function name
- If function name is `_`, function is not saved, and function variable is returned
- Second and on (not including last parameter) are parameter names of type `Ident`
- Last parameter is `body`

`return`:
- Returns values from function
- Takes any number of parameters with any type

`parse`:
- Takes one parameter of type `String`
- Returns either `Int` or `Float` of parsed value

`<`:
- Takes two parameters of type `Int` or `Float`
- Returns `true` if first parameter is less than the second, false if not

`<=`:
- Takes two parameters of type `Int` or `Float`
- Returns `true` if first parameter is less than or equal to the second, false if not

`>`:
- Takes two parameters of type `Int` or `Float`
- Returns `true` if first parameter is greater than the second, false if not

`>=`:
- Takes two parameters of type `Int` or `Float`
- Returns `true` if first parameter is greater than or equal to the second, false if not

`read`:
- Takes one parameters of type `String`
- Returns file data as type `String`

`write`:
- Takes two parameters of type `String`
- First parameter is file name
- Second parameter is file value

`substr`:
- Two parameters
	- First parameter is type `String`
	- Second parameter is `Int`
	- Returns `String` of substring from 0 to first parameter
- Three parameters
	- First parameter is type `String`
	- Second parameter is type `Int`
	- Third parameter is type `Int`
	- Returns `String` of substring from first parameter to second parameter

`struct`:
- Takes even number of parameters (0, 2, 4, ...)
- Odd number parameters:
	- Type `Ident`
	- Key of the struct
- Even number parameters:
	- Any type
	- Value for previous parameter key

`.`:
- Takes two or more parameters
- First parameter is struct
- Second parameter is attribute on struct that is a function
- When calling method on struct, first parameter is struct reference
- 3rd parameter and on are passed to struct method

Struct Example:
```lisp
(struct key "value" number 12)
```

Json equivalent:

```json
{
	"key": "value",
	"number": 12
}
```
