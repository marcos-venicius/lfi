# LFI

_LFI extends for "Log Filter", not for "Local File inclusion" :)_

> [!WARNING]
> I already have an experimental version of Quang, which is an internal language used to make queries in the logs
> You can access by checking out [this branch](https://github.com/marcos-venicius/lfi/tree/feat/quang-experiment)

_**The following version, for now, is only available in the branch [feat/quang-experiment](https://github.com/marcos-venicius/lfi/tree/feat/quang-experiment)**_

https://github.com/user-attachments/assets/63e2102c-4dbb-48c0-ada2-6a12e50fe6f7

**It'll be in the main soon with some improvements after all tests**

If you wanna test this "beta" version, you can download as follows:

```bash
go install github.com/marcos-venicius/lfi@0.1.0.beta
```

## Installing

```bash
go install github.com/marcos-venicius/lfi@latest
```

## The tool

**For now, this tool can only parse kong logs format.**

We have some initial options:

```bash
  -de
        disable error lines output
  -f string
        format the log in a specific way (default "%time %ip %method %resource %version %status %size %host %agent")
  -fi string
        filter by a specific ip
  -fm string
        filter by a specific method
  -fs int
        filter by a specific status code. when -1 the filter is not used (default -1)
  -nefm string
        filter by logs where the method is not equal to the provided one
  -nefs int
        filter by logs where the status code is not equal to the provided one. when -1 the filter is not used (default -1)
```

I'll add more options later, like:

- [ ] allow to specify a list of values to each filter argument
- [ ] allow the user to search by a pattern in the url
- [ ] allow the user to search by a pattern in the user agent
- [ ] allow the user to colorize the output
- [ ] allow the user to search by status codes greater than/greater than or equal/less than/less than or equal
- [ ] allow the user to search by size equal/not equal/greater than/greater than or equal/less than/less than or equal
- [ ] allow the user to search by a host equal/not equal
- [ ] allow the user to format the date time

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
