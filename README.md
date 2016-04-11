Order Processing System
------


### The source file tree of the project

        .
        ├── config                            // configuration
        │   ├── database.gcfg
        │   ├── log.gcfg
        │   └── service.gcfg
        ├── main.go                           // The entry of the service
        ├── process
        │   ├── consumer
        │   │   └── consumer.go
        │   ├── db                            // database
        │   │   └── db.go
        │   ├── diagnostic                    // diagnostic
        │   │   └── diagnostic.go
        │   ├── env                           // environment
        │   │   └── env.go
        │   ├── model                         // the model of service
        │   │   ├── cluster                   // cluster management
        │   │   │   └── cluster.go
        │   │   ├── order                     // order definition
        │   │   │   └── order.go
        │   │   ├── pipeline                  // processing logic
        │   │   │   ├── job.go
        │   │   │   ├── manager.go
        │   │   │   ├── pipeline.go
        │   │   │   └── task_handler.go
        │   │   └── transfer                  // job transfer
        │   │       └── transfer.go
        │   ├── service                       // the controller of the service
        │   │   └── order_process_service.go
        │   └── util                          // util
        │       └── util.go
        └── README.md
        
                └── README.md


### How one order will be processed in Order Processing Service?

> Each order will be dispatched to one pipeline

![image](http://img.blog.csdn.net/20160410212318322 "order prossing system")

### How does one service live in the cluster?

> One service can be leader or follower.

![image](http://img.blog.csdn.net/20160410213116856 "order prossing system")

### How to access the services from outside of cluster?

> Use nginx+keepalived or other tools as load balance. 

![image](http://img.blog.csdn.net/20160411232732855 "order prossing system")

### How to start Order Processing Service?
> install and start redis database 

> go build

> ./order_process

        Order Processing Service Start!
        time="2016-04-10T10:42:36+08:00" level=debug msg="Redis configuration loaded: {127.0.0.1 6379}" 
        time="2016-04-10T10:42:36+08:00" level=debug msg="Log configuration loaded: {debug}" 
        time="2016-04-10T10:42:36+08:00" level=debug msg="Service configuration loaded: {127.0.0.1 8080}" 
        time="2016-04-10T10:42:36+08:00" level=info msg="Initializing Raft Server" 
        time="2016-04-10T10:42:36+08:00" level=info msg="[bc8df584-c5c8-4e5a-6146-261835d06ded] stateChange initialized -> follower\n" 
        time="2016-04-10T10:42:36+08:00" level=info msg="Recovered from log" 
        time="2016-04-10T10:42:36+08:00" level=info msg="Initializing HTTP server" 
        time="2016-04-10T10:42:36+08:00" level=info msg="Listening at: 127.0.0.1:8080" 
        time="2016-04-10T10:42:36+08:00" level=info msg="[bc8df584-c5c8-4e5a-6146-261835d06ded] stateChange follower -> candidate\n" 
        time="2016-04-10T10:42:36+08:00" level=info msg="[bc8df584-c5c8-4e5a-6146-261835d06ded] stateChange candidate -> leader\n" 
        time="2016-04-10T10:42:36+08:00" level=info msg="[bc8df584-c5c8-4e5a-6146-261835d06ded] leaderChange  -> bc8df584-c5c8-4e5a-6146-261835d06ded" 
        time="2016-04-10T10:42:36+08:00" level=info msg="Start to perform leader tasks."  

### How to join the existed Order Processing Service Cluster?

> ./order_process --join localhost:8080

        Order Processing Service Start!
        time="2016-04-10T10:44:25+08:00" level=debug msg="Redis configuration loaded: {127.0.0.1 6379}"
        time="2016-04-10T10:44:25+08:00" level=debug msg="Log configuration loaded: {debug}"
        time="2016-04-10T10:44:25+08:00" level=debug msg="Service configuration loaded: {127.0.0.1 8082}"
        time="2016-04-10T10:44:25+08:00" level=info msg="Initializing Raft Server"
        time="2016-04-10T10:44:25+08:00" level=info msg="[bddda060-2c82-41e8-7b42-67ba6c39ba41] stateChange initialized -> follower\n"
        time="2016-04-10T10:44:25+08:00" level=info msg="Attempting to join leader: 127.0.0.1:8080"
        time="2016-04-10T10:44:25+08:00" level=info msg="Initializing HTTP server"
        time="2016-04-10T10:44:25+08:00" level=info msg="Listening at: 127.0.0.1:8082"
        time="2016-04-10T10:44:25+08:00" level=info msg="[bddda060-2c82-41e8-7b42-67ba6c39ba41] termChange 0 -> 1\n"
        time="2016-04-10T10:44:25+08:00" level=info msg="[bddda060-2c82-41e8-7b42-67ba6c39ba41] leaderChange  -> bc8df584-c5c8-4e5a-6146-261835d06ded"

### How to submit new order?

> curl -X POST --data "{}" -H "Authorization:user" http://localhost:8080/orders

        {"order_id":"8cc227c0-8dac-42cf-783e-f7bcb95bf455","start_time":"2016-03-27 10:22:31.4492618 +0000 UTC"}


### How to qurey the order state?

> curl -H "Authorization:user" http://localhost:8080/orders/8cc227c0-8dac-42cf-783e-f7bcb95bf455

        {
            "complete_time": "2016-03-27T10:22:51.4504397Z",
            "curent_step": "Completed",
            "order_id": "8cc227c0-8dac-42cf-783e-f7bcb95bf455",
            "start_time": "2016-03-27T10:22:31.4492618Z",
            "steps": [
                {
                    "step_complete_time": "2016-03-27T10:22:36.4500059Z",
                    "step_name": "Scheduling",
                    "step_start_time": "2016-03-27T10:22:31.4492618Z"
                },
                {
                    "step_complete_time": "2016-03-27T10:22:41.4501038Z",
                    "step_name": "Pre-Processing",
                    "step_start_time": "2016-03-27T10:22:36.4500059Z"
                },
                {
                    "step_complete_time": "2016-03-27T10:22:46.4504326Z",
                    "step_name": "Processing",
                    "step_start_time": "2016-03-27T10:22:41.4501038Z"
                },
                {
                    "step_complete_time": "2016-03-27T10:22:51.4504397Z",
                    "step_name": "Post-Processing",
                    "step_start_time": "2016-03-27T10:22:46.4504326Z"
                },
                {
                    "step_complete_time": "2016-03-27T10:22:51.4504397Z",
                    "step_name": "Completed",
                    "step_start_time": "2016-03-27T10:22:51.4504397Z"
                }
            ]

### What kind of action the System will take when error occurs during processing?
> The order should be marked as fail and rollback the steps.
> For example, the following order failed during "Post-Processing" step, all the steps performed before would be revoked.

        time="2016-03-28T23:53:25+08:00" level=debug msg="New order created with ID: [e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]" 
        time="2016-03-28T23:53:25+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]DispatchStepTask,current step: [Scheduling], next step:[Scheduling]" 
        time="2016-03-28T23:53:25+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]handling step[Scheduling]" 
        time="2016-03-28T23:53:25+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]Start step[Scheduling]" 
        time="2016-03-28T23:53:30+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]Finish step[Scheduling]" 
        time="2016-03-28T23:53:30+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]DispatchStepTask,current step: [Scheduling], next step:[Pre-Processing]" 
        time="2016-03-28T23:53:30+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]handling step[Pre-Processing]" 
        time="2016-03-28T23:53:30+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]Start step[Pre-Processing]" 
        time="2016-03-28T23:53:35+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]Finish step[Pre-Processing]" 
        time="2016-03-28T23:53:35+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]DispatchStepTask,current step: [Pre-Processing], next step:[Processing]" 
        time="2016-03-28T23:53:35+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]handling step[Processing]" 
        time="2016-03-28T23:53:35+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]Start step[Processing]" 
        time="2016-03-28T23:53:40+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]Finish step[Processing]" 
        time="2016-03-28T23:53:40+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]DispatchStepTask,current step: [Processing], next step:[Post-Processing]" 
        time="2016-03-28T23:53:40+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]handling step[Post-Processing]" 
        time="2016-03-28T23:53:40+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]Start step[Post-Processing]" 
        time="2016-03-28T23:53:45+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]Failure occurs when handling step[Post-Processing]" 
        time="2016-03-28T23:53:45+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]DispatchStepTask,current step: [Post-Processing], next step:[Failed]" 
        time="2016-03-28T23:53:45+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]handling step[Failed]" 
        time="2016-03-28T23:53:45+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]Start step[Failed]" 
        time="2016-03-28T23:53:45+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]Finish step[Failed]" 
        time="2016-03-28T23:53:45+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]DispatchStepTask,current step: [Failed], next step:[Post-Processing]" 
        time="2016-03-28T23:53:45+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]handling step[Post-Processing]" 
        time="2016-03-28T23:53:45+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]Rollback step[Post-Processing]" 
        time="2016-03-28T23:53:45+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]DispatchStepTask,current step: [Failed], next step:[Processing]" 
        time="2016-03-28T23:53:45+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]handling step[Processing]" 
        time="2016-03-28T23:53:45+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]Rollback step[Processing]" 
        time="2016-03-28T23:53:45+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]DispatchStepTask,current step: [Failed], next step:[Pre-Processing]" 
        time="2016-03-28T23:53:45+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]handling step[Pre-Processing]" 
        time="2016-03-28T23:53:45+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]Rollback step[Pre-Processing]" 
        time="2016-03-28T23:53:45+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]DispatchStepTask,current step: [Failed], next step:[Scheduling]" 
        time="2016-03-28T23:53:45+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]handling step[Scheduling]" 
        time="2016-03-28T23:53:45+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]Rollback step[Scheduling]" 
        time="2016-03-28T23:53:45+08:00" level=debug msg="[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]Finish Order" 
        
> Qurey the state of a failed job[e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1]

        {
            "complete_time": "2016-03-28 15:53:45.7443195 +0000 UTC",
            "current_step": "Failed",
            "order_id": "e6cbfc9e-91f8-4fd1-4d3b-1ca848f860c1",
            "start_time": "2016-03-28 15:53:25.6967625 +0000 UTC",
            "steps": [
                {
                    "step_complete_time": "2016-03-28 15:53:30.7328037 +0000 UTC",
                    "step_name": "Scheduling",
                    "step_start_time": "2016-03-28 15:53:25.6967625 +0000 UTC"
                },
                {
                    "step_complete_time": "2016-03-28 15:53:35.7344273 +0000 UTC",
                    "step_name": "Pre-Processing",
                    "step_start_time": "2016-03-28 15:53:30.7338099 +0000 UTC"
                },
                {
                    "step_complete_time": "2016-03-28 15:53:40.7359376 +0000 UTC",
                    "step_name": "Processing",
                    "step_start_time": "2016-03-28 15:53:35.7354256 +0000 UTC"
                },
                {
                    "step_name": "Post-Processing",
                    "step_start_time": "2016-03-28 15:53:40.7374388 +0000 UTC"
                },
                {
                    "step_complete_time": "2016-03-28 15:53:45.7398167 +0000 UTC",
                    "step_name": "Failed",
                    "step_start_time": "2016-03-28 15:53:45.7388152 +0000 UTC"
                }
            ]
        }

### What kind of action the System will take when one service of the cluster down?

> The leader of the cluster will select one service to take over the orders from service which is down. If the leader is down, new leader will be elected and it will transfer orders to peer

> The following request will be performed automatically:

> curl -X POST --data "{\"service_id\":\"630c4a80-11bc-447f-7a88-300d860132ae\"}" -H "Authorization:user" http://localhost:8080/service/transfer

        {"current_service_id":"9fb58d56-7e7c-4810-6610-5995f5075519","tranferred_service_id":"630c4a80-11bc-447f-7a88-300d860132ae"}

### How to qurey the status of Order Processing Service?

> curl http://localhost:8080/diagnostic/heartbeat

        {
            "generated_at": "2016-04-10 10:37:46.6735819 +0800 CST",
            "service_id": "bc8df584-c5c8-4e5a-6146-261835d06ded",
            "service_name": "order_process",
            "status": "OK",
            "version": "0.1"
        }
		
### How to qurey the status of the Cluster?

> curl http://localhost:8080/diagnostic/cluster

        {
            "generated_at": "2016-04-10 10:48:49.4810078 +0800 CST",
            "leader_name": "bc8df584-c5c8-4e5a-6146-261835d06ded",
            "nodes": [
                {
                    "connected": true,
                    "connection_string": "http://127.0.0.1:8082",
                    "last_activity": "2016-04-10T10:48:49.4349774+08:00",
                    "name": "bddda060-2c82-41e8-7b42-67ba6c39ba41"
                },
                {
                    "connected": true,
                    "connection_string": "http://127.0.0.1:8080",
                    "last_activity": "2016-04-10T10:48:49.4810078+08:00",
                    "name": "bc8df584-c5c8-4e5a-6146-261835d06ded"
                }
            ],
            "nodes_count": 2
        }
		
