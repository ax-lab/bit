package types

/*

	=> Any `(*)`

		- Can hold any value. How it does so is up to the implementation.
		- By itself is useless, but can be paired with RTTI and downcast
		  operations to be used at runtime.
		- `Any + T = Any + T` (mostly for type hinting, same as `Any` still)

	=> None `(none)`

		- Represent a lack of value (e.g. missing else branch). Has zero size.
		- It is a value, but represents the lack of one.
		- Sum with another type to make an optional type.
		- `None + None = None`, so `Option<T> + Option<U> = Option<T + U>`

	=> Never `(!)`

		- Represents a never occurring value (e.g. infinite loop, non-returning
		  function, exit / halt / throw operations).
		- Has zero size and no possible value.
		- `Never + T = T`
		- At compile time, it signals that a branch is impossible.
		- In runtime, this would be a panic.

	=> Unit `( )`

		- Global single value zero-sized type.
		- `Unit + Unit = Unit`.
		- None and Unit are similar and, in practice, could be the same type.

	=> Unknown `(?)`

		- Represents a type that is not yet fully known.
		- Has no size, no value, and cannot be output.
		- Can be used anywhere and accepts any operation.
		- Must be eventually solved to a concrete type.

	=> Invalid `(?!)`

		- Flags a type with an error.
		- Has no size, no value, and cannot be output.
		- Invalid variant is available for all types.
			- Can be used in place of the type.
			- In practice, this works as a flag on the given type.
			- Non-specific `Invalid` type can be used anywhere and accepts
			  any operation.

	=> Named

		- A named type creates a functionally equivalent but different type.
			- The named type is effectively a copy.
			- It cannot be used as the source type by default.
			- Conversion between the source and named type is zero cost.
		- This is not the same as an alias.
		- No two types can have the same name, but a name is scoped to the
		  source file and any additional namespaces.
		- Any named type has a canonical and unique full name which includes
		  its scope.

	=> Enum

		- Tagged union type.
		- Each branch is its own type.
		- Each branch can be a tuple or struct.
		- Allows all branches to share a base struct.

	=> Sum type

		- Result of T + U
		- Implementation may be an Enum


	Type Operations
	===============

	=> Set operations: union, intersect, difference

	=> Reduce(T, U) =
		-> U if T is a base type of U
		-> T if U is a base type of T
		-> Otherwise (?!) + T + U

	=> Common(T, U) =
		-> T if T is a base type of U
		-> U if U is a base type of T
		-> Otherwise (*)
*/
