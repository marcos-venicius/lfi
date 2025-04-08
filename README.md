# LFI

_LFI extends for "Log Filter", not for "Local File inclusion" :)_

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
