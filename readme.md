# Commands orchestrator

A simple tools which allows to run set of shell commands in a script like manner but on steroids.

## Why?

An initial problem which this tool tries to solve is to run several subscripts in parallel.
This can easily be achieved by using bash `&` but you will not be able to wait until all background tasks will be
finished to run next loop.

## How?

Tool usage is:

```sh
  -u    Check if newer version is available and self-update
  -v    Print version end exit
  -vv   Verbose output

```

It takes config of the following format:

```
vars:
  test: 123
  sleep_time: 1

play:
  - steps:
      - sleep {{.sleep_time}}
      - ./examples/test.sh
      - exit 121

  - steps:
      - echo {{.test}}

  - steps:
      - sleep 2
      - ./examples/test.sh
      - ./examples/test.sh
      - exit 1
```

And executes steps one by one. Step is running all commands of single step in parallel. So in example above there will
be 3 parallel run on the first step and 4 for second step.
Output is piped into back into terminal.

## Output

In result run it will print table with run summary and also dump all output to temporary location. Example output:

```
NAME    START           END             DURATION        STATUS  COMMAND                 LOGS AT                                   
0_0     16:46:09        16:46:10        1.004608292s    success  sleep 1                /tmp/___go_build_main_go_log/0_0/full.log       
0_1     16:46:09        16:46:16        7.00972742s     success  ./examples/test.sh     /tmp/___go_build_main_go_log/0_1/full.log       
0_2     16:46:09        16:46:09        3.484781ms      failed   exit 121               /tmp/___go_build_main_go_log/0_2/full.log       
1_0     16:46:16        16:46:16        2.088552ms      success  echo 123               /tmp/___go_build_main_go_log/1_0/full.log       
2_0     16:46:16        16:46:18        2.003713021s    success  sleep 2                /tmp/___go_build_main_go_log/2_0/full.log       
2_1     16:46:16        16:46:23        7.00931349s     success  ./examples/test.sh     /tmp/___go_build_main_go_log/2_1/full.log       
2_2     16:46:16        16:46:23        7.010632844s    success  ./examples/test.sh     /tmp/___go_build_main_go_log/2_2/full.log       
2_3     16:46:16        16:46:16        3.790103ms      failed   exit 1                 /tmp/___go_build_main_go_log/2_3/full.log 
```

## Vars templating

It is possible to use templating for commands. In order to specify variables for templating define `vars` top level key
with set of values.
In the command it is required to add gotpl syntax expression to reference variable name.

For example:

```
vars:
    here_goes_keys: "value"

play:
    - steps:
        - echo {{.here_goes_keys}}
```

Will use value from `vars` section.

Another example is [here](./examples/test.yml).
