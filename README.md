Order Processing System
------

### How to start Order Processing Service?
> install and start redis database 

> go build

> ./order_process

        Order Processing Service Start!
        time="2016-04-07T23:44:12+08:00" level=debug msg="Redis configuration loaded: {127.0.0.1 6379}" 
        time="2016-04-07T23:44:12+08:00" level=debug msg="Log configuration loaded: {debug}" 
        time="2016-04-07T23:44:12+08:00" level=debug msg="Service configuration loaded: {192.168.1.101 8080}" 
        time="2016-04-07T23:44:12+08:00" level=info msg="Initializing Raft Server" 
        time="2016-04-07T23:44:12+08:00" level=info msg="[9fe69c0b-d9a1-4dd9-6186-eec96f3b9b45] stateChange initialized -> follower\n" 
        time="2016-04-07T23:44:12+08:00" level=info msg="Initializing new cluster" 
        time="2016-04-07T23:44:12+08:00" level=info msg="[9fe69c0b-d9a1-4dd9-6186-eec96f3b9b45] stateChange follower -> leader\n" 
        time="2016-04-07T23:44:12+08:00" level=info msg="[9fe69c0b-d9a1-4dd9-6186-eec96f3b9b45] leaderChange  -> 9fe69c0b-d9a1-4dd9-6186-eec96f3b9b45" 
        time="2016-04-07T23:44:12+08:00" level=info msg="Start to perform leader tasks." 
        time="2016-04-07T23:44:12+08:00" level=info msg=9fe69c0b-d9a1-4dd9-6186-eec96f3b9b45 
        time="2016-04-07T23:44:12+08:00" level=info msg="map[]" 
        time="2016-04-07T23:44:12+08:00" level=info msg="Initializing HTTP server" 
        time="2016-04-07T23:44:12+08:00" level=info msg="Listening at: 192.168.1.101:8080" 

### How to join the existed Order Processing Service Cluster?

> ./order_process -join 192.168.1.101:8080

        Order Processing Service Start!
        time="2016-04-07T23:45:40+08:00" level=debug msg="Redis configuration loaded: {127.0.0.1 6379}" 
        time="2016-04-07T23:45:40+08:00" level=debug msg="Log configuration loaded: {debug}" 
        time="2016-04-07T23:45:40+08:00" level=debug msg="Service configuration loaded: {192.168.1.101 8082}" 
        time="2016-04-07T23:45:40+08:00" level=info msg="Initializing Raft Server" 
        time="2016-04-07T23:45:40+08:00" level=info msg="[15c7d264-1b30-4f54-77b3-76cab522f64b] stateChange initialized -> follower\n" 
        time="2016-04-07T23:45:40+08:00" level=info msg="Attempting to join leader: 192.168.1.101:8080" 
        time="2016-04-07T23:45:40+08:00" level=info 
        time="2016-04-07T23:45:40+08:00" level=info msg="map[]" 
        time="2016-04-07T23:45:40+08:00" level=info msg="Initializing HTTP server" 
        time="2016-04-07T23:45:40+08:00" level=info msg="Listening at: 192.168.1.101:8082" 

### How to qurey the status of Order Processing Service?

> curl -H "Authorization:user" http://localhost:8080/diagnostic/heartbeat

        {"serive_id":"f3bf183f-76b6-45eb-74d5-a970adfcfa99","status":"OK"}


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

### What will happen if error occurs during processing?
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

### How to take over the orders from service which is down?

> curl -X POST --data "{\"service_id\":\"630c4a80-11bc-447f-7a88-300d860132ae\"}" -H "Authorization:user" http://localhost:8080/service/transfer

        {"current_service_id":"9fb58d56-7e7c-4810-6610-5995f5075519","tranferred_service_id":"630c4a80-11bc-447f-7a88-300d860132ae"}
