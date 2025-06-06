# LFI

_LFI extends for "Log Filter", not for "Local File inclusion" :)_

**And it uses [Quang](https://github.com/marcos-venicius/quang) as built-in query language.**

## Installing

```bash
go install github.com/marcos-venicius/lfi@latest
```

![image](https://github.com/user-attachments/assets/c9e68087-1bd2-4183-89af-f405cbf5ec74)

## The tool

**For now, this tool can only parse kong logs format.**

We have some initial options:

```bash
Usage of ./lfi:
  -de
        disable error lines output
  -f string
        format the log in a specific way (default "%time %ip %method %resource %version %status %size %host %agent")
  -q string
        provide any valid filter using quang syntax https://github.com/marcos-venicius/quang.
        available variables: time, ip, method, resource, version, status, size, host, agent.
        available method atoms :get, :post, :delete, :patch, :put, :options.
```

# Documentation

Here I'll documentate all the features and how to use it.

### Formatting

We have the following tokens to format:

- labels `%time %ip %method %resource %version %status %size %host %agent`.
    - `%time` display the log date and time
    - `%ip` display the log ip
    - `%method` display the request method
    - `%resource` display the request url path resource
    - `%version` display the http version
    - `%status` display the response status code
    - `%size` display the size of the response
    - `%host` display the host
    - `%agent` display the user agent

To add strings, you can just use `'this is a string'`. To escape them, you can do `'this is \'my string\''`.

Line breaks with `\n`. Tabs with `\t`. And you can Add as many spaces as you want.

### Query building

This tool is using [Quang](https://github.com/marcos-venicius/quang) as a query builder, so, you can read more in the docs.

We have some available variables for the logs, they are:

- `ip: string`
- `time: string`
- `method: quang.AtomType`
- `host: string`
- `resource: string`
- `version: string`
- `status: quang.IntegerType`
- `size: quang.IntegerType`
- `user: string`

We have some available atoms for the method: `:get, :post, :delete, :patch, :put, :options`.

> [!WARNING]
> The documentation bellow is from [Quang](https://github.com/marcos-venicius/quang), it may change in the future
> To a more precise documentation, checkout [Quang](https://github.com/marcos-venicius/quang) repository.

**Data Types**

| name     | supported | format        | description                     |
| -------- | --------- | ------------- | ------------------------------- |
| Integers | yes       | `[0-9]+`      | golang 64bit signed integers    |
| Atoms    | yes       | `:[a-zA-Z_]+` | it works like enumerators       |
| String   | yes       | `'.*'`        | you can scape string with `\'`  |
| Boolean  | yes       | `true\|false` |                                 |
| Floats   | yes       | `\d+\.\d*`    | golang 64bit floats             |

**Keywords**

| name     | description   | usage            |
| -------- | ------------- | ---------------- |
| and      | logical and   | `true and true`  |
| or       | logical or    | `true or false`  |
| true     | boolean true  | `alive eq true`  |
| false    | boolean false | `alive eq false` |

**Operators**

| name     | description                                                                                        | example                 |
| -------- | -------------------------------------------------------------------------------------------------- | ----------------------- |
| eq       | check if `a` is equal to `b`. strict types. (Integers, Strings, Booleans, Nils, Floats, Atoms)     | `a eq b`                |
| ne       | check if `a` is not equal to `b`. strict types. (Integers, Strings, Booleans, Nils, Floats, Atoms) | `a ne b`                |
| lt       | check if `a` is less than `b`. strict types. (Integers, Strings)                                   | `a lt b`                |
| gt       | check if `a` is greater than `b`. strict types. (Integers, Strings)                                | `a gt b`                |
| lte      | check if `a` is less than or equal to `b`. strict types. (Integers, Strings)                       | `a lte b`               |
| gte      | check if `a` is greater than or equal to `b`. strict types. (Integers, Strings)                    | `a gte b`               |
| reg      | check if `a` matches pattern `b`. `b` accepts valid regex. `a` should be a string                  | `a reg b`               |

**Basic syntax**

```elixir
(method eq :put or method eq :options) and resource reg '^/dashboard/user/[a-z0-9]{32}/info' and size gt 0
```

