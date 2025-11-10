# Hermes
Hermes: Swift coding for qualitative research

_Praise be the pure dry functions that guide us_

## First setup

### Commit Hooks
Enables git commit hooks. Linting & unit testing etc
```shell
makefile setup
```

## Run tests
```shell
make tests
```

## TODO
- Add unit tests for struct param/query binding and struct default value setting (specifically around embedded structs)
- Add unit tests for cross-aggregate events (within same registry) - e.g., DeletedCode event affecting multiple Files
