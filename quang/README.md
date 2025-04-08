# QUANG

`Quang` extends for **Query Lang**.

It's a built-in language that will serve as the tool language to make queries and filters.

The language should be able to have a syntax like:

```py
(size > 0 and size < 1024 and method = :get: and status = 200) or (method = :post: and size = 0 and status = 204)
```

This will have 
