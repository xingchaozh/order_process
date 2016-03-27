# order_process
Order processing system

1. How to Qurey the status of Order Processing Service?
C:\workspace\tools\curl\bin>curl -H "Authorization:user" http://localhost:8080/diagnostic/heartbeat

{"serive_id":"f3bf183f-76b6-45eb-74d5-a970adfcfa99","status":"OK"}

2. Success Order process

2.A Post a order
C:\workspace\tools\curl\bin>curl -X POST --data "{}" -H "Authorization:user" http://127.0.0.1:8080/orders

2.B The response of Order Processing System
{"order_id":"8cc227c0-8dac-42cf-783e-f7bcb95bf455","start_time":"2016-03-27 10:22:31.4492618 +0000 UTC"}

2.C The log of Order Processing System
time="2016-03-27T18:22:31+08:00" level=debug msg="POST Request Body: [map[]]" 
time="2016-03-27T18:22:31+08:00" level=debug msg="New order created with ID: [8cc227c0-8dac-42cf-783e-f7bcb95bf455]" 
time="2016-03-27T18:22:31+08:00" level=debug msg="[8cc227c0-8dac-42cf-783e-f7bcb95bf455]DispatchStepTask,current step: [Scheduling], next step:[Scheduling]" 
time="2016-03-27T18:22:31+08:00" level=debug msg="[8cc227c0-8dac-42cf-783e-f7bcb95bf455]handling step[Scheduling]" 
time="2016-03-27T18:22:31+08:00" level=debug msg="[8cc227c0-8dac-42cf-783e-f7bcb95bf455]Start step[Scheduling]" 
time="2016-03-27T18:22:36+08:00" level=debug msg="[8cc227c0-8dac-42cf-783e-f7bcb95bf455]Finish step[Scheduling]" 
time="2016-03-27T18:22:36+08:00" level=debug msg="[8cc227c0-8dac-42cf-783e-f7bcb95bf455]DispatchStepTask,current step: [Scheduling], next step:[Pre-Processing]" 
time="2016-03-27T18:22:36+08:00" level=debug msg="[8cc227c0-8dac-42cf-783e-f7bcb95bf455]handling step[Pre-Processing]" 
time="2016-03-27T18:22:36+08:00" level=debug msg="[8cc227c0-8dac-42cf-783e-f7bcb95bf455]Start step[Pre-Processing]" 
time="2016-03-27T18:22:41+08:00" level=debug msg="[8cc227c0-8dac-42cf-783e-f7bcb95bf455]Finish step[Pre-Processing]" 
time="2016-03-27T18:22:41+08:00" level=debug msg="[8cc227c0-8dac-42cf-783e-f7bcb95bf455]DispatchStepTask,current step: [Pre-Processing], next step:[Processing]" 
time="2016-03-27T18:22:41+08:00" level=debug msg="[8cc227c0-8dac-42cf-783e-f7bcb95bf455]handling step[Processing]" 
time="2016-03-27T18:22:41+08:00" level=debug msg="[8cc227c0-8dac-42cf-783e-f7bcb95bf455]Start step[Processing]" 
time="2016-03-27T18:22:46+08:00" level=debug msg="[8cc227c0-8dac-42cf-783e-f7bcb95bf455]Finish step[Processing]" 
time="2016-03-27T18:22:46+08:00" level=debug msg="[8cc227c0-8dac-42cf-783e-f7bcb95bf455]DispatchStepTask,current step: [Processing], next step:[Post-Processing]" 
time="2016-03-27T18:22:46+08:00" level=debug msg="[8cc227c0-8dac-42cf-783e-f7bcb95bf455]handling step[Post-Processing]" 
time="2016-03-27T18:22:46+08:00" level=debug msg="[8cc227c0-8dac-42cf-783e-f7bcb95bf455]Start step[Post-Processing]" 
time="2016-03-27T18:22:51+08:00" level=debug msg="[8cc227c0-8dac-42cf-783e-f7bcb95bf455]Finish step[Post-Processing]" 
time="2016-03-27T18:22:51+08:00" level=debug msg="[8cc227c0-8dac-42cf-783e-f7bcb95bf455]DispatchStepTask,current step: [Post-Processing], next step:[Completed]" 
time="2016-03-27T18:22:51+08:00" level=debug msg="[8cc227c0-8dac-42cf-783e-f7bcb95bf455]handling step[Completed]" 
time="2016-03-27T18:22:51+08:00" level=debug msg="[8cc227c0-8dac-42cf-783e-f7bcb95bf455]Start step[Completed]" 
time="2016-03-27T18:22:51+08:00" level=debug msg="[8cc227c0-8dac-42cf-783e-f7bcb95bf455]Finish step[Completed]" 
time="2016-03-27T18:22:51+08:00" level=debug msg="[8cc227c0-8dac-42cf-783e-f7bcb95bf455]Finish Order" 
time="2016-03-27T18:22:51+08:00" level=debug 

