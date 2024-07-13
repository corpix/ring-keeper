# ring-keeper

A cli tool to keep specified amount of files in directory and/or amount of file of specified size, where older is discarded.

## configuration file

Run tool with `--config` / `-c` flag to pass a config file path. It uses `config.json` by default (from CWD).

```json
[{
    "directory": "./test",
    "max-amount": 10,
    "max-size": "1MB",
    "delay": 60
}]
```

- files sorted by date, older may get discarded
- directory is an abosulte or relative path to the directory to watch
- max-amount is a maximal amount of files, optional
- max-size specified in human readable format, optional
- delay between shecks in seconds, optional
