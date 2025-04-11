# QUANG

`Quang` extends for **Query Lang**.

It's a built-in language that will serve as the tool language to make queries and filters.

The language should be able to have a syntax like:

```lua
(size gt 0 and size lt 1024 and method eq :get and status eq 200) or (method eq :post and size eq 0 and status eq 204)
```

- Symbols (name, size, asdflk)
- Integer (positive integers only)
- Atom (:name)
- Keywords (and, or)
    - Operators (gt, gte, lt, lte, eq, ne)
- Parenthesis