2.D Qurey the order state
C:\workspace\tools\curl\bin>curl -H "Authorization:user" http://localhost:8080/orders/8cc227c0-8dac-42cf-783e-f7bcb95bf455
{
    "complete_time": "2016-03-27T10:22:51.4504397Z",
    "curent_step": "Completed",
    "finished": true,
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
}


3. Order process with failure

3.A Post a order
C:\workspace\tools\curl\bin>curl -X POST --data "{}" -H "Authorization:user" http://127.0.0.1:8080/orders

3.B The response of Order Processing System
{"order_id":"670af647-7871-4aa1-6691-7ff39024b100","start_time":"2016-03-27 10:30:00.0257925 +0000 UTC"}

3.C The log of Order Processing System with Rollback
time="2016-03-27T18:30:00+08:00" level=debug msg="POST Request Body: [map[]]" 
time="2016-03-27T18:30:00+08:00" level=debug msg="New order created with ID: [670af647-7871-4aa1-6691-7ff39024b100]" 
time="2016-03-27T18:30:00+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]DispatchStepTask,current step: [Scheduling], next step:[Scheduling]" 
time="2016-03-27T18:30:00+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]handling step[Scheduling]" 
time="2016-03-27T18:30:00+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]Start step[Scheduling]" 
time="2016-03-27T18:30:05+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]Finish step[Scheduling]" 
time="2016-03-27T18:30:05+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]DispatchStepTask,current step: [Scheduling], next step:[Pre-Processing]" 
time="2016-03-27T18:30:05+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]handling step[Pre-Processing]" 
time="2016-03-27T18:30:05+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]Start step[Pre-Processing]" 
time="2016-03-27T18:30:10+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]Finish step[Pre-Processing]" 
time="2016-03-27T18:30:10+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]DispatchStepTask,current step: [Pre-Processing], next step:[Processing]" 
time="2016-03-27T18:30:10+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]handling step[Processing]" 
time="2016-03-27T18:30:10+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]Start step[Processing]" 
time="2016-03-27T18:30:15+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]Failure occurs when handling step[Processing]" 
time="2016-03-27T18:30:15+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]DispatchStepTask,current step: [Processing], next step:[Failed]" 
time="2016-03-27T18:30:15+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]handling step[Failed]" 
time="2016-03-27T18:30:15+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]Start step[Failed]" 
time="2016-03-27T18:30:15+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]Finish step[Failed]" 
time="2016-03-27T18:30:15+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]DispatchStepTask,current step: [Failed], next step:[Processing]" 
time="2016-03-27T18:30:15+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]handling step[Processing]" 
time="2016-03-27T18:30:15+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]Rollback step[Processing]" 
time="2016-03-27T18:30:15+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]DispatchStepTask,current step: [Failed], next step:[Pre-Processing]" 
time="2016-03-27T18:30:15+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]handling step[Pre-Processing]" 
time="2016-03-27T18:30:15+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]Rollback step[Pre-Processing]" 
time="2016-03-27T18:30:15+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]DispatchStepTask,current step: [Failed], next step:[Scheduling]" 
time="2016-03-27T18:30:15+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]handling step[Scheduling]" 
time="2016-03-27T18:30:15+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]Rollback step[Scheduling]" 
time="2016-03-27T18:30:15+08:00" level=debug msg="[670af647-7871-4aa1-6691-7ff39024b100]Finish Order" 
time="2016-03-27T18:30:15+08:00" level=debug 

3.D Qurey the order state
C:\workspace\tools\curl\bin>curl -H "Authorization:user" http://localhost:8080/orders/670af647-7871-4aa1-6691-7ff39024b100
time="2016-03-27T18:30:40+08:00" level=debug msg="Get Request ID: [670af647-7871-4aa1-6691-7ff39024b100]" 
{
    "complete_time": "2016-03-27T10:30:15.0261067Z",
    "curent_step": "Failed",
    "finished": true,
    "order_id": "670af647-7871-4aa1-6691-7ff39024b100",
    "start_time": "2016-03-27T10:30:00.0257925Z",
    "steps": [
        {
            "step_complete_time": "2016-03-27T10:30:05.0257992Z",
            "step_name": "Scheduling",
            "step_start_time": "2016-03-27T10:30:00.0257925Z"
        },
        {
            "step_complete_time": "2016-03-27T10:30:10.0258102Z",
            "step_name": "Pre-Processing",
            "step_start_time": "2016-03-27T10:30:05.0257992Z"
        },
        {
            "step_name": "Processing",
            "step_start_time": "2016-03-27T10:30:10.0258102Z"
        },
        {
            "step_complete_time": "2016-03-27T10:30:15.0261067Z",
            "step_name": "Failed",
            "step_start_time": "2016-03-27T10:30:15.0261067Z"
        }
    ]
}
