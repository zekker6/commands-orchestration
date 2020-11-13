# Commands orchestrator

A simple tools which allows to run set of shell commands in a script like manner but on steroids.

## Why?

An initial problem which this tool tries to solve is to run several subscripts in parallel. 
This can easily be achieved by using bash `&` but you will not be able to wait until all background tasks will be finished to run next loop.

## How?

Tool usage is:
```sh
2020/11/13 22:48:07 Specify config path
2020/11/13 22:48:07 Usage: co [config path]
```

It takes config of the following format:
```
play:
  - steps:
      - sleep 10
      - ./examples/test.sh
      - exit 121


  - steps:
      - sleep 2
      - ./examples/test.sh
      - ./examples/test.sh
      - exit 1
```

And executes steps one by one by running all commands of single step in parallel. So in example above there will be 3 parallel run on the first step and 4 for second step.
Output is piped into back into terminal.

## Output

In result run it will print table with run summary and also dump all output to temporary location. Example output:
```
NAME    START           END             DURATION        STATUS  COMMAND                 LOGS AT                                   
0_0     22:50:57        22:51:07        10.005969051s   success  sleep 10               /tmp/___go_build_main_go_log/0_0/full.log       
0_1     22:50:57        22:51:04        7.011305948s    success  ./examples/test.sh     /tmp/___go_build_main_go_log/0_1/full.log       
0_2     22:50:57        22:50:57        3.706375ms      failed   exit 121               /tmp/___go_build_main_go_log/0_2/full.log       
1_0     22:51:07        22:51:09        2.007209033s    success  sleep 2                /tmp/___go_build_main_go_log/1_0/full.log       
1_1     22:51:07        22:51:14        7.015508958s    success  ./examples/test.sh     /tmp/___go_build_main_go_log/1_1/full.log       
1_2     22:51:07        22:51:14        7.015408134s    success  ./examples/test.sh     /tmp/___go_build_main_go_log/1_2/full.log       
1_3     22:51:07        22:51:07        6.203645ms      failed   exit 1                 /tmp/___go_build_main_go_log/1_3/full.log
```